import { FormEvent, useState } from 'react'
import { useAuth } from '../auth/AuthContext'
import { AuthError, AuthErrorCode } from '../auth/types'

type Props = { onSwitchToLogin: () => void }

const errorMessages: Record<AuthErrorCode, string> = {
  EMAIL_TAKEN: 'An account with this email already exists.',
  INVALID_EMAIL: 'Please enter a valid email address.',
  PASSWORD_TOO_SHORT: 'Password must be at least 8 characters.',
  INVALID_CREDENTIALS: 'Something went wrong. Please try again.',
  NETWORK_ERROR: 'Something went wrong. Please try again.',
}

export function RegisterPage({ onSwitchToLogin }: Props) {
  const { register } = useAuth()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [pending, setPending] = useState(false)

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setPending(true)
    setError(null)
    try {
      await register(email, password)
    } catch (err) {
      setError(err instanceof AuthError ? errorMessages[err.code] : 'Something went wrong. Please try again.')
    } finally {
      setPending(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="auth-form">
      <h1>Create account</h1>
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
        placeholder="Password (min 8 characters)"
        autoComplete="new-password"
        value={password}
        onChange={e => setPassword(e.target.value)}
        required
      />
      <button type="submit" disabled={pending}>
        {pending ? 'Creating account…' : 'Create account'}
      </button>
      <button type="button" className="auth-switch" onClick={onSwitchToLogin}>
        Already have an account? Sign in
      </button>
    </form>
  )
}
