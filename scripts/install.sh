#!/bin/bash

set -e

echo "Installing Gateway CD Platform..."

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo "Error: kubectl is required but not installed."
    exit 1
fi

# Check if cluster is accessible
if ! kubectl cluster-info &> /dev/null; then
    echo "Error: Cannot connect to Kubernetes cluster."
    exit 1
fi

# Install Gateway API CRDs if not present
echo "Checking Gateway API CRDs..."
if ! kubectl get crd gateways.gateway.networking.k8s.io &> /dev/null; then
    echo "Installing Gateway API CRDs..."
    kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.0.0/standard-install.yaml
    echo "Waiting for Gateway API CRDs to be ready..."
    kubectl wait --for condition=established --timeout=60s crd/gateways.gateway.networking.k8s.io
    kubectl wait --for condition=established --timeout=60s crd/httproutes.gateway.networking.k8s.io
else
    echo "Gateway API CRDs already installed."
fi

# Apply Gateway CD manifests
echo "Installing Gateway CD components..."

# Create namespace
kubectl apply -f deploy/k8s/namespace.yaml

# Install CRDs
kubectl apply -f deploy/k8s/crds/

# Wait for CRDs to be ready
echo "Waiting for Gateway CD CRDs to be ready..."
kubectl wait --for condition=established --timeout=60s crd/canarydeployments.gateway-cd.io

# Install RBAC
kubectl apply -f deploy/k8s/rbac.yaml

# Install controller
kubectl apply -f deploy/k8s/controller.yaml

# Install API server
kubectl apply -f deploy/k8s/api-server.yaml

# Install web dashboard
kubectl apply -f deploy/k8s/web-dashboard.yaml

# Install ingress (optional)
if kubectl get ingressclass nginx &> /dev/null; then
    echo "Installing ingress..."
    kubectl apply -f deploy/k8s/ingress.yaml
else
    echo "NGINX ingress controller not found, skipping ingress installation."
    echo "You can manually apply deploy/k8s/ingress.yaml after installing an ingress controller."
fi

echo "Waiting for deployments to be ready..."
kubectl wait --for=condition=available --timeout=300s deployment/gateway-cd-controller -n gateway-cd
kubectl wait --for=condition=available --timeout=300s deployment/gateway-cd-api -n gateway-cd
kubectl wait --for=condition=available --timeout=300s deployment/gateway-cd-web -n gateway-cd

echo ""
echo "Gateway CD Platform installed successfully!"
echo ""
echo "To access the dashboard:"
echo "1. Port forward: kubectl port-forward svc/gateway-cd-web 3000:3000 -n gateway-cd"
echo "2. Open http://localhost:3000"
echo ""
echo "To access the API:"
echo "1. Port forward: kubectl port-forward svc/gateway-cd-api 8080:8080 -n gateway-cd"
echo "2. API available at http://localhost:8080/api/v1"
echo ""
echo "Example usage:"
echo "kubectl apply -f examples/sample-canary.yaml"