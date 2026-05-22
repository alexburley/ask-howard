import { DocumentError, DocumentResponse, DocumentSet, UploadSlot } from './types'

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
  return res.json()
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
  return res.json()
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
  return res.json()
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
  return res.json()
}
