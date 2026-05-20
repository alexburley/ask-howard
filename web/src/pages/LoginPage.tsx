import { FormEvent, useState } from 'react'
import { useAuth } from '../auth/AuthContext'
import { AuthError } from '../auth/types'

type Props = { onSwitchToRegister: () => void }

export function LoginPage({ onSwitchToRegister }: Props) {
  const { login } = useAuth()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [pending, setPending] = useState(false)

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setPending(true)
    setError(null)
    try {
      await login(email, password)
    } catch (err) {
      setError(err instanceof AuthError ? 'Invalid email or password.' : 'Something went wrong. Please try again.')
    } finally {
      setPending(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="auth-form">
      <h1>Sign in</h1>
      {error && <p className="auth-error">{error}</p>}
      <input
        type="email"
        placeholder="Email"
        autoComplete="email"
        value={email}
        onChange={e => setEmail(e.target.value)}
        required
      />
      <input
        type="password"
        placeholder="Password"
        autoComplete="current-password"
        value={password}
        onChange={e => setPassword(e.target.value)}
        required
      />
      <button type="submit" disabled={pending}>
        {pending ? 'Signing in…' : 'Sign in'}
      </button>
      <button type="button" className="auth-switch" onClick={onSwitchToRegister}>
        Create an account
      </button>
    </form>
  )
}
