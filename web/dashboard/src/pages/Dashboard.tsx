import React from 'react'
import {
  Box,
  Card,
  CardContent,
  Grid,
  Typography,
  Chip,
  LinearProgress,
} from '@mui/material'
import {
  Rocket as RocketIcon,
  CheckCircle as SuccessIcon,
  Error as ErrorIcon,
  Pause as PauseIcon,
} from '@mui/icons-material'
import { useQuery } from '@tanstack/react-query'
import { canaryApi } from '../services/api'
import type { CanaryDeployment } from '../services/api'

const Dashboard: React.FC = () => {
  const { data: canaries = [], isLoading } = useQuery({
    queryKey: ['canaries'],
    queryFn: () => canaryApi.list().then(res => res.data),
  })

  const getPhaseIcon = (phase: string) => {
    switch (phase) {
      case 'Succeeded':
        return <SuccessIcon color="success" />
      case 'Failed':
        return <ErrorIcon color="error" />
      case 'Paused':
        return <PauseIcon color="warning" />
      default:
        return <RocketIcon color="primary" />
    }
  }

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

  const activeCanaries = canaries.filter(c =>
    c.status.phase === 'Progressing' || c.status.phase === 'Paused'
  )
  const completedCanaries = canaries.filter(c =>
    c.status.phase === 'Succeeded' || c.status.phase === 'Failed'
  )

  const stats = {
    total: canaries.length,
    active: activeCanaries.length,
    succeeded: canaries.filter(c => c.status.phase === 'Succeeded').length,
    failed: canaries.filter(c => c.status.phase === 'Failed').length,
  }

  if (isLoading) {
    return <LinearProgress />
  }

  return (
    <Box>
      <Typography variant="h4" component="h1" gutterBottom>
        Dashboard
      </Typography>

      {/* Stats Cards */}
      <Grid container spacing={3} sx={{ mb: 4 }}>
        <Grid item xs={12} sm={6} md={3}>
          <Card>
            <CardContent>
              <Typography color="textSecondary" gutterBottom>
                Total Deployments
              </Typography>
              <Typography variant="h4">
                {stats.total}
              </Typography>
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <Card>
            <CardContent>
              <Typography color="textSecondary" gutterBottom>
                Active
              </Typography>
              <Typography variant="h4" color="primary">
                {stats.active}
              </Typography>
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <Card>
            <CardContent>
              <Typography color="textSecondary" gutterBottom>
                Succeeded
              </Typography>
              <Typography variant="h4" color="success.main">
                {stats.succeeded}
              </Typography>
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <Card>
            <CardContent>
              <Typography color="textSecondary" gutterBottom>
                Failed
              </Typography>
              <Typography variant="h4" color="error.main">
                {stats.failed}
              </Typography>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      {/* Active Deployments */}
      <Grid container spacing={3}>
        <Grid item xs={12} md={6}>
          <Card>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                Active Canary Deployments
              </Typography>
              {activeCanaries.length === 0 ? (
                <Typography color="textSecondary">
                  No active deployments
                </Typography>
              ) : (
                <Box>
                  {activeCanaries.map((canary) => (
                    <Box
                      key={`${canary.metadata.namespace}/${canary.metadata.name}`}
                      sx={{
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'space-between',
                        py: 1,
                        borderBottom: '1px solid',
                        borderColor: 'divider',
                        '&:last-child': { borderBottom: 'none' },
                      }}
                    >
                      <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                        {getPhaseIcon(canary.status.phase || '')}
                        <Typography variant="subtitle2">
                          {canary.metadata.namespace}/{canary.metadata.name}
                        </Typography>
                      </Box>
                      <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                        <Typography variant="body2" color="textSecondary">
                          {canary.status.canaryWeight || 0}% canary
                        </Typography>
                        <Chip
                          label={canary.status.phase}
                          color={getPhaseColor(canary.status.phase || '')}
                          size="small"
                        />
                      </Box>
                    </Box>
                  ))}
                </Box>
              )}
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12} md={6}>
          <Card>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                Recent Completions
              </Typography>
              {completedCanaries.length === 0 ? (
                <Typography color="textSecondary">
                  No completed deployments
                </Typography>
              ) : (
                <Box>
                  {completedCanaries.slice(0, 5).map((canary) => (
                    <Box
                      key={`${canary.metadata.namespace}/${canary.metadata.name}`}
                      sx={{
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'space-between',
                        py: 1,
                        borderBottom: '1px solid',
                        borderColor: 'divider',
                        '&:last-child': { borderBottom: 'none' },
                      }}
                    >
                      <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                        {getPhaseIcon(canary.status.phase || '')}
                        <Typography variant="subtitle2">
                          {canary.metadata.namespace}/{canary.metadata.name}
                        </Typography>
                      </Box>
                      <Chip
                        label={canary.status.phase}
                        color={getPhaseColor(canary.status.phase || '')}
                        size="small"
                      />
                    </Box>
                  ))}
                </Box>
              )}
            </CardContent>
          </Card>
        </Grid>
      </Grid>
    </Box>
  )
}

export default Dashboard