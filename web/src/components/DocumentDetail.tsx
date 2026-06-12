import { useEffect, useState } from 'react'
import { getDocument } from '../documents/api'
import { DocumentError, DocumentResponse } from '../documents/types'

type Props = {
  docId: string
  onClose: () => void
}

function formatBytes(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' })
}

function Preview({ doc }: { doc: DocumentResponse }) {
  if (doc.contentType.startsWith('image/')) {
    return <img src={doc.presignedUrl} alt={doc.filename} className="detail-preview-image" />
  }
  if (doc.contentType === 'application/pdf') {
    return <iframe src={doc.presignedUrl} title={doc.filename} className="detail-preview-pdf" />
  }
  return (
    <div className="detail-preview-other">
      <p className="detail-preview-other-label">No preview available</p>
    </div>
  )
}

// TODO: Epic 4 — AI tagging will populate this component
function DocumentTags() {
  return null
}

export function DocumentDetail({ docId, onClose }: Props) {
  const [doc, setDoc] = useState<DocumentResponse | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    setDoc(null)
    setError(null)
    getDocument(docId).then(setDoc).catch((err) => {
      setError(err instanceof DocumentError ? err.message : 'Failed to load document')
    })
  }, [docId])

  useEffect(() => {
    function onKeyDown(e: KeyboardEvent) {
      if (e.key === 'Escape') onClose()
    }
    window.addEventListener('keydown', onKeyDown)
    return () => window.removeEventListener('keydown', onKeyDown)
  }, [onClose])

  return (
    <aside className="detail-panel">
      <div className="detail-header">
        <h2 className="detail-title" title={doc?.filename ?? ''}>{doc?.filename ?? '…'}</h2>
        <button className="detail-close" onClick={onClose} aria-label="Close">✕</button>
      </div>

      {error && <p className="detail-error">{error}</p>}

      {doc && (
        <>
          <div className="detail-preview">
            <Preview doc={doc} />
          </div>

          <div className="detail-meta">
            <div className="detail-meta-row">
              <span className="detail-meta-label">Type</span>
              <span className="detail-meta-value">{doc.contentType}</span>
            </div>
            <div className="detail-meta-row">
              <span className="detail-meta-label">Size</span>
              <span className="detail-meta-value">{formatBytes(doc.sizeBytes)}</span>
            </div>
            <div className="detail-meta-row">
              <span className="detail-meta-label">Uploaded</span>
              <span className="detail-meta-value">{formatDate(doc.createdAt)}</span>
            </div>
          </div>

          <DocumentTags />

          <a href={doc.presignedUrl} download={doc.filename} className="detail-download">
            Download
          </a>
        </>
      )}
    </aside>
  )
}
