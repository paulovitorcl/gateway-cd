import React, { useState } from 'react'
import {
  Box,
  Button,
  Card,
  CardContent,
  TextField,
  Typography,
  Grid,
  IconButton,
  Switch,
  FormControlLabel,
  Alert,
} from '@mui/material'
import {
  ArrowBack as BackIcon,
  Add as AddIcon,
  Remove as RemoveIcon,
} from '@mui/icons-material'
import { useNavigate } from 'react-router-dom'
import { useMutation } from '@tanstack/react-query'
import { canaryApi, CanaryDeployment } from '../services/api'

interface TrafficStep {
  weight: number
  duration: string
  pause: boolean
}

const CreateCanary: React.FC = () => {
  const navigate = useNavigate()
  const [formData, setFormData] = useState({
    name: '',
    namespace: 'default',
    targetWorkload: '',
    serviceName: '',
    servicePort: 80,
    httpRoute: '',
    gateway: '',
    autoPromote: false,
    skipAnalysis: false,
    successRate: 0.95,
    maxLatency: 500,
  })

  const [trafficSteps, setTrafficSteps] = useState<TrafficStep[]>([
    { weight: 10, duration: '2m', pause: false },
    { weight: 25, duration: '2m', pause: false },
    { weight: 50, duration: '2m', pause: true },
    { weight: 100, duration: '', pause: false },
  ])

  const [error, setError] = useState<string | null>(null)

  const createMutation = useMutation({
    mutationFn: (canary: Partial<CanaryDeployment>) => canaryApi.create(canary),
    onSuccess: (response) => {
      navigate(`/canaries/${response.data.metadata.namespace}/${response.data.metadata.name}`)
    },
    onError: (error: any) => {
      setError(error.response?.data?.error || 'Failed to create canary deployment')
    },
  })

  const handleInputChange = (field: string, value: any) => {
    setFormData(prev => ({ ...prev, [field]: value }))
  }

  const handleStepChange = (index: number, field: keyof TrafficStep, value: any) => {
    setTrafficSteps(prev =>
      prev.map((step, i) => (i === index ? { ...step, [field]: value } : step))
    )
  }

  const addStep = () => {
    setTrafficSteps(prev => [
      ...prev,
      { weight: Math.min(100, (prev[prev.length - 1]?.weight || 0) + 25), duration: '2m', pause: false },
    ])
  }

  const removeStep = (index: number) => {
    if (trafficSteps.length > 1) {
      setTrafficSteps(prev => prev.filter((_, i) => i !== index))
    }
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    setError(null)

    const canaryDeployment: Partial<CanaryDeployment> = {
      metadata: {
        name: formData.name,
        namespace: formData.namespace,
        creationTimestamp: '',
      },
      spec: {
        targetRef: {
          apiVersion: 'apps/v1',
          kind: 'Deployment',
          name: formData.targetWorkload,
        },
        service: {
          name: formData.serviceName,
          port: formData.servicePort,
        },
        gateway: {
          httpRoute: formData.httpRoute,
          gateway: formData.gateway || undefined,
          namespace: formData.namespace,
        },
        trafficSplit: trafficSteps.map(step => ({
          weight: step.weight,
          duration: step.duration || undefined,
          pause: step.pause,
        })),
        autoPromote: formData.autoPromote,
        skipAnalysis: formData.skipAnalysis,
        analysis: formData.skipAnalysis ? undefined : {
          successRate: formData.successRate,
          maxLatency: formData.maxLatency,
          analysisInterval: '30s',
        },
      },
      status: {},
    }

    createMutation.mutate(canaryDeployment)
  }

  return (
    <Box>
      <Box sx={{ display: 'flex', alignItems: 'center', mb: 3 }}>
        <Button
          startIcon={<BackIcon />}
          onClick={() => navigate('/canaries')}
          sx={{ mr: 2 }}
        >
          Back
        </Button>
        <Typography variant="h4" component="h1">
          Create Canary Deployment
        </Typography>
      </Box>

      {error && (
        <Alert severity="error" sx={{ mb: 3 }}>
          {error}
        </Alert>
      )}

      <form onSubmit={handleSubmit}>
        <Grid container spacing={3}>
          {/* Basic Information */}
          <Grid item xs={12}>
            <Card>
              <CardContent>
                <Typography variant="h6" gutterBottom>
                  Basic Information
                </Typography>
                <Grid container spacing={2}>
                  <Grid item xs={12} md={6}>
                    <TextField
                      fullWidth
                      label="Name"
                      value={formData.name}
                      onChange={(e) => handleInputChange('name', e.target.value)}
                      required
                    />
                  </Grid>
                  <Grid item xs={12} md={6}>
                    <TextField
                      fullWidth
                      label="Namespace"
                      value={formData.namespace}
                      onChange={(e) => handleInputChange('namespace', e.target.value)}
                      required
                    />
                  </Grid>
                  <Grid item xs={12} md={6}>
                    <TextField
                      fullWidth
                      label="Target Workload"
                      value={formData.targetWorkload}
                      onChange={(e) => handleInputChange('targetWorkload', e.target.value)}
                      helperText="Name of the Deployment to canary"
                      required
                    />
                  </Grid>
                </Grid>
              </CardContent>
            </Card>
          </Grid>

          {/* Service Configuration */}
          <Grid item xs={12}>
            <Card>
              <CardContent>
                <Typography variant="h6" gutterBottom>
                  Service Configuration
                </Typography>
                <Grid container spacing={2}>
                  <Grid item xs={12} md={6}>
                    <TextField
                      fullWidth
                      label="Service Name"
                      value={formData.serviceName}
                      onChange={(e) => handleInputChange('serviceName', e.target.value)}
                      required
                    />
                  </Grid>
                  <Grid item xs={12} md={6}>
                    <TextField
                      fullWidth
                      label="Service Port"
                      type="number"
                      value={formData.servicePort}
                      onChange={(e) => handleInputChange('servicePort', parseInt(e.target.value))}
                      required
                    />
                  </Grid>
                </Grid>
              </CardContent>
            </Card>
          </Grid>

          {/* Gateway Configuration */}
          <Grid item xs={12}>
            <Card>
              <CardContent>
                <Typography variant="h6" gutterBottom>
                  Gateway API Configuration
                </Typography>
                <Grid container spacing={2}>
                  <Grid item xs={12} md={6}>
                    <TextField
                      fullWidth
                      label="HTTPRoute Name"
                      value={formData.httpRoute}
                      onChange={(e) => handleInputChange('httpRoute', e.target.value)}
                      required
                    />
                  </Grid>
                  <Grid item xs={12} md={6}>
                    <TextField
                      fullWidth
                      label="Gateway Name (Optional)"
                      value={formData.gateway}
                      onChange={(e) => handleInputChange('gateway', e.target.value)}
                    />
                  </Grid>
                </Grid>
              </CardContent>
            </Card>
          </Grid>

          {/* Traffic Split Configuration */}
          <Grid item xs={12}>
            <Card>
              <CardContent>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
                  <Typography variant="h6">
                    Traffic Split Steps
                  </Typography>
                  <Button
                    startIcon={<AddIcon />}
                    onClick={addStep}
                    size="small"
                  >
                    Add Step
                  </Button>
                </Box>

                {trafficSteps.map((step, index) => (
                  <Box key={index} sx={{ mb: 2, p: 2, border: '1px solid #e0e0e0', borderRadius: 1 }}>
                    <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 1 }}>
                      <Typography variant="subtitle2">
                        Step {index + 1}
                      </Typography>
                      {trafficSteps.length > 1 && (
                        <IconButton
                          size="small"
                          onClick={() => removeStep(index)}
                          color="error"
                        >
                          <RemoveIcon />
                        </IconButton>
                      )}
                    </Box>
                    <Grid container spacing={2}>
                      <Grid item xs={4}>
                        <TextField
                          fullWidth
                          label="Weight (%)"
                          type="number"
                          value={step.weight}
                          onChange={(e) => handleStepChange(index, 'weight', parseInt(e.target.value))}
                          inputProps={{ min: 0, max: 100 }}
                        />
                      </Grid>
                      <Grid item xs={4}>
                        <TextField
                          fullWidth
                          label="Duration"
                          value={step.duration}
                          onChange={(e) => handleStepChange(index, 'duration', e.target.value)}
                          helperText="e.g., 2m, 30s"
                        />
                      </Grid>
                      <Grid item xs={4}>
                        <FormControlLabel
                          control={
                            <Switch
                              checked={step.pause}
                              onChange={(e) => handleStepChange(index, 'pause', e.target.checked)}
                            />
                          }
                          label="Pause for approval"
                        />
                      </Grid>
                    </Grid>
                  </Box>
                ))}
              </CardContent>
            </Card>
          </Grid>

          {/* Analysis Configuration */}
          <Grid item xs={12}>
            <Card>
              <CardContent>
                <Typography variant="h6" gutterBottom>
                  Analysis Configuration
                </Typography>
                <Grid container spacing={2}>
                  <Grid item xs={12}>
                    <FormControlLabel
                      control={
                        <Switch
                          checked={formData.skipAnalysis}
                          onChange={(e) => handleInputChange('skipAnalysis', e.target.checked)}
                        />
                      }
                      label="Skip analysis (deploy without health checks)"
                    />
                  </Grid>
                  {!formData.skipAnalysis && (
                    <>
                      <Grid item xs={12} md={6}>
                        <TextField
                          fullWidth
                          label="Minimum Success Rate"
                          type="number"
                          value={formData.successRate}
                          onChange={(e) => handleInputChange('successRate', parseFloat(e.target.value))}
                          inputProps={{ min: 0, max: 1, step: 0.01 }}
                          helperText="0.0 to 1.0"
                        />
                      </Grid>
                      <Grid item xs={12} md={6}>
                        <TextField
                          fullWidth
                          label="Maximum Latency (ms)"
                          type="number"
                          value={formData.maxLatency}
                          onChange={(e) => handleInputChange('maxLatency', parseInt(e.target.value))}
                          inputProps={{ min: 0 }}
                        />
                      </Grid>
                    </>
                  )}
                  <Grid item xs={12}>
                    <FormControlLabel
                      control={
                        <Switch
                          checked={formData.autoPromote}
                          onChange={(e) => handleInputChange('autoPromote', e.target.checked)}
                        />
                      }
                      label="Auto-promote on success"
                    />
                  </Grid>
                </Grid>
              </CardContent>
            </Card>
          </Grid>

          {/* Submit */}
          <Grid item xs={12}>
            <Box sx={{ display: 'flex', gap: 2 }}>
              <Button
                type="submit"
                variant="contained"
                disabled={createMutation.isPending}
                size="large"
              >
                {createMutation.isPending ? 'Creating...' : 'Create Canary Deployment'}
              </Button>
              <Button
                variant="outlined"
                onClick={() => navigate('/canaries')}
                size="large"
              >
                Cancel
              </Button>
            </Box>
          </Grid>
        </Grid>
      </form>
    </Box>
  )
}

export default CreateCanary