export type DocumentSetStatus = 'UPLOADING' | 'PROCESSING' | 'READY' | 'FAILED'

export type DocumentSet = {
  id: string
  status: DocumentSetStatus
  originalFilename: string
  error?: string
}

export type UploadSlot = {
  documentSetId: string
  presignedUrl: string
  objectKey: string
}

export type DocumentResponse = {
  id: string
  filename: string
  contentType: string
  sizeBytes: number
  presignedUrl: string
  presignedUrlExpiresAt: string
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
