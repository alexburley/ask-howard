import { useEffect, useState } from 'react'
import { useAuth } from './auth/AuthContext'
import { HealthBanner } from './components/HealthBanner'
import { LoginPage } from './pages/LoginPage'
import { RegisterPage } from './pages/RegisterPage'
import { Workspace } from './pages/Workspace'

function App() {
  const { user, isLoading } = useAuth()
  const [page, setPage] = useState<'login' | 'register'>('login')

  useEffect(() => {
    if (!isLoading && !user) setPage('login')
  }, [user, isLoading])

  if (isLoading) return null

  if (!user) {
    return (
      <div className="workspace">
        {page === 'login'
          ? <LoginPage onSwitchToRegister={() => setPage('register')} />
          : <RegisterPage onSwitchToLogin={() => setPage('login')} />}
      </div>
    )
  }

  return (
    <div className="workspace">
      <HealthBanner />
      <Workspace />
    </div>
  )
}

export default App
