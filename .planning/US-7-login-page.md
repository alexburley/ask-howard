# US-7 — Login Page

## User Story

**As a** returning user,  
**I want** to enter my email and password to sign in,  
**So that** I can access the application.

## Acceptance Criteria

1. A login form is shown when no session exists (determined by auth context).
2. Form has `email` and `password` fields with a Submit button.
3. On submit, calls `login()` from auth context; on success the app transitions to the authenticated view.
4. If credentials are wrong (401), an inline error message reads: **"Invalid email or password."**
5. While the request is in-flight the submit button is disabled.
6. A link/button to switch to the Register page is present.

## Out of Scope

- Client-side field validation (server returns clear errors; keep the form simple).
- "Forgot password" flow.

---

## Implementation Plan

### `web/src/pages/LoginPage.tsx`

```tsx
export function LoginPage({ onSwitchToRegister }: { onSwitchToRegister: () => void }) {
  const { login } = useAuth()
  const [error, setError] = useState<string | null>(null)
  const [pending, setPending] = useState(false)

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setPending(true)
    setError(null)
    try {
      await login(email, password)
    } catch (err) {
      if (err instanceof AuthError) setError('Invalid email or password.')
      else setError('Something went wrong. Please try again.')
    } finally {
      setPending(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="auth-form">
      <h1>Sign in</h1>
      {error && <p className="auth-error">{error}</p>}
      <input type="email" ... />
      <input type="password" ... />
      <button type="submit" disabled={pending}>Sign in</button>
      <button type="button" onClick={onSwitchToRegister}>Create an account</button>
    </form>
  )
}
```

### CSS additions to `index.css`

- `.auth-form` — centered card, max-width ~360 px, gap between fields
- `.auth-error` — red text, small, above the submit button
- `input` — full-width, dark-themed, consistent padding
- `button[type=submit]` — primary style; `disabled` state muted

## Files Changed

| File | Action |
|------|--------|
| `web/src/pages/LoginPage.tsx` | Create |
| `web/src/index.css` | Update — auth form styles |
