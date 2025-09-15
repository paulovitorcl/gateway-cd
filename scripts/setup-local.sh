#!/bin/bash

set -e

echo "🚀 Setting up local Gateway CD development environment..."

# Check prerequisites
echo "Checking prerequisites..."

if ! command -v docker &> /dev/null; then
    echo "❌ Error: Docker is required but not installed."
    echo "Please install Docker Desktop and try again."
    exit 1
fi

if ! command -v kind &> /dev/null; then
    echo "❌ Error: kind is required but not installed."
    echo "Installing kind..."
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        if command -v brew &> /dev/null; then
            brew install kind
        else
            echo "Please install Homebrew or install kind manually: https://kind.sigs.k8s.io/docs/user/quick-start/"
            exit 1
        fi
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        # Linux
        curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.20.0/kind-linux-amd64
        chmod +x ./kind
        sudo mv ./kind /usr/local/bin/kind
    else
        echo "Please install kind manually: https://kind.sigs.k8s.io/docs/user/quick-start/"
        exit 1
    fi
fi

if ! command -v kubectl &> /dev/null; then
    echo "❌ Error: kubectl is required but not installed."
    echo "Installing kubectl..."
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        if command -v brew &> /dev/null; then
            brew install kubectl
        else
            curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/darwin/amd64/kubectl"
            chmod +x kubectl
            sudo mv kubectl /usr/local/bin/
        fi
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        # Linux
        curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
        chmod +x kubectl
        sudo mv kubectl /usr/local/bin/
    fi
fi

echo "✅ Prerequisites check complete"

# Create kind cluster
echo "🏗️  Creating kind cluster..."

# Check if cluster already exists
if kind get clusters | grep -q "^gateway-cd$"; then
    echo "ℹ️  Cluster 'gateway-cd' already exists"
    read -p "Do you want to delete and recreate it? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo "🗑️  Deleting existing cluster..."
        kind delete cluster --name gateway-cd
    else
        echo "Using existing cluster..."
        kind export kubeconfig --name gateway-cd
        kubectl cluster-info --context kind-gateway-cd
    fi
fi

if ! kind get clusters | grep -q "^gateway-cd$"; then
    echo "Creating new kind cluster with custom configuration..."

    # Create kind config if it doesn't exist
    if [ ! -f "scripts/kind-config.yaml" ]; then
        cat > scripts/kind-config.yaml << EOF
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
name: gateway-cd
nodes:
- role: control-plane
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
  extraPortMappings:
  - containerPort: 80
    hostPort: 8080
    protocol: TCP
  - containerPort: 443
    hostPort: 8443
    protocol: TCP
- role: worker
- role: worker
EOF
    fi

    kind create cluster --config scripts/kind-config.yaml
    echo "✅ Kind cluster created successfully"
fi

# Verify cluster is ready
echo "🔍 Verifying cluster is ready..."
kubectl wait --for=condition=Ready nodes --all --timeout=300s

# Install Gateway API CRDs
echo "📦 Installing Gateway API CRDs..."
kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.0.0/standard-install.yaml

echo "⏳ Waiting for Gateway API CRDs to be ready..."
kubectl wait --for condition=established --timeout=60s crd/gateways.gateway.networking.k8s.io
kubectl wait --for condition=established --timeout=60s crd/httproutes.gateway.networking.k8s.io

# Install NGINX Ingress Controller (optional, for ingress support)
echo "🌐 Installing NGINX Ingress Controller..."
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml

echo "⏳ Waiting for NGINX Ingress Controller to be ready..."
kubectl wait --namespace ingress-nginx \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/component=controller \
  --timeout=90s

echo ""
echo "🎉 Local Kubernetes cluster is ready!"
echo ""
echo "Cluster info:"
kubectl cluster-info --context kind-gateway-cd
echo ""
echo "Next steps:"
echo "1. Build and load Docker images: ./scripts/build-and-load.sh"
echo "2. Install Gateway CD platform: ./scripts/install.sh"
echo "3. Access dashboard: kubectl port-forward svc/gateway-cd-web 3000:3000 -n gateway-cd"
echo ""
echo "To delete the cluster when done: kind delete cluster --name gateway-cd"