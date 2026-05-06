import { Routes, Route, Navigate } from 'react-router-dom'
import { useAuth } from './hooks/useAuth'
import Layout from './components/Layout'
import Login from './pages/Login'
import Setup from './pages/Setup'
import Dashboard from './pages/Dashboard'
import ConfigEditor from './pages/ConfigEditor'
import Zones from './pages/Zones'
import Blocklist from './pages/Blocklist'

function App() {
  const { isAuthenticated, isLoading, setupRequired, login, setup, logout } = useAuth()

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>
    )
  }

  if (setupRequired) {
    return <Setup onSetup={setup} />
  }

  if (!isAuthenticated) {
    return <Login onLogin={login} />
  }

  return (
    <Layout onLogout={logout}>
      <Routes>
        <Route path="/" element={<Dashboard />} />
        <Route path="/config" element={<ConfigEditor />} />
        <Route path="/zones" element={<Zones />} />
        <Route path="/blocklist" element={<Blocklist />} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </Layout>
  )
}

export default App
