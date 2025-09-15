import React from 'react'
import {
  Box,
  Button,
  Card,
  CardContent,
  Chip,
  Grid,
  LinearProgress,
  Typography,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Alert,
} from '@mui/material'
import {
  PlayArrow as ResumeIcon,
  Pause as PauseIcon,
  Stop as AbortIcon,
  TrendingUp as PromoteIcon,
  ArrowBack as BackIcon,
} from '@mui/icons-material'
import { useParams, useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts'
import { canaryApi } from '../services/api'

const CanaryDetail: React.FC = () => {
  const { namespace, name } = useParams<{ namespace: string; name: string }>()
  const navigate = useNavigate()
  const queryClient = useQueryClient()

  const { data: canary, isLoading: canaryLoading } = useQuery({
    queryKey: ['canary', namespace, name],
    queryFn: () => canaryApi.get(namespace!, name!).then(res => res.data),
    enabled: !!namespace && !!name,
  })

  const { data: status, isLoading: statusLoading } = useQuery({
    queryKey: ['canary-status', namespace, name],
    queryFn: () => canaryApi.getStatus(namespace!, name!).then(res => res.data),
    enabled: !!namespace && !!name,
  })

  const { data: metrics } = useQuery({
    queryKey: ['canary-metrics', namespace, name],
    queryFn: () => canaryApi.getMetrics(namespace!, name!).then(res => res.data),
    enabled: !!namespace && !!name,
  })

  const { data: history = [] } = useQuery({
    queryKey: ['canary-history', namespace, name],
    queryFn: () => canaryApi.getHistory(namespace!, name!, 20).then(res => res.data),
    enabled: !!namespace && !!name,
  })

  const resumeMutation = useMutation({
    mutationFn: () => canaryApi.resume(namespace!, name!),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['canary', namespace, name] })
      queryClient.invalidateQueries({ queryKey: ['canary-status', namespace, name] })
    },
  })

  const pauseMutation = useMutation({
    mutationFn: () => canaryApi.pause(namespace!, name!),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['canary', namespace, name] })
      queryClient.invalidateQueries({ queryKey: ['canary-status', namespace, name] })
    },
  })

  const abortMutation = useMutation({
    mutationFn: () => canaryApi.abort(namespace!, name!),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['canary', namespace, name] })
      queryClient.invalidateQueries({ queryKey: ['canary-status', namespace, name] })
    },
  })

  const promoteMutation = useMutation({
    mutationFn: () => canaryApi.promote(namespace!, name!),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['canary', namespace, name] })
      queryClient.invalidateQueries({ queryKey: ['canary-status', namespace, name] })
    },
  })

  const getPhaseColor = (phase: string): 'default' | 'primary' | 'secondary' | 'error' | 'info' | 'success' | 'warning' => {
    switch (phase) {
      case 'Succeeded':
        return 'success'
      case 'Failed':
        return 'error'
      case 'Paused':
        return 'warning'
      case 'Progressing':
        return 'primary'
      default:
        return 'default'
    }
  }

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString()
  }

  // Generate chart data from history
  const chartData = history.map((entry, index) => ({
    step: index + 1,
    weight: entry.weight,
    timestamp: formatDate(entry.timestamp),
  })).reverse()

  if (canaryLoading || statusLoading) {
    return <LinearProgress />
  }

  if (!canary || !status) {
    return (
      <Alert severity="error">
        Canary deployment not found
      </Alert>
    )
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
          {namespace}/{name}
        </Typography>
      </Box>

      <Grid container spacing={3}>
        {/* Status Overview */}
        <Grid item xs={12} md={8}>
          <Card sx={{ mb: 3 }}>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                Deployment Status
              </Typography>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, mb: 2 }}>
                <Chip
                  label={status.phase}
                  color={getPhaseColor(status.phase)}
                />
                <Typography variant="body2" color="textSecondary">
                  Step {status.currentStep} of {status.totalSteps}
                </Typography>
              </Box>
              <Typography variant="body1" sx={{ mb: 2 }}>
                {status.message}
              </Typography>

              {/* Progress Bar */}
              <Box sx={{ mb: 2 }}>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 1 }}>
                  <Typography variant="body2">Traffic Split</Typography>
                  <Typography variant="body2">
                    {status.canaryWeight}% canary / {status.stableWeight}% stable
                  </Typography>
                </Box>
                <LinearProgress
                  variant="determinate"
                  value={status.canaryWeight}
                  sx={{ height: 8, borderRadius: 4 }}
                />
              </Box>

              {/* Analysis Results */}
              {status.analysisRun && (
                <Box sx={{ mt: 2 }}>
                  <Typography variant="subtitle2" gutterBottom>
                    Analysis Results
                  </Typography>
                  <Grid container spacing={2}>
                    <Grid item xs={6}>
                      <Typography variant="body2" color="textSecondary">
                        Success Rate
                      </Typography>
                      <Typography variant="body1">
                        {(status.analysisRun.successRate * 100).toFixed(2)}%
                      </Typography>
                    </Grid>
                    <Grid item xs={6}>
                      <Typography variant="body2" color="textSecondary">
                        Avg Latency
                      </Typography>
                      <Typography variant="body1">
                        {status.analysisRun.averageLatency}ms
                      </Typography>
                    </Grid>
                  </Grid>
                </Box>
              )}
            </CardContent>
          </Card>

          {/* Traffic Weight Chart */}
          <Card sx={{ mb: 3 }}>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                Traffic Weight History
              </Typography>
              <ResponsiveContainer width="100%" height={300}>
                <LineChart data={chartData}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="step" />
                  <YAxis domain={[0, 100]} />
                  <Tooltip
                    labelFormatter={(label) => `Step ${label}`}
                    formatter={(value) => [`${value}%`, 'Canary Weight']}
                  />
                  <Legend />
                  <Line
                    type="monotone"
                    dataKey="weight"
                    stroke="#1976d2"
                    strokeWidth={2}
                    dot={{ fill: '#1976d2' }}
                    name="Canary Weight"
                  />
                </LineChart>
              </ResponsiveContainer>
            </CardContent>
          </Card>
        </Grid>

        {/* Controls and Metrics */}
        <Grid item xs={12} md={4}>
          <Card sx={{ mb: 3 }}>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                Controls
              </Typography>
              <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1 }}>
                {status.canResume && (
                  <Button
                    variant="contained"
                    startIcon={<ResumeIcon />}
                    onClick={() => resumeMutation.mutate()}
                    disabled={resumeMutation.isPending}
                    fullWidth
                  >
                    Resume
                  </Button>
                )}
                {status.canPause && (
                  <Button
                    variant="outlined"
                    startIcon={<PauseIcon />}
                    onClick={() => pauseMutation.mutate()}
                    disabled={pauseMutation.isPending}
                    fullWidth
                  >
                    Pause
                  </Button>
                )}
                {status.canPromote && (
                  <Button
                    variant="contained"
                    color="success"
                    startIcon={<PromoteIcon />}
                    onClick={() => promoteMutation.mutate()}
                    disabled={promoteMutation.isPending}
                    fullWidth
                  >
                    Promote
                  </Button>
                )}
                {status.canAbort && (
                  <Button
                    variant="outlined"
                    color="error"
                    startIcon={<AbortIcon />}
                    onClick={() => abortMutation.mutate()}
                    disabled={abortMutation.isPending}
                    fullWidth
                  >
                    Abort
                  </Button>
                )}
              </Box>
            </CardContent>
          </Card>

          {/* Live Metrics */}
          {metrics && (
            <Card sx={{ mb: 3 }}>
              <CardContent>
                <Typography variant="h6" gutterBottom>
                  Live Metrics
                </Typography>
                <Grid container spacing={2}>
                  <Grid item xs={6}>
                    <Typography variant="body2" color="textSecondary">
                      Success Rate
                    </Typography>
                    <Typography variant="h6">
                      {(metrics.successRate * 100).toFixed(2)}%
                    </Typography>
                  </Grid>
                  <Grid item xs={6}>
                    <Typography variant="body2" color="textSecondary">
                      Avg Latency
                    </Typography>
                    <Typography variant="h6">
                      {metrics.averageLatency}ms
                    </Typography>
                  </Grid>
                  <Grid item xs={6}>
                    <Typography variant="body2" color="textSecondary">
                      Requests/sec
                    </Typography>
                    <Typography variant="h6">
                      {metrics.throughput.toFixed(1)}
                    </Typography>
                  </Grid>
                  <Grid item xs={6}>
                    <Typography variant="body2" color="textSecondary">
                      Error Rate
                    </Typography>
                    <Typography variant="h6">
                      {(metrics.errorRate * 100).toFixed(2)}%
                    </Typography>
                  </Grid>
                </Grid>
              </CardContent>
            </Card>
          )}
        </Grid>

        {/* Traffic Split Configuration */}
        <Grid item xs={12}>
          <Card>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                Traffic Split Configuration
              </Typography>
              <TableContainer>
                <Table size="small">
                  <TableHead>
                    <TableRow>
                      <TableCell>Step</TableCell>
                      <TableCell>Weight</TableCell>
                      <TableCell>Duration</TableCell>
                      <TableCell>Pause</TableCell>
                      <TableCell>Status</TableCell>
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {canary.spec.trafficSplit.map((step, index) => (
                      <TableRow key={index}>
                        <TableCell>{index + 1}</TableCell>
                        <TableCell>{step.weight}%</TableCell>
                        <TableCell>{step.duration || 'Auto'}</TableCell>
                        <TableCell>{step.pause ? 'Yes' : 'No'}</TableCell>
                        <TableCell>
                          {index < status.currentStep ? (
                            <Chip label="Completed" color="success" size="small" />
                          ) : index === status.currentStep ? (
                            <Chip label="Current" color="primary" size="small" />
                          ) : (
                            <Chip label="Pending" color="default" size="small" />
                          )}
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </TableContainer>
            </CardContent>
          </Card>
        </Grid>
      </Grid>
    </Box>
  )
}

export default CanaryDetail