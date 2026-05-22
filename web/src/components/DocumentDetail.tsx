import { DocumentResponse } from '../documents/types'

type Props = {
  doc: DocumentResponse
  onClose: () => void
}

function formatBytes(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

function Preview({ doc }: { doc: DocumentResponse }) {
  if (doc.content_type.startsWith('image/')) {
    return <img src={doc.presigned_url} alt={doc.filename} className="detail-preview-image" />
  }
  if (doc.content_type === 'application/pdf') {
    return (
      <iframe
        src={doc.presigned_url}
        title={doc.filename}
        className="detail-preview-pdf"
      />
    )
  }
  return (
    <div className="detail-preview-other">
      <p className="detail-preview-other-label">No preview available</p>
    </div>
  )
}

export function DocumentDetail({ doc, onClose }: Props) {
  return (
    <aside className="detail-panel">
      <div className="detail-header">
        <h2 className="detail-title" title={doc.filename}>{doc.filename}</h2>
        <button className="detail-close" onClick={onClose} aria-label="Close">✕</button>
      </div>

      <div className="detail-preview">
        <Preview doc={doc} />
      </div>

      <div className="detail-meta">
        <div className="detail-meta-row">
          <span className="detail-meta-label">Type</span>
          <span className="detail-meta-value">{doc.content_type}</span>
        </div>
        <div className="detail-meta-row">
          <span className="detail-meta-label">Size</span>
          <span className="detail-meta-value">{formatBytes(doc.size_bytes)}</span>
        </div>
      </div>

      {/* Tags seam — AI tagging will populate this area in a future epic */}
      <div className="detail-tags-seam" />

      <a
        href={doc.presigned_url}
        download={doc.filename}
        className="detail-download"
      >
        Download
      </a>
    </aside>
  )
}
