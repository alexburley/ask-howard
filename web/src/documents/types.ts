export type DocumentSetStatus = 'UPLOADING' | 'PROCESSING' | 'READY' | 'FAILED'

export type DocumentSet = {
  id: string
  status: DocumentSetStatus
  original_filename: string
  error?: string
}

export type UploadSlot = {
  document_set_id: string
  presigned_url: string
  object_key: string
}

export type DocumentResponse = {
  id: string
  filename: string
  content_type: string
  size_bytes: number
  presigned_url: string
}

export type DocumentErrorCode = 'UNAUTHORIZED' | 'NOT_FOUND' | 'NETWORK_ERROR'

export class DocumentError extends Error {
  constructor(
    public readonly code: DocumentErrorCode,
    message: string,
  ) {
    super(message)
    this.name = 'DocumentError'
  }
}
