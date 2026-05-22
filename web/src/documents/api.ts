import { DocumentError, DocumentResponse, DocumentSet, UploadSlot } from './types'

type WireDocumentSet = {
  id: string
  status: string
  original_filename: string
  error?: string
}

type WireUploadSlot = {
  document_set_id: string
  presigned_url: string
  object_key: string
}

type WireDocument = {
  id: string
  filename: string
  content_type: string
  size_bytes: number
  presigned_url: string
  presigned_url_expires_at: string
}

function toDocumentSet(w: WireDocumentSet): DocumentSet {
  return {
    id: w.id,
    status: w.status as DocumentSet['status'],
    originalFilename: w.original_filename,
    error: w.error,
  }
}

function toUploadSlot(w: WireUploadSlot): UploadSlot {
  return {
    documentSetId: w.document_set_id,
    presignedUrl: w.presigned_url,
    objectKey: w.object_key,
  }
}

function toDocument(w: WireDocument): DocumentResponse {
  return {
    id: w.id,
    filename: w.filename,
    contentType: w.content_type,
    sizeBytes: w.size_bytes,
    presignedUrl: w.presigned_url,
    presignedUrlExpiresAt: w.presigned_url_expires_at,
  }
}

export async function requestUploadSlot(filename: string): Promise<UploadSlot> {
  let res: Response
  try {
    res = await fetch('/api/documents/upload', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ filename }),
    })
  } catch {
    throw new DocumentError('NETWORK_ERROR', 'Network error')
  }

  if (res.status === 401) throw new DocumentError('UNAUTHORIZED', 'Not authenticated')
  if (!res.ok) throw new DocumentError('NETWORK_ERROR', 'Failed to request upload slot')
  return toUploadSlot(await res.json())
}

export async function uploadToPresignedUrl(
  url: string,
  file: File,
  onProgress?: (percent: number) => void,
): Promise<void> {
  return new Promise((resolve, reject) => {
    const xhr = new XMLHttpRequest()
    xhr.open('PUT', url)
    xhr.setRequestHeader('Content-Type', 'application/zip')

    if (onProgress) {
      xhr.upload.onprogress = (e) => {
        if (e.lengthComputable) onProgress(Math.round((e.loaded / e.total) * 100))
      }
    }

    xhr.onload = () => {
      if (xhr.status >= 200 && xhr.status < 300) resolve()
      else reject(new DocumentError('NETWORK_ERROR', `Upload failed: ${xhr.status}`))
    }
    xhr.onerror = () => reject(new DocumentError('NETWORK_ERROR', 'Upload network error'))
    xhr.send(file)
  })
}

export async function completeUpload(setID: string): Promise<DocumentSet> {
  let res: Response
  try {
    res = await fetch(`/api/documents/sets/${setID}/complete`, { method: 'POST' })
  } catch {
    throw new DocumentError('NETWORK_ERROR', 'Network error')
  }

  if (res.status === 401) throw new DocumentError('UNAUTHORIZED', 'Not authenticated')
  if (res.status === 404) throw new DocumentError('NOT_FOUND', 'Document set not found')
  if (!res.ok) throw new DocumentError('NETWORK_ERROR', 'Failed to complete upload')
  return toDocumentSet(await res.json())
}

export async function listDocuments(): Promise<DocumentResponse[]> {
  let res: Response
  try {
    res = await fetch('/api/documents')
  } catch {
    throw new DocumentError('NETWORK_ERROR', 'Network error')
  }

  if (res.status === 401) throw new DocumentError('UNAUTHORIZED', 'Not authenticated')
  if (!res.ok) throw new DocumentError('NETWORK_ERROR', 'Failed to fetch documents')
  const wire: WireDocument[] = await res.json()
  return wire.map(toDocument)
}

export async function pollDocumentSet(setID: string): Promise<DocumentSet> {
  let res: Response
  try {
    res = await fetch(`/api/documents/sets/${setID}`)
  } catch {
    throw new DocumentError('NETWORK_ERROR', 'Network error')
  }

  if (res.status === 401) throw new DocumentError('UNAUTHORIZED', 'Not authenticated')
  if (res.status === 404) throw new DocumentError('NOT_FOUND', 'Document set not found')
  if (!res.ok) throw new DocumentError('NETWORK_ERROR', 'Failed to fetch document set')
  return toDocumentSet(await res.json())
}
