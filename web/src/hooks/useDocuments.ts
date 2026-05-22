import { useCallback, useEffect, useRef, useState } from 'react'
import { listDocuments } from '../documents/api'
import { DocumentError, DocumentResponse } from '../documents/types'

type DocumentsState =
  | { phase: 'loading' }
  | { phase: 'loaded'; documents: DocumentResponse[] }
  | { phase: 'error'; message: string }

const EXPIRY_BUFFER_MS = 60_000

export function useDocuments() {
  const [state, setState] = useState<DocumentsState>({ phase: 'loading' })
  const refreshTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  const refresh = useCallback(async () => {
    if (refreshTimerRef.current) {
      clearTimeout(refreshTimerRef.current)
      refreshTimerRef.current = null
    }

    setState({ phase: 'loading' })
    try {
      const docs = await listDocuments()
      setState({ phase: 'loaded', documents: docs })

      if (docs.length > 0) {
        const earliestExpiry = Math.min(
          ...docs.map((d) => new Date(d.presigned_url_expires_at).getTime()),
        )
        const refreshAt = earliestExpiry - Date.now() - EXPIRY_BUFFER_MS
        if (refreshAt > 0) {
          refreshTimerRef.current = setTimeout(refresh, refreshAt)
        } else {
          // Already expired or expiring imminently — refresh immediately.
          void refresh()
        }
      }
    } catch (err) {
      const message = err instanceof DocumentError ? err.message : 'Failed to load documents'
      setState({ phase: 'error', message })
    }
  }, [])

  useEffect(() => {
    refresh()
    return () => {
      if (refreshTimerRef.current) clearTimeout(refreshTimerRef.current)
    }
  }, [refresh])

  return { state, refresh }
}
