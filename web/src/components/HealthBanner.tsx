import { useApiHealth } from '../hooks/useApiHealth'

export function HealthBanner() {
  const healthy = useApiHealth()

  if (healthy !== false) return null

  return (
    <div role="alert" className="health-banner">
      Service is currently unavailable. Please try again later.
    </div>
  )
}
