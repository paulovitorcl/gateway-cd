.PHONY: build test deploy install clean

# Build the controller and API server
build:
	go build -o bin/controller cmd/controller/main.go
	go build -o bin/api-server cmd/api-server/main.go

# Install dependencies
install:
	go mod download
	cd web/dashboard && npm install

# Run tests
test:
	go test -v ./...

# Build Docker images
docker-build:
	docker build -t gateway-cd-controller:latest -f deploy/Dockerfile.controller .
	docker build -t gateway-cd-api:latest -f deploy/Dockerfile.api .
	cd web/dashboard && docker build -t gateway-cd-web:latest -f Dockerfile .

# Deploy to Kubernetes
deploy:
	kubectl apply -f deploy/k8s/

# Clean build artifacts
clean:
	rm -rf bin/
	go clean

# Development server
dev-controller:
	go run cmd/controller/main.go

dev-api:
	go run cmd/api-server/main.go

dev-web:
	cd web/dashboard && npm start

# Generate CRDs
generate:
	controller-gen crd paths="./pkg/api/..." output:crd:artifacts:config=deploy/k8s/crds/