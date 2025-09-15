#!/bin/bash

set -e

echo "🔨 Building and loading Docker images for Gateway CD..."

# Check if kind cluster exists
if ! kind get clusters | grep -q "^gateway-cd$"; then
    echo "❌ Error: kind cluster 'gateway-cd' not found."
    echo "Please run './scripts/setup-local.sh' first."
    exit 1
fi

# Set context to kind cluster
kubectl config use-context kind-gateway-cd

echo "🏗️  Building Gateway CD Controller..."
docker build -t gateway-cd-controller:latest -f deploy/Dockerfile.controller .

echo "🏗️  Building Gateway CD API Server..."
docker build -t gateway-cd-api:latest -f deploy/Dockerfile.api .

echo "🏗️  Building Gateway CD Web Dashboard..."
cd web/dashboard
docker build -t gateway-cd-web:latest .
cd ../..

echo "📦 Loading images into kind cluster..."
kind load docker-image gateway-cd-controller:latest --name gateway-cd
kind load docker-image gateway-cd-api:latest --name gateway-cd
kind load docker-image gateway-cd-web:latest --name gateway-cd

echo "✅ All images built and loaded successfully!"
echo ""
echo "Images loaded in cluster:"
docker exec -it gateway-cd-control-plane crictl images | grep gateway-cd || echo "No gateway-cd images found yet"
echo ""
echo "Next step: Run './scripts/install.sh' to deploy the platform"