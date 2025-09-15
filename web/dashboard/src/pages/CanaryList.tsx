import React from 'react'
import {
  Box,
  Button,
  Card,
  CardContent,
  Chip,
  IconButton,
  LinearProgress,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Typography,
  Menu,
  MenuItem,
} from '@mui/material'
import {
  Add as AddIcon,
  MoreVert as MoreIcon,
  Visibility as ViewIcon,
  PlayArrow as ResumeIcon,
  Pause as PauseIcon,
  Stop as AbortIcon,
  TrendingUp as PromoteIcon,
} from '@mui/icons-material'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Link as RouterLink, useNavigate } from 'react-router-dom'
import { canaryApi, CanaryDeployment } from '../services/api'

const CanaryList: React.FC = () => {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [anchorEl, setAnchorEl] = React.useState<null | HTMLElement>(null)
  const [selectedCanary, setSelectedCanary] = React.useState<CanaryDeployment | null>(null)

  const { data: canaries = [], isLoading } = useQuery({
    queryKey: ['canaries'],
    queryFn: () => canaryApi.list().then(res => res.data),
  })

  const resumeMutation = useMutation({
    mutationFn: ({ namespace, name }: { namespace: string; name: string }) =>
      canaryApi.resume(namespace, name),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['canaries'] })
      handleMenuClose()
    },
  })

  const pauseMutation = useMutation({
    mutationFn: ({ namespace, name }: { namespace: string; name: string }) =>
      canaryApi.pause(namespace, name),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['canaries'] })
      handleMenuClose()
    },
  })

  const abortMutation = useMutation({
    mutationFn: ({ namespace, name }: { namespace: string; name: string }) =>
      canaryApi.abort(namespace, name),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['canaries'] })
      handleMenuClose()
    },
  })

  const promoteMutation = useMutation({
    mutationFn: ({ namespace, name }: { namespace: string; name: string }) =>
      canaryApi.promote(namespace, name),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['canaries'] })
      handleMenuClose()
    },
  })

  const handleMenuOpen = (event: React.MouseEvent<HTMLElement>, canary: CanaryDeployment) => {
    setAnchorEl(event.currentTarget)
    setSelectedCanary(canary)
  }

  const handleMenuClose = () => {
    setAnchorEl(null)
    setSelectedCanary(null)
  }

  const handleAction = (action: string) => {
    if (!selectedCanary) return

    const { namespace, name } = selectedCanary.metadata

    switch (action) {
      case 'view':
        navigate(`/canaries/${namespace}/${name}`)
        break
      case 'resume':
        resumeMutation.mutate({ namespace, name })
        break
      case 'pause':
        pauseMutation.mutate({ namespace, name })
        break
      case 'abort':
        abortMutation.mutate({ namespace, name })
        break
      case 'promote':
        promoteMutation.mutate({ namespace, name })
        break
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

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString()
  }

  const canShowResume = (canary: CanaryDeployment) => canary.status.phase === 'Paused'
  const canShowPause = (canary: CanaryDeployment) => canary.status.phase === 'Progressing'
  const canShowAbort = (canary: CanaryDeployment) =>
    canary.status.phase === 'Progressing' || canary.status.phase === 'Paused'
  const canShowPromote = (canary: CanaryDeployment) => canary.status.phase === 'Paused'

  if (isLoading) {
    return <LinearProgress />
  }

  return (
    <Box>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Typography variant="h4" component="h1">
          Canary Deployments
        </Typography>
        <Button
          variant="contained"
          startIcon={<AddIcon />}
          component={RouterLink}
          to="/canaries/new"
        >
          Create Canary
        </Button>
      </Box>

      <Card>
        <CardContent>
          <TableContainer>
            <Table>
              <TableHead>
                <TableRow>
                  <TableCell>Name</TableCell>
                  <TableCell>Namespace</TableCell>
                  <TableCell>Phase</TableCell>
                  <TableCell>Progress</TableCell>
                  <TableCell>Canary Weight</TableCell>
                  <TableCell>Created</TableCell>
                  <TableCell align="right">Actions</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {canaries.map((canary) => (
                  <TableRow key={`${canary.metadata.namespace}/${canary.metadata.name}`}>
                    <TableCell>
                      <Typography variant="subtitle2">
                        {canary.metadata.name}
                      </Typography>
                    </TableCell>
                    <TableCell>{canary.metadata.namespace}</TableCell>
                    <TableCell>
                      <Chip
                        label={canary.status.phase || 'Unknown'}
                        color={getPhaseColor(canary.status.phase || '')}
                        size="small"
                      />
                    </TableCell>
                    <TableCell>
                      {canary.status.currentStep !== undefined && canary.spec.trafficSplit ? (
                        <Typography variant="body2">
                          {canary.status.currentStep} / {canary.spec.trafficSplit.length}
                        </Typography>
                      ) : (
                        '-'
                      )}
                    </TableCell>
                    <TableCell>
                      <Typography variant="body2">
                        {canary.status.canaryWeight || 0}%
                      </Typography>
                    </TableCell>
                    <TableCell>
                      <Typography variant="body2">
                        {formatDate(canary.metadata.creationTimestamp)}
                      </Typography>
                    </TableCell>
                    <TableCell align="right">
                      <IconButton
                        onClick={(e) => handleMenuOpen(e, canary)}
                        size="small"
                      >
                        <MoreIcon />
                      </IconButton>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </TableContainer>

          {canaries.length === 0 && (
            <Box sx={{ textAlign: 'center', py: 4 }}>
              <Typography variant="body1" color="textSecondary">
                No canary deployments found. Create one to get started.
              </Typography>
            </Box>
          )}
        </CardContent>
      </Card>

      <Menu
        anchorEl={anchorEl}
        open={Boolean(anchorEl)}
        onClose={handleMenuClose}
      >
        <MenuItem onClick={() => handleAction('view')}>
          <ViewIcon sx={{ mr: 1 }} />
          View Details
        </MenuItem>
        {selectedCanary && canShowResume(selectedCanary) && (
          <MenuItem onClick={() => handleAction('resume')}>
            <ResumeIcon sx={{ mr: 1 }} />
            Resume
          </MenuItem>
        )}
        {selectedCanary && canShowPause(selectedCanary) && (
          <MenuItem onClick={() => handleAction('pause')}>
            <PauseIcon sx={{ mr: 1 }} />
            Pause
          </MenuItem>
        )}
        {selectedCanary && canShowPromote(selectedCanary) && (
          <MenuItem onClick={() => handleAction('promote')}>
            <PromoteIcon sx={{ mr: 1 }} />
            Promote
          </MenuItem>
        )}
        {selectedCanary && canShowAbort(selectedCanary) && (
          <MenuItem onClick={() => handleAction('abort')}>
            <AbortIcon sx={{ mr: 1 }} />
            Abort
          </MenuItem>
        )}
      </Menu>
    </Box>
  )
}

export default CanaryList