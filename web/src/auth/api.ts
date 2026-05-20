import { AuthError, AuthUser } from './types'

type ApiViolation = { pointer: string; detail: string }
type ApiProblem = { status: number; violations?: ApiViolation[] }

async function parseProblem(res: Response): Promise<ApiProblem> {
  try {
    const body = await res.json()
    return { status: res.status, violations: body.violations }
  } catch {
    return { status: res.status }
  }
}

export async function fetchMe(): Promise<AuthUser | null> {
  try {
    const res = await fetch('/api/auth/me')
    if (res.status === 401) return null
    if (!res.ok) throw new AuthError('NETWORK_ERROR', 'Failed to fetch session')
    return res.json()
  } catch (err) {
    if (err instanceof AuthError) throw err
    return null
  }
}

export async function apiLogin(email: string, password: string): Promise<AuthUser> {
  let res: Response
  try {
    res = await fetch('/api/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password }),
    })
  } catch {
    throw new AuthError('NETWORK_ERROR', 'Network error')
  }

  if (res.status === 401) throw new AuthError('INVALID_CREDENTIALS', 'Invalid email or password')
  if (!res.ok) throw new AuthError('NETWORK_ERROR', 'Login failed')
  return res.json()
}

export async function apiRegister(email: string, password: string): Promise<AuthUser> {
  let res: Response
  try {
    res = await fetch('/api/auth/register', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password }),
    })
  } catch {
    throw new AuthError('NETWORK_ERROR', 'Network error')
  }

  if (res.ok) return res.json()

  const problem = await parseProblem(res)

  if (problem.status === 409) throw new AuthError('EMAIL_TAKEN', 'Email already registered')
  if (problem.status === 422) {
    const hasPasswordViolation = problem.violations?.some(v => v.pointer === '/password') ?? false
    if (hasPasswordViolation) throw new AuthError('PASSWORD_TOO_SHORT', 'Password too short')
    throw new AuthError('INVALID_EMAIL', 'Invalid email address')
  }

  throw new AuthError('NETWORK_ERROR', 'Registration failed')
}

export async function apiLogout(): Promise<void> {
  try {
    await fetch('/api/auth/logout', { method: 'POST' })
  } catch {
    // best-effort — clear local state regardless
  }
}
