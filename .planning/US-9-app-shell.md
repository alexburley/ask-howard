# US-9 — App Shell & Auth Gating

## User Story

**As a** user,  
**I want** the app to automatically send me to the login page when I'm not authenticated and into the app when I am,  
**So that** I never see content I shouldn't and don't have to manually navigate.

## Acceptance Criteria

1. While the initial session check is in-flight (`isLoading = true`), nothing is rendered (blank screen — no flash of login or app content).
2. If unauthenticated, the Login page is shown. Switching to Register and back works without a page reload.
3. If authenticated, the main workspace is shown with a visible **Sign out** button.
4. Clicking Sign out calls `logout()`, clears the session, and returns the user to the Login page.
5. No client-side router is required — conditional rendering is sufficient at this scale.

---

## Implementation Plan

### Update `App.tsx`

```tsx
function App() {
  const { user, isLoading } = useAuth()
  const [page, setPage] = useState<'login' | 'register'>('login')

  if (isLoading) return null

  if (!user) {
    return page === 'login'
      ? <LoginPage onSwitchToRegister={() => setPage('register')} />
      : <RegisterPage onSwitchToLogin={() => setPage('login')} />
  }

  return (
    <div className="workspace">
      <HealthBanner />
      <Workspace />
    </div>
  )
}
```

### `web/src/pages/Workspace.tsx`

Extracts the authenticated shell — currently just `<p>Ask Howard</p>` plus a Sign out button.

```tsx
export function Workspace() {
  const { logout } = useAuth()
  return (
    <div className="workspace-inner">
      <button className="sign-out" onClick={logout}>Sign out</button>
      <p>Ask Howard</p>
    </div>
  )
}
```

### Update `main.tsx`

Wrap `<App />` with `<AuthProvider>`.

### CSS additions to `index.css`

- `.sign-out` — small, top-right corner, ghost/muted style

## Files Changed

| File | Action |
|------|--------|
| `web/src/App.tsx` | Update — auth gating logic |
| `web/src/pages/Workspace.tsx` | Create |
| `web/src/main.tsx` | Update — wrap with `AuthProvider` |
| `web/src/index.css` | Update — sign-out button style |
