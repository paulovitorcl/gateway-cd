package gateway

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1"

	gatewaycdv1alpha1 "gateway-cd/pkg/api/v1alpha1"
)

// Manager handles Gateway API operations for canary deployments
type Manager struct {
	client client.Client
}

// NewManager creates a new Gateway API manager
func NewManager(client client.Client) *Manager {
	return &Manager{
		client: client,
	}
}

// UpdateTrafficSplit updates the HTTPRoute to split traffic between stable and canary services
func (m *Manager) UpdateTrafficSplit(ctx context.Context, canary *gatewaycdv1alpha1.CanaryDeployment, canaryWeight int) error {
	// Get the HTTPRoute
	httpRoute := &gatewayapi.HTTPRoute{}
	httpRouteNamespace := canary.Spec.Gateway.Namespace
	if httpRouteNamespace == "" {
		httpRouteNamespace = canary.Namespace
	}

	err := m.client.Get(ctx, types.NamespacedName{
		Name:      canary.Spec.Gateway.HTTPRoute,
		Namespace: httpRouteNamespace,
	}, httpRoute)
	if err != nil {
		return fmt.Errorf("failed to get HTTPRoute %s/%s: %w", httpRouteNamespace, canary.Spec.Gateway.HTTPRoute, err)
	}

	// Update the HTTPRoute with new traffic split
	if err := m.updateHTTPRouteBackends(httpRoute, canary, canaryWeight); err != nil {
		return fmt.Errorf("failed to update HTTPRoute backends: %w", err)
	}

	// Update the HTTPRoute in the cluster
	if err := m.client.Update(ctx, httpRoute); err != nil {
		return fmt.Errorf("failed to update HTTPRoute: %w", err)
	}

	return nil
}

// updateHTTPRouteBackends modifies the HTTPRoute to include traffic splitting
func (m *Manager) updateHTTPRouteBackends(httpRoute *gatewayapi.HTTPRoute, canary *gatewaycdv1alpha1.CanaryDeployment, canaryWeight int) error {
	stableWeight := 100 - canaryWeight

	// Create backend references
	stableBackend := gatewayapi.HTTPBackendRef{
		BackendRef: gatewayapi.BackendRef{
			BackendObjectReference: gatewayapi.BackendObjectReference{
				Name: gatewayapi.ObjectName(canary.Spec.Service.Name),
				Port: (*gatewayapi.PortNumber)(&canary.Spec.Service.Port),
			},
			Weight: func(w int) *int32 { i := int32(w); return &i }(stableWeight),
		},
	}

	canaryBackend := gatewayapi.HTTPBackendRef{
		BackendRef: gatewayapi.BackendRef{
			BackendObjectReference: gatewayapi.BackendObjectReference{
				Name: gatewayapi.ObjectName(fmt.Sprintf("%s-canary", canary.Spec.Service.Name)),
				Port: (*gatewayapi.PortNumber)(&canary.Spec.Service.Port),
			},
			Weight: func(w int) *int32 { i := int32(w); return &i }(canaryWeight),
		},
	}

	// Update all rules with the new backend configuration
	for i := range httpRoute.Spec.Rules {
		// Find or create the default match (all traffic)
		if len(httpRoute.Spec.Rules[i].Matches) == 0 {
			httpRoute.Spec.Rules[i].Matches = []gatewayapi.HTTPRouteMatch{{}}
		}

		// Update backend references
		if canaryWeight == 0 {
			// Only stable backend
			httpRoute.Spec.Rules[i].BackendRefs = []gatewayapi.HTTPBackendRef{stableBackend}
		} else if canaryWeight == 100 {
			// Only canary backend (promotion complete)
			httpRoute.Spec.Rules[i].BackendRefs = []gatewayapi.HTTPBackendRef{canaryBackend}
		} else {
			// Both backends with weights
			httpRoute.Spec.Rules[i].BackendRefs = []gatewayapi.HTTPBackendRef{stableBackend, canaryBackend}
		}
	}

	return nil
}

// CreateCanaryService creates a canary service for the deployment
func (m *Manager) CreateCanaryService(ctx context.Context, canary *gatewaycdv1alpha1.CanaryDeployment) error {
	// This would create a canary service that points to the canary deployment
	// Implementation depends on your specific service creation strategy
	return nil
}

// Cleanup removes any Gateway API resources created for the canary deployment
func (m *Manager) Cleanup(ctx context.Context, canary *gatewaycdv1alpha1.CanaryDeployment) error {
	// Reset HTTPRoute to only point to stable service
	if err := m.UpdateTrafficSplit(ctx, canary, 0); err != nil {
		return fmt.Errorf("failed to cleanup traffic split: %w", err)
	}

	// Clean up any canary-specific services if needed
	return nil
}

// ValidateGatewayConfiguration validates that the required Gateway API resources exist
func (m *Manager) ValidateGatewayConfiguration(ctx context.Context, canary *gatewaycdv1alpha1.CanaryDeployment) error {
	// Check if HTTPRoute exists
	httpRoute := &gatewayapi.HTTPRoute{}
	httpRouteNamespace := canary.Spec.Gateway.Namespace
	if httpRouteNamespace == "" {
		httpRouteNamespace = canary.Namespace
	}

	err := m.client.Get(ctx, types.NamespacedName{
		Name:      canary.Spec.Gateway.HTTPRoute,
		Namespace: httpRouteNamespace,
	}, httpRoute)
	if err != nil {
		return fmt.Errorf("HTTPRoute %s/%s not found: %w", httpRouteNamespace, canary.Spec.Gateway.HTTPRoute, err)
	}

	// Check if Gateway exists (if specified)
	if canary.Spec.Gateway.Gateway != "" {
		gateway := &gatewayapi.Gateway{}
		gatewayNamespace := canary.Spec.Gateway.Namespace
		if gatewayNamespace == "" {
			gatewayNamespace = canary.Namespace
		}

		err := m.client.Get(ctx, types.NamespacedName{
			Name:      canary.Spec.Gateway.Gateway,
			Namespace: gatewayNamespace,
		}, gateway)
		if err != nil {
			return fmt.Errorf("Gateway %s/%s not found: %w", gatewayNamespace, canary.Spec.Gateway.Gateway, err)
		}
	}

	return nil
}