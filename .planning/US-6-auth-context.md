# US-6 — Auth Context

## User Story

**As a** frontend developer,  
**I want** a central auth context that tracks the current user and exposes login/logout/register actions,  
**So that** all components can read auth state and trigger auth flows without duplicating fetch logic.

## Acceptance Criteria

1. `AuthProvider` wraps the app and calls `GET /api/auth/me` on mount to determine the initial session state.
2. While the initial check is in-flight, `isLoading` is `true` — callers can gate rendering on this.
3. If the session is valid, `user` is set to `{ id: string; email: string }`.
4. If no session, `user` is `null`.
5. `login(email, password)` calls `POST /api/auth/login` and updates `user` on success; throws a typed `AuthError` on 401 or network failure.
6. `register(email, password)` calls `POST /api/auth/register`; throws a typed `AuthError` with a `code` field (`EMAIL_TAKEN | INVALID_EMAIL | PASSWORD_TOO_SHORT`) on 4xx.
7. `logout()` calls `POST /api/auth/logout` and clears `user`.
8. `useAuth()` hook throws if called outside `AuthProvider`.

## Out of Scope

- Token refresh, session expiry handling.
- Anything UI — this story is purely data/logic.

---

## Implementation Plan

### 1. Types — `web/src/auth/types.ts`

```ts
export type AuthUser = { id: string; email: string }

export type AuthErrorCode =
  | 'INVALID_CREDENTIALS'
  | 'EMAIL_TAKEN'
  | 'INVALID_EMAIL'
  | 'PASSWORD_TOO_SHORT'
  | 'NETWORK_ERROR'

export class AuthError extends Error {
  constructor(public code: AuthErrorCode, message: string) {
    super(message)
  }
}
```

### 2. API client — `web/src/auth/api.ts`

Thin fetch wrappers; each throws `AuthError` on failure.

- `fetchMe(): Promise<AuthUser | null>` — GET /api/auth/me; returns null on 401
- `apiLogin(email, password): Promise<AuthUser>` — POST /api/auth/login; throws on 401
- `apiRegister(email, password): Promise<AuthUser>` — POST /api/auth/register; maps 409/422 to error codes
- `apiLogout(): Promise<void>` — POST /api/auth/logout

### 3. Context — `web/src/auth/AuthContext.tsx`

```tsx
type AuthContextValue = {
  user: AuthUser | null
  isLoading: boolean
  login(email: string, password: string): Promise<void>
  register(email: string, password: string): Promise<void>
  logout(): Promise<void>
}

export const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) { ... }

export function useAuth(): AuthContextValue { ... }  // throws if no provider
```

`AuthProvider` calls `fetchMe()` in a `useEffect` on mount; sets `isLoading = false` once resolved.

## Files Created

| File | Purpose |
|------|---------|
| `web/src/auth/types.ts` | `AuthUser`, `AuthError`, `AuthErrorCode` |
| `web/src/auth/api.ts` | Raw fetch functions |
| `web/src/auth/AuthContext.tsx` | Provider + `useAuth` hook |
