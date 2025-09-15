# Gateway CD - Kubernetes Gateway API Canary Deployment Platform

A platform for managing canary deployments using Kubernetes Gateway API with visual monitoring and interactive controls.

## Features

- 🚀 **Canary Deployments**: Gradual traffic shifting using Kubernetes Gateway API
- 📊 **Visual Monitoring**: Real-time deployment metrics and health monitoring
- 🔄 **Automated Rollback**: Safety checks and automatic rollback on failures
- 🎛️ **Manual Controls**: Pause, resume, or abort deployments
- 📈 **Historical Data**: Deployment timeline and audit logs
- 🔗 **CI/CD Integration**: API-driven deployment management

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Web Dashboard │    │   REST API      │    │   Controller    │
│   (React)       │◄──►│   (Go/Gin)      │◄──►│   (K8s)         │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │                        │
                                ▼                        ▼
                       ┌─────────────────┐    ┌─────────────────┐
                       │   Database      │    │   Gateway API   │
                       │   (PostgreSQL)  │    │   (HTTPRoute)   │
                       └─────────────────┘    └─────────────────┘
```

## Quick Start

### Prerequisites
- Kubernetes cluster with Gateway API CRDs installed
- Go 1.21+
- Node.js 18+
- Docker

### Installation

1. Clone the repository
2. Install dependencies: `make install`
3. Deploy to cluster: `make deploy`
4. Access dashboard: `kubectl port-forward svc/gateway-cd-web 3000:3000`

## Project Structure

```
gateway-cd/
├── cmd/                    # Main applications
├── pkg/                    # Library code
│   ├── controller/        # Kubernetes controller
│   ├── api/              # REST API handlers
│   ├── metrics/          # Metrics collection
│   └── gateway/          # Gateway API integration
├── internal/             # Private application code
│   ├── config/          # Configuration
│   ├── db/              # Database models
│   └── models/          # Domain models
├── web/dashboard/        # React dashboard
└── deploy/k8s/          # Kubernetes manifests
```