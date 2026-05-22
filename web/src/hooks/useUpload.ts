import { useCallback, useRef, useState } from 'react'
import { completeUpload, pollDocumentSet, requestUploadSlot, uploadToPresignedUrl } from '../documents/api'
import { DocumentError, DocumentSet } from '../documents/types'

type UploadState =
  | { phase: 'idle' }
  | { phase: 'uploading'; progress: number }
  | { phase: 'processing'; setID: string }
  | { phase: 'done'; set: DocumentSet }
  | { phase: 'error'; message: string }

const POLL_INTERVAL_MS = 2000
const POLL_TIMEOUT_MS = 5 * 60 * 1000

export function useUpload(onReady?: () => void) {
  const [state, setState] = useState<UploadState>({ phase: 'idle' })
  const cancelledRef = useRef(false)

  const upload = useCallback(async (file: File) => {
    cancelledRef.current = false
    setState({ phase: 'uploading', progress: 0 })
    try {
      const slot = await requestUploadSlot(file.name)

      await uploadToPresignedUrl(slot.presignedUrl, file, (pct) => {
        if (!cancelledRef.current) setState({ phase: 'uploading', progress: pct })
      })

      const set = await completeUpload(slot.documentSetId)
      if (!cancelledRef.current) setState({ phase: 'processing', setID: set.id })

      await pollUntilDone(set.id)
    } catch (err) {
      if (!cancelledRef.current) {
        const message = err instanceof DocumentError ? err.message : 'Upload failed'
        setState({ phase: 'error', message })
      }
    }
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  async function pollUntilDone(setID: string) {
    const deadline = Date.now() + POLL_TIMEOUT_MS
    while (Date.now() < deadline) {
      if (cancelledRef.current) return
      await sleep(POLL_INTERVAL_MS)
      if (cancelledRef.current) return
      const set = await pollDocumentSet(setID)
      if (set.status === 'READY') {
        if (!cancelledRef.current) {
          setState({ phase: 'done', set })
          onReady?.()
        }
        return
      }
      if (set.status === 'FAILED') {
        if (!cancelledRef.current) {
          setState({ phase: 'error', message: set.error ?? 'Processing failed' })
        }
        return
      }
    }
    if (!cancelledRef.current) {
      setState({ phase: 'error', message: 'Processing timed out' })
    }
  }

  function reset() {
    cancelledRef.current = true
    setState({ phase: 'idle' })
  }

  return { state, upload, reset }
}

function sleep(ms: number) {
  return new Promise((resolve) => setTimeout(resolve, ms))
}
