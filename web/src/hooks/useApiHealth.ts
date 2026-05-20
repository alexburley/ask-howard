import { useEffect, useState } from 'react'

const POLL_INTERVAL_MS = 30_000

async function checkHealth(): Promise<boolean> {
  try {
    const res = await fetch('/api/health')
    return res.ok
  } catch {
    return false
  }
}

export function useApiHealth(): boolean | null {
  const [healthy, setHealthy] = useState<boolean | null>(null)

  useEffect(() => {
    checkHealth().then(setHealthy)

    const id = setInterval(() => {
      checkHealth().then(setHealthy)
    }, POLL_INTERVAL_MS)

    return () => clearInterval(id)
  }, [])

  return healthy
}
