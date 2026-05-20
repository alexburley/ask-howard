export type AuthUser = { id: string; email: string }

export type AuthErrorCode =
  | 'INVALID_CREDENTIALS'
  | 'EMAIL_TAKEN'
  | 'INVALID_EMAIL'
  | 'PASSWORD_TOO_SHORT'
  | 'NETWORK_ERROR'

export class AuthError extends Error {
  constructor(
    public readonly code: AuthErrorCode,
    message: string,
  ) {
    super(message)
    this.name = 'AuthError'
  }
}
