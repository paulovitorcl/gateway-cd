import { Routes, Route } from 'react-router-dom'
import { Box } from '@mui/material'
import Layout from './components/Layout'
import Dashboard from './pages/Dashboard'
import CanaryList from './pages/CanaryList'
import CanaryDetail from './pages/CanaryDetail'
import CreateCanary from './pages/CreateCanary'

function App() {
  return (
    <Box sx={{ display: 'flex', minHeight: '100vh' }}>
      <Layout>
        <Routes>
          <Route path="/" element={<Dashboard />} />
          <Route path="/canaries" element={<CanaryList />} />
          <Route path="/canaries/:namespace/:name" element={<CanaryDetail />} />
          <Route path="/canaries/new" element={<CreateCanary />} />
        </Routes>
      </Layout>
    </Box>
  )
}

export default App