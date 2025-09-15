#!/bin/bash

set -e

echo "🎯 Gateway CD Platform Demo"
echo "=========================="

# Check if platform is installed
if ! kubectl get deployment gateway-cd-controller -n gateway-cd &> /dev/null; then
    echo "❌ Gateway CD platform not found. Please install it first:"
    echo "   ./scripts/setup-local.sh"
    echo "   ./scripts/build-and-load.sh"
    echo "   ./scripts/install.sh"
    exit 1
fi

echo "📋 Platform Status:"
kubectl get pods -n gateway-cd

echo ""
echo "🚀 Creating demo application..."

# Create demo namespace
kubectl create namespace demo --dry-run=client -o yaml | kubectl apply -f -

# Create demo application
cat << EOF | kubectl apply -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: demo-app
  namespace: demo
  labels:
    app: demo-app
    version: stable
spec:
  replicas: 3
  selector:
    matchLabels:
      app: demo-app
      version: stable
  template:
    metadata:
      labels:
        app: demo-app
        version: stable
    spec:
      containers:
      - name: app
        image: nginx:1.21
        ports:
        - containerPort: 80
        env:
        - name: VERSION
          value: "stable"
        volumeMounts:
        - name: html
          mountPath: /usr/share/nginx/html
      volumes:
      - name: html
        configMap:
          name: demo-app-html
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: demo-app-canary
  namespace: demo
  labels:
    app: demo-app
    version: canary
spec:
  replicas: 1
  selector:
    matchLabels:
      app: demo-app
      version: canary
  template:
    metadata:
      labels:
        app: demo-app
        version: canary
    spec:
      containers:
      - name: app
        image: nginx:1.22
        ports:
        - containerPort: 80
        env:
        - name: VERSION
          value: "canary"
        volumeMounts:
        - name: html
          mountPath: /usr/share/nginx/html
      volumes:
      - name: html
        configMap:
          name: demo-app-canary-html
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: demo-app-html
  namespace: demo
data:
  index.html: |
    <!DOCTYPE html>
    <html>
    <head><title>Demo App - Stable</title></head>
    <body style="background-color: #4CAF50; color: white; text-align: center; padding: 50px;">
      <h1>Demo Application</h1>
      <h2>Version: STABLE</h2>
      <p>This is the stable version of the demo application.</p>
    </body>
    </html>
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: demo-app-canary-html
  namespace: demo
data:
  index.html: |
    <!DOCTYPE html>
    <html>
    <head><title>Demo App - Canary</title></head>
    <body style="background-color: #FF9800; color: white; text-align: center; padding: 50px;">
      <h1>Demo Application</h1>
      <h2>Version: CANARY</h2>
      <p>This is the canary version of the demo application.</p>
    </body>
    </html>
---
apiVersion: v1
kind: Service
metadata:
  name: demo-app-service
  namespace: demo
spec:
  selector:
    app: demo-app
    version: stable
  ports:
  - port: 80
    targetPort: 80
---
apiVersion: v1
kind: Service
metadata:
  name: demo-app-service-canary
  namespace: demo
spec:
  selector:
    app: demo-app
    version: canary
  ports:
  - port: 80
    targetPort: 80
EOF

echo "⏳ Waiting for demo applications to be ready..."
kubectl wait --for=condition=available --timeout=120s deployment/demo-app -n demo
kubectl wait --for=condition=available --timeout=120s deployment/demo-app-canary -n demo

# Install Gateway (using NGINX as example)
echo "🌐 Creating Gateway and HTTPRoute..."
cat << EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: demo-gateway
  namespace: demo
spec:
  gatewayClassName: nginx
  listeners:
  - name: http
    port: 80
    protocol: HTTP
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: demo-app-route
  namespace: demo
spec:
  parentRefs:
  - name: demo-gateway
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /
    backendRefs:
    - name: demo-app-service
      port: 80
      weight: 100
EOF

echo "⏳ Waiting for Gateway to be ready..."
sleep 10

echo ""
echo "🎯 Creating Canary Deployment..."
cat << EOF | kubectl apply -f -
apiVersion: gateway-cd.io/v1alpha1
kind: CanaryDeployment
metadata:
  name: demo-app-canary
  namespace: demo
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: demo-app-canary
  service:
    name: demo-app-service
    port: 80
  gateway:
    httpRoute: demo-app-route
    namespace: demo
  trafficSplit:
    - weight: 20
      duration: "30s"
      pause: false
    - weight: 50
      duration: "30s"
      pause: true
    - weight: 100
      duration: ""
      pause: false
  analysis:
    successRate: 0.90
    maxLatency: 1000
    analysisInterval: "15s"
  autoPromote: false
  skipAnalysis: true
EOF

echo ""
echo "✅ Demo setup complete!"
echo ""
echo "🎯 Demo Resources Created:"
echo "- Demo applications (stable + canary versions)"
echo "- Gateway and HTTPRoute for traffic management"
echo "- CanaryDeployment resource"
echo ""
echo "📊 Monitor the canary deployment:"
echo "kubectl get canarydeployment -n demo -w"
echo ""
echo "🌐 Access the application:"
echo "kubectl port-forward svc/demo-gateway 8080:80 -n demo"
echo "Then visit: http://localhost:8080"
echo ""
echo "🎛️ Control the deployment:"
echo "# Resume paused deployment:"
echo "kubectl annotate canarydeployment demo-app-canary gateway-cd.io/resume=true -n demo"
echo ""
echo "# Abort deployment:"
echo "kubectl annotate canarydeployment demo-app-canary gateway-cd.io/abort=true -n demo"
echo ""
echo "🗑️ Cleanup demo:"
echo "kubectl delete namespace demo"