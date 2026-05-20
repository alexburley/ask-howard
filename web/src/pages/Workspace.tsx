import { useAuth } from '../auth/AuthContext'

export function Workspace() {
  const { logout } = useAuth()

  return (
    <div className="workspace-inner">
      <button className="sign-out" onClick={logout}>Sign out</button>
      <p>Ask Howard</p>
    </div>
  )
}
