package controller

import (
	"context"
	"fmt"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	gatewaycdv1alpha1 "gateway-cd/pkg/api/v1alpha1"
	"gateway-cd/pkg/gateway"
	"gateway-cd/pkg/metrics"
)

// CanaryDeploymentReconciler reconciles a CanaryDeployment object
type CanaryDeploymentReconciler struct {
	client.Client
	Scheme          *runtime.Scheme
	GatewayManager  *gateway.Manager
	MetricsProvider metrics.Provider
}

//+kubebuilder:rbac:groups=gateway-cd.io,resources=canarydeployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=gateway-cd.io,resources=canarydeployments/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=gateway-cd.io,resources=canarydeployments/finalizers,verbs=update
//+kubebuilder:rbac:groups=gateway.networking.k8s.io,resources=httproutes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop
func (r *CanaryDeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Fetch the CanaryDeployment instance
	var canary gatewaycdv1alpha1.CanaryDeployment
	if err := r.Get(ctx, req.NamespacedName, &canary); err != nil {
		if apierrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			return ctrl.Result{}, nil
		}
		log.Error(err, "unable to fetch CanaryDeployment")
		return ctrl.Result{}, err
	}

	// Handle deletion
	if canary.DeletionTimestamp != nil {
		return r.handleDeletion(ctx, &canary)
	}

	// Initialize status if needed
	if canary.Status.Phase == "" {
		canary.Status.Phase = gatewaycdv1alpha1.CanaryDeploymentPhasePending
		canary.Status.CurrentStep = 0
		canary.Status.CanaryWeight = 0
		canary.Status.StableWeight = 100
		canary.Status.LastTransitionTime = &metav1.Time{Time: time.Now()}
		if err := r.Status().Update(ctx, &canary); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: time.Second * 5}, nil
	}

	// Main reconciliation logic based on phase
	switch canary.Status.Phase {
	case gatewaycdv1alpha1.CanaryDeploymentPhasePending:
		return r.handlePending(ctx, &canary)
	case gatewaycdv1alpha1.CanaryDeploymentPhaseProgressing:
		return r.handleProgressing(ctx, &canary)
	case gatewaycdv1alpha1.CanaryDeploymentPhasePaused:
		return r.handlePaused(ctx, &canary)
	case gatewaycdv1alpha1.CanaryDeploymentPhaseRollingBack:
		return r.handleRollingBack(ctx, &canary)
	case gatewaycdv1alpha1.CanaryDeploymentPhaseSucceeded,
		 gatewaycdv1alpha1.CanaryDeploymentPhaseFailed:
		// Terminal phases - no action needed
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

func (r *CanaryDeploymentReconciler) handlePending(ctx context.Context, canary *gatewaycdv1alpha1.CanaryDeployment) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Validate the canary deployment configuration
	if err := r.validateCanaryDeployment(ctx, canary); err != nil {
		canary.Status.Phase = gatewaycdv1alpha1.CanaryDeploymentPhaseFailed
		canary.Status.Message = fmt.Sprintf("Validation failed: %v", err)
		r.Status().Update(ctx, canary)
		return ctrl.Result{}, err
	}

	// Start the canary deployment
	log.Info("Starting canary deployment", "canary", canary.Name)
	canary.Status.Phase = gatewaycdv1alpha1.CanaryDeploymentPhaseProgressing
	canary.Status.Message = "Starting canary deployment"
	canary.Status.LastTransitionTime = &metav1.Time{Time: time.Now()}

	if err := r.Status().Update(ctx, canary); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: time.Second * 5}, nil
}

func (r *CanaryDeploymentReconciler) handleProgressing(ctx context.Context, canary *gatewaycdv1alpha1.CanaryDeployment) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Check if we have more steps to process
	if int(canary.Status.CurrentStep) >= len(canary.Spec.TrafficSplit) {
		// All steps completed successfully
		canary.Status.Phase = gatewaycdv1alpha1.CanaryDeploymentPhaseSucceeded
		canary.Status.Message = "Canary deployment completed successfully"
		canary.Status.CanaryWeight = 100
		canary.Status.StableWeight = 0
		canary.Status.LastTransitionTime = &metav1.Time{Time: time.Now()}
		r.Status().Update(ctx, canary)
		return ctrl.Result{}, nil
	}

	currentStep := canary.Spec.TrafficSplit[canary.Status.CurrentStep]

	// Update traffic split
	if err := r.GatewayManager.UpdateTrafficSplit(ctx, canary, int(currentStep.Weight)); err != nil {
		log.Error(err, "Failed to update traffic split")
		canary.Status.Message = fmt.Sprintf("Failed to update traffic split: %v", err)
		r.Status().Update(ctx, canary)
		return ctrl.Result{RequeueAfter: time.Second * 30}, nil
	}

	// Update status
	canary.Status.CanaryWeight = currentStep.Weight
	canary.Status.StableWeight = 100 - currentStep.Weight
	canary.Status.Message = fmt.Sprintf("Traffic split updated: %d%% canary, %d%% stable",
		currentStep.Weight, 100-currentStep.Weight)

	// Check if step requires pause
	if currentStep.Pause {
		canary.Status.Phase = gatewaycdv1alpha1.CanaryDeploymentPhasePaused
		canary.Status.Message = fmt.Sprintf("Paused at step %d for manual approval", canary.Status.CurrentStep+1)
		canary.Status.LastTransitionTime = &metav1.Time{Time: time.Now()}
		r.Status().Update(ctx, canary)
		return ctrl.Result{}, nil
	}

	// Run analysis if configured
	if !canary.Spec.SkipAnalysis && canary.Spec.Analysis.SuccessRate > 0 {
		passed, err := r.runAnalysis(ctx, canary)
		if err != nil {
			log.Error(err, "Analysis failed")
			canary.Status.Message = fmt.Sprintf("Analysis failed: %v", err)
			r.Status().Update(ctx, canary)
			return ctrl.Result{RequeueAfter: time.Second * 30}, nil
		}

		if !passed {
			log.Info("Analysis failed, initiating rollback")
			canary.Status.Phase = gatewaycdv1alpha1.CanaryDeploymentPhaseRollingBack
			canary.Status.Message = "Analysis failed, rolling back"
			canary.Status.LastTransitionTime = &metav1.Time{Time: time.Now()}
			r.Status().Update(ctx, canary)
			return ctrl.Result{RequeueAfter: time.Second * 5}, nil
		}
	}

	// Move to next step
	canary.Status.CurrentStep++
	r.Status().Update(ctx, canary)

	// Calculate requeue time based on step duration
	requeueAfter := time.Second * 30 // default
	if currentStep.Duration != "" {
		if duration, err := time.ParseDuration(currentStep.Duration); err == nil {
			requeueAfter = duration
		}
	}

	return ctrl.Result{RequeueAfter: requeueAfter}, nil
}

func (r *CanaryDeploymentReconciler) handlePaused(ctx context.Context, canary *gatewaycdv1alpha1.CanaryDeployment) (ctrl.Result, error) {
	// Check for resume annotation or other resume conditions
	if canary.Annotations["gateway-cd.io/resume"] == "true" {
		delete(canary.Annotations, "gateway-cd.io/resume")
		canary.Status.Phase = gatewaycdv1alpha1.CanaryDeploymentPhaseProgressing
		canary.Status.CurrentStep++
		canary.Status.Message = "Resumed from pause"
		canary.Status.LastTransitionTime = &metav1.Time{Time: time.Now()}

		if err := r.Update(ctx, canary); err != nil {
			return ctrl.Result{}, err
		}
		if err := r.Status().Update(ctx, canary); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: time.Second * 5}, nil
	}

	// Check for abort annotation
	if canary.Annotations["gateway-cd.io/abort"] == "true" {
		canary.Status.Phase = gatewaycdv1alpha1.CanaryDeploymentPhaseRollingBack
		canary.Status.Message = "Aborted by user"
		canary.Status.LastTransitionTime = &metav1.Time{Time: time.Now()}
		r.Status().Update(ctx, canary)
		return ctrl.Result{RequeueAfter: time.Second * 5}, nil
	}

	// Stay paused
	return ctrl.Result{RequeueAfter: time.Second * 30}, nil
}

func (r *CanaryDeploymentReconciler) handleRollingBack(ctx context.Context, canary *gatewaycdv1alpha1.CanaryDeployment) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Reset traffic to 100% stable
	if err := r.GatewayManager.UpdateTrafficSplit(ctx, canary, 0); err != nil {
		log.Error(err, "Failed to rollback traffic split")
		return ctrl.Result{RequeueAfter: time.Second * 10}, nil
	}

	canary.Status.Phase = gatewaycdv1alpha1.CanaryDeploymentPhaseFailed
	canary.Status.CanaryWeight = 0
	canary.Status.StableWeight = 100
	canary.Status.Message = "Rollback completed"
	canary.Status.LastTransitionTime = &metav1.Time{Time: time.Now()}

	r.Status().Update(ctx, canary)
	return ctrl.Result{}, nil
}

func (r *CanaryDeploymentReconciler) handleDeletion(ctx context.Context, canary *gatewaycdv1alpha1.CanaryDeployment) (ctrl.Result, error) {
	// Cleanup Gateway API resources if needed
	if err := r.GatewayManager.Cleanup(ctx, canary); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *CanaryDeploymentReconciler) validateCanaryDeployment(ctx context.Context, canary *gatewaycdv1alpha1.CanaryDeployment) error {
	// Validate target workload exists
	// Validate service exists
	// Validate Gateway API resources exist
	// This is a simplified validation - implement full validation as needed
	return nil
}

func (r *CanaryDeploymentReconciler) runAnalysis(ctx context.Context, canary *gatewaycdv1alpha1.CanaryDeployment) (bool, error) {
	log := log.FromContext(ctx)

	if r.MetricsProvider == nil {
		log.Info("No metrics provider configured, skipping analysis")
		return true, nil
	}

	// Run analysis using the metrics provider
	result, err := r.MetricsProvider.RunAnalysis(ctx, canary)
	if err != nil {
		return false, err
	}

	// Update analysis run status
	canary.Status.AnalysisRun = &gatewaycdv1alpha1.AnalysisRunStatus{
		Phase:          result.Phase,
		SuccessRate:    result.SuccessRate,
		AverageLatency: result.AverageLatency,
		MetricResults:  result.MetricResults,
		StartedAt:      result.StartedAt,
		CompletedAt:    result.CompletedAt,
	}

	return result.Passed, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CanaryDeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&gatewaycdv1alpha1.CanaryDeployment{}).
		Complete(r)
}