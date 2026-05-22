import { useState } from 'react'
import { completeUpload, requestUploadSlot, uploadToPresignedUrl } from '../documents/api'
import { DocumentError, DocumentSet } from '../documents/types'

type UploadState =
  | { phase: 'idle' }
  | { phase: 'uploading'; progress: number }
  | { phase: 'processing' }
  | { phase: 'done'; set: DocumentSet }
  | { phase: 'error'; message: string }

export function useUpload() {
  const [state, setState] = useState<UploadState>({ phase: 'idle' })

  async function upload(file: File) {
    setState({ phase: 'uploading', progress: 0 })
    try {
      const slot = await requestUploadSlot(file.name)

      await uploadToPresignedUrl(slot.presigned_url, file, (pct) => {
        setState({ phase: 'uploading', progress: pct })
      })

      setState({ phase: 'processing' })
      const set = await completeUpload(slot.document_set_id)
      setState({ phase: 'done', set })
    } catch (err) {
      const message = err instanceof DocumentError ? err.message : 'Upload failed'
      setState({ phase: 'error', message })
    }
  }

  function reset() {
    setState({ phase: 'idle' })
  }

  return { state, upload, reset }
}
