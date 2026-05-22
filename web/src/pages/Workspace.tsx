import { useAuth } from '../auth/AuthContext'
import { UploadControl } from '../components/UploadControl'
import { useDocuments } from '../hooks/useDocuments'

export function Workspace() {
  const { logout } = useAuth()
  const { state: docsState, refresh } = useDocuments()

  return (
    <div className="workspace-inner">
      <button className="sign-out" onClick={logout}>Sign out</button>
      <UploadControl onReady={refresh} />
      {docsState.phase === 'loaded' && docsState.documents.length > 0 && (
        <p className="doc-count">{docsState.documents.length} document{docsState.documents.length !== 1 ? 's' : ''} in your workspace</p>
      )}
    </div>
  )
}
