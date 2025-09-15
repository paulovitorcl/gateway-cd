package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//+kubebuilder:object:generate=true

// CanaryDeploymentPhase represents the current phase of a canary deployment
type CanaryDeploymentPhase string

const (
	CanaryDeploymentPhasePending    CanaryDeploymentPhase = "Pending"
	CanaryDeploymentPhaseProgressing CanaryDeploymentPhase = "Progressing"
	CanaryDeploymentPhasePaused     CanaryDeploymentPhase = "Paused"
	CanaryDeploymentPhaseSucceeded  CanaryDeploymentPhase = "Succeeded"
	CanaryDeploymentPhaseFailed     CanaryDeploymentPhase = "Failed"
	CanaryDeploymentPhaseRollingBack CanaryDeploymentPhase = "RollingBack"
)

// TrafficSplitStep defines a traffic split configuration
type TrafficSplitStep struct {
	// Weight is the percentage of traffic to route to canary version (0-100)
	Weight int32 `json:"weight"`
	// Duration is how long to maintain this weight before moving to next step
	Duration string `json:"duration,omitempty"`
	// Pause indicates whether to pause at this step for manual approval
	Pause bool `json:"pause,omitempty"`
}

// AnalysisTemplate defines success criteria for canary analysis
type AnalysisTemplate struct {
	// Metrics to evaluate during canary analysis
	Metrics []AnalysisMetric `json:"metrics,omitempty"`
	// SuccessRate is the minimum success rate threshold (0.0-1.0)
	SuccessRate float64 `json:"successRate,omitempty"`
	// MaxLatency is the maximum acceptable latency in milliseconds
	MaxLatency int32 `json:"maxLatency,omitempty"`
	// AnalysisInterval is how often to run analysis
	AnalysisInterval string `json:"analysisInterval,omitempty"`
}

// AnalysisMetric defines a metric to monitor during canary analysis
type AnalysisMetric struct {
	// Name of the metric
	Name string `json:"name"`
	// Query is the Prometheus query to execute
	Query string `json:"query"`
	// Threshold is the threshold value for this metric
	Threshold float64 `json:"threshold"`
	// Operator is the comparison operator (>, <, >=, <=, ==, !=)
	Operator string `json:"operator"`
}

// CanaryDeploymentSpec defines the desired state of CanaryDeployment
type CanaryDeploymentSpec struct {
	// TargetRef references the target workload for canary deployment
	TargetRef WorkloadRef `json:"targetRef"`

	// Service is the Kubernetes service associated with the workload
	Service ServiceRef `json:"service"`

	// Gateway configuration for traffic management
	Gateway GatewayRef `json:"gateway"`

	// TrafficSplit defines the traffic splitting strategy
	TrafficSplit []TrafficSplitStep `json:"trafficSplit"`

	// Analysis defines success criteria and rollback conditions
	Analysis AnalysisTemplate `json:"analysis,omitempty"`

	// AutoPromote automatically promotes canary to stable if analysis succeeds
	AutoPromote bool `json:"autoPromote,omitempty"`

	// SkipAnalysis skips canary analysis (useful for testing)
	SkipAnalysis bool `json:"skipAnalysis,omitempty"`
}

// WorkloadRef references a Kubernetes workload
type WorkloadRef struct {
	// APIVersion of the target workload
	APIVersion string `json:"apiVersion"`
	// Kind of the target workload (Deployment, ReplicaSet, etc.)
	Kind string `json:"kind"`
	// Name of the target workload
	Name string `json:"name"`
}

// ServiceRef references a Kubernetes service
type ServiceRef struct {
	// Name of the service
	Name string `json:"name"`
	// Port is the service port to use for canary traffic
	Port int32 `json:"port"`
}

// GatewayRef references Gateway API resources
type GatewayRef struct {
	// HTTPRoute is the name of the HTTPRoute to manage
	HTTPRoute string `json:"httpRoute"`
	// Gateway is the name of the Gateway (optional)
	Gateway string `json:"gateway,omitempty"`
	// Namespace is the namespace of the Gateway API resources
	Namespace string `json:"namespace,omitempty"`
}

// CanaryDeploymentStatus defines the observed state of CanaryDeployment
type CanaryDeploymentStatus struct {
	// Phase is the current phase of the canary deployment
	Phase CanaryDeploymentPhase `json:"phase,omitempty"`

	// Message provides human-readable details about the current state
	Message string `json:"message,omitempty"`

	// CurrentStep is the index of the current traffic split step
	CurrentStep int32 `json:"currentStep,omitempty"`

	// CanaryWeight is the current percentage of traffic routed to canary
	CanaryWeight int32 `json:"canaryWeight,omitempty"`

	// StableWeight is the current percentage of traffic routed to stable
	StableWeight int32 `json:"stableWeight,omitempty"`

	// Conditions represent the latest available observations
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// LastTransitionTime is when the current phase was entered
	LastTransitionTime *metav1.Time `json:"lastTransitionTime,omitempty"`

	// Analysis results from the current or last analysis run
	AnalysisRun *AnalysisRunStatus `json:"analysisRun,omitempty"`
}

// AnalysisRunStatus contains the results of a canary analysis run
type AnalysisRunStatus struct {
	// Phase of the analysis run
	Phase string `json:"phase,omitempty"`
	// SuccessRate observed during analysis
	SuccessRate float64 `json:"successRate,omitempty"`
	// AverageLatency observed during analysis
	AverageLatency int32 `json:"averageLatency,omitempty"`
	// MetricResults contains results for each configured metric
	MetricResults []MetricResult `json:"metricResults,omitempty"`
	// StartedAt is when the analysis run started
	StartedAt *metav1.Time `json:"startedAt,omitempty"`
	// CompletedAt is when the analysis run completed
	CompletedAt *metav1.Time `json:"completedAt,omitempty"`
}

// MetricResult contains the result of evaluating a specific metric
type MetricResult struct {
	// Name of the metric
	Name string `json:"name"`
	// Value is the measured value
	Value float64 `json:"value"`
	// Threshold is the configured threshold
	Threshold float64 `json:"threshold"`
	// Passed indicates whether the metric passed the threshold check
	Passed bool `json:"passed"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"
//+kubebuilder:printcolumn:name="Canary Weight",type="integer",JSONPath=".status.canaryWeight"
//+kubebuilder:printcolumn:name="Step",type="integer",JSONPath=".status.currentStep"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// CanaryDeployment is the Schema for the canarydeployments API
type CanaryDeployment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CanaryDeploymentSpec   `json:"spec,omitempty"`
	Status CanaryDeploymentStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CanaryDeploymentList contains a list of CanaryDeployment
type CanaryDeploymentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CanaryDeployment `json:"items"`
}