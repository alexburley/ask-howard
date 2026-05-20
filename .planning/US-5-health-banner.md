# US-5 — API Health Banner

## User Story

**As a** user visiting the Ask Howard web app,  
**I want** to see a visible banner when the API is unavailable,  
**So that** I know the service is degraded and it is not my mistake.

## Acceptance Criteria

1. On initial page load the app silently polls `GET /api/health`.
2. If the endpoint returns anything other than `200 OK`, a full-width banner appears at the top of the viewport reading **"Service is currently unavailable. Please try again later."**
3. The banner is visually distinct (e.g. warning/error colour) and does not obstruct the rest of the UI (page scrolls beneath it).
4. The app re-polls every 30 seconds. If the health check recovers, the banner disappears automatically.
5. No user action is required to dismiss the banner — it self-heals.
6. While the first check is still in-flight the banner is not shown (avoid flash on fast connections).

## Out of Scope

- Toasts, modals, or manual dismiss buttons.
- Per-service breakdown (DB vs. API).
- Retry-with-backoff (fixed 30 s interval is sufficient for MVP).

---

## Implementation Plan

### 1. `useApiHealth` hook — `web/src/hooks/useApiHealth.ts`

- Uses `useState<boolean | null>` — `null` = still loading, `true` = healthy, `false` = unhealthy.
- On mount: fetch `GET /api/health`. On 200 set `true`, on any error/non-200 set `false`.
- `setInterval(poll, 30_000)` — clears on unmount via `useEffect` cleanup.
- No external libraries required — plain `fetch` is sufficient.

```ts
// shape
function useApiHealth(): boolean | null
```

### 2. `HealthBanner` component — `web/src/components/HealthBanner.tsx`

- Receives no props; calls `useApiHealth()` internally.
- Returns `null` while loading (`null` state) — prevents flash.
- Returns a `<div role="alert">` banner with a warning message when unhealthy.
- Returns `null` when healthy.

```tsx
// rough structure
export function HealthBanner() {
  const healthy = useApiHealth()
  if (healthy !== false) return null   // null (loading) or true (ok) → hide
  return (
    <div role="alert" className="health-banner">
      Service is currently unavailable. Please try again later.
    </div>
  )
}
```

### 3. Styles — `web/src/index.css`

Add `.health-banner` rule:
- `position: sticky; top: 0; z-index: 100` so it pins to the viewport top.
- Background: a muted red/amber; white text; `padding: 0.75rem 1rem; text-align: center`.

### 4. Wire into `App.tsx`

```tsx
import { HealthBanner } from './components/HealthBanner'

function App() {
  return (
    <div className="workspace">
      <HealthBanner />
      <p>Ask Howard</p>
    </div>
  )
}
```

### 5. Verify manually

- `make start` → visit `http://localhost:5173`.
- Stop the Postgres container (`docker compose stop postgres`) so `GET /api/health` returns 503 → banner appears.
- Restart Postgres → banner disappears within one poll cycle.

---

## Files Changed

| File | Action |
|------|--------|
| `web/src/hooks/useApiHealth.ts` | Create |
| `web/src/components/HealthBanner.tsx` | Create |
| `web/src/App.tsx` | Update — render `<HealthBanner />` |
| `web/src/index.css` | Update — add `.health-banner` styles |

## No Backend Changes Required

`GET /api/health` already exists and returns the correct shape.
