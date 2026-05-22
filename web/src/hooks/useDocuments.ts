import { useCallback, useEffect, useState } from 'react'
import { listDocuments } from '../documents/api'
import { DocumentError, DocumentResponse } from '../documents/types'

type DocumentsState =
  | { phase: 'loading' }
  | { phase: 'loaded'; documents: DocumentResponse[] }
  | { phase: 'error'; message: string }

export function useDocuments() {
  const [state, setState] = useState<DocumentsState>({ phase: 'loading' })

  const refresh = useCallback(async () => {
    setState({ phase: 'loading' })
    try {
      const docs = await listDocuments()
      setState({ phase: 'loaded', documents: docs })
    } catch (err) {
      const message = err instanceof DocumentError ? err.message : 'Failed to load documents'
      setState({ phase: 'error', message })
    }
  }, [])

  useEffect(() => {
    refresh()
  }, [refresh])

  return { state, refresh }
}
