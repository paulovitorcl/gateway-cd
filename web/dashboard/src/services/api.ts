import axios from 'axios'

const api = axios.create({
  baseURL: '/api/v1',
  timeout: 30000,
})

export interface CanaryDeployment {
  metadata: {
    name: string
    namespace: string
    creationTimestamp: string
    annotations?: Record<string, string>
  }
  spec: {
    targetRef: {
      apiVersion: string
      kind: string
      name: string
    }
    service: {
      name: string
      port: number
    }
    gateway: {
      httpRoute: string
      gateway?: string
      namespace?: string
    }
    trafficSplit: Array<{
      weight: number
      duration?: string
      pause?: boolean
    }>
    analysis?: {
      metrics?: Array<{
        name: string
        query: string
        threshold: number
        operator: string
      }>
      successRate?: number
      maxLatency?: number
      analysisInterval?: string
    }
    autoPromote?: boolean
    skipAnalysis?: boolean
  }
  status: {
    phase?: string
    message?: string
    currentStep?: number
    canaryWeight?: number
    stableWeight?: number
    lastTransitionTime?: string
    conditions?: Array<{
      type: string
      status: string
      reason: string
      message: string
      lastTransitionTime: string
    }>
    analysisRun?: {
      phase?: string
      successRate?: number
      averageLatency?: number
      metricResults?: Array<{
        name: string
        value: number
        threshold: number
        passed: boolean
      }>
      startedAt?: string
      completedAt?: string
    }
  }
}

export interface CanaryStatus {
  phase: string
  message: string
  currentStep: number
  totalSteps: number
  canaryWeight: number
  stableWeight: number
  lastTransition: string
  conditions: Array<{
    type: string
    status: string
    reason: string
    message: string
    lastTransitionTime: string
  }>
  analysisRun?: {
    phase: string
    successRate: number
    averageLatency: number
    metricResults: Array<{
      name: string
      value: number
      threshold: number
      passed: boolean
    }>
    startedAt: string
    completedAt: string
  }
  canPause: boolean
  canResume: boolean
  canAbort: boolean
  canPromote: boolean
}

export interface CanaryMetrics {
  successRate: number
  averageLatency: number
  requestCount: number
  errorRate: number
  throughput: number
  timestamp: string
}

export interface HistoryEntry {
  timestamp: string
  phase: string
  step: number
  weight: number
  message: string
}

export const canaryApi = {
  // List all canary deployments
  list: (namespace?: string) =>
    api.get<CanaryDeployment[]>('/canaries', {
      params: namespace ? { namespace } : {},
    }),

  // Get a specific canary deployment
  get: (namespace: string, name: string) =>
    api.get<CanaryDeployment>(`/canaries/${namespace}/${name}`),

  // Create a new canary deployment
  create: (canary: Partial<CanaryDeployment>) =>
    api.post<CanaryDeployment>('/canaries', canary),

  // Update a canary deployment
  update: (namespace: string, name: string, canary: Partial<CanaryDeployment>) =>
    api.put<CanaryDeployment>(`/canaries/${namespace}/${name}`, canary),

  // Delete a canary deployment
  delete: (namespace: string, name: string) =>
    api.delete(`/canaries/${namespace}/${name}`),

  // Control operations
  resume: (namespace: string, name: string) =>
    api.post(`/canaries/${namespace}/${name}/resume`),

  pause: (namespace: string, name: string) =>
    api.post(`/canaries/${namespace}/${name}/pause`),

  abort: (namespace: string, name: string) =>
    api.post(`/canaries/${namespace}/${name}/abort`),

  promote: (namespace: string, name: string) =>
    api.post(`/canaries/${namespace}/${name}/promote`),

  // Status and metrics
  getStatus: (namespace: string, name: string) =>
    api.get<CanaryStatus>(`/canaries/${namespace}/${name}/status`),

  getMetrics: (namespace: string, name: string) =>
    api.get<CanaryMetrics>(`/canaries/${namespace}/${name}/metrics`),

  getHistory: (namespace: string, name: string, limit?: number) =>
    api.get<HistoryEntry[]>(`/canaries/${namespace}/${name}/history`, {
      params: limit ? { limit } : {},
    }),

  // Health check
  health: () => api.get('/health'),
}

export default api