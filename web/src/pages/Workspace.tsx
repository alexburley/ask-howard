import { useState } from 'react'
import { useAuth } from '../auth/AuthContext'
import { DocumentCanvas } from '../components/DocumentCanvas'
import { UploadControl } from '../components/UploadControl'
import { useDocuments } from '../hooks/useDocuments'

export function Workspace() {
  const { logout } = useAuth()
  const { state: docsState, refresh } = useDocuments()
  const [uploading, setUploading] = useState(false)

  const hasDocs = docsState.phase === 'loaded' && docsState.documents.length > 0

  function handleReady() {
    setUploading(false)
    refresh()
  }

  if (!hasDocs || uploading) {
    return (
      <div className="workspace-inner">
        <button className="sign-out" onClick={logout}>Sign out</button>
        {hasDocs && (
          <button className="upload-btn canvas-back-btn" onClick={() => setUploading(false)}>
            ← Back to canvas
          </button>
        )}
        <UploadControl onReady={handleReady} />
      </div>
    )
  }

  return (
    <div className="workspace-canvas">
      <button className="sign-out" onClick={logout}>Sign out</button>
      <button className="upload-btn canvas-upload-btn" onClick={() => setUploading(true)}>
        Upload more
      </button>
      <DocumentCanvas documents={docsState.documents} />
    </div>
  )
}
