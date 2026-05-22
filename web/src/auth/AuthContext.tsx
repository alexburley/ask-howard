import { createContext, ReactNode, useContext, useEffect, useState } from 'react'
import { apiLogin, apiLogout, apiRegister, fetchMe } from './api'
import { AuthError, AuthUser } from './types'

type AuthContextValue = {
  user: AuthUser | null
  isLoading: boolean
  sessionError: string | null
  login(email: string, password: string): Promise<void>
  register(email: string, password: string): Promise<void>
  logout(): Promise<void>
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<AuthUser | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [sessionError, setSessionError] = useState<string | null>(null)

  useEffect(() => {
    fetchMe()
      .then(setUser)
      .catch((err) => {
        setSessionError(err instanceof AuthError ? err.message : 'Unable to connect to the server')
      })
      .finally(() => setIsLoading(false))
  }, [])

  async function login(email: string, password: string) {
    const u = await apiLogin(email, password)
    setUser(u)
  }

  async function register(email: string, password: string) {
    const u = await apiRegister(email, password)
    setUser(u)
  }

  async function logout() {
    await apiLogout()
    setUser(null)
  }

  return (
    <AuthContext.Provider value={{ user, isLoading, sessionError, login, register, logout }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}
