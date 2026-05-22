import { useAuth } from '../auth/AuthContext'
import { UploadControl } from '../components/UploadControl'

export function Workspace() {
  const { logout } = useAuth()

  return (
    <div className="workspace-inner">
      <button className="sign-out" onClick={logout}>Sign out</button>
      <UploadControl />
    </div>
  )
}
