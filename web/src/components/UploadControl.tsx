import { useRef } from 'react'
import { useUpload } from '../hooks/useUpload'

type Props = {
  onReady?: () => void
}

export function UploadControl({ onReady }: Props) {
  const { state, upload, reset } = useUpload(onReady)
  const inputRef = useRef<HTMLInputElement>(null)

  function handleFiles(files: FileList | null) {
    const file = files?.[0]
    if (!file) return
    upload(file)
  }

  function handleDrop(e: React.DragEvent) {
    e.preventDefault()
    handleFiles(e.dataTransfer.files)
  }

  function handleDragOver(e: React.DragEvent) {
    e.preventDefault()
  }

  if (state.phase === 'uploading') {
    return (
      <div className="upload-control">
        <p className="upload-status">Uploading… {state.progress}%</p>
        <div className="upload-progress-bar">
          <div className="upload-progress-fill" style={{ width: `${state.progress}%` }} />
        </div>
      </div>
    )
  }

  if (state.phase === 'processing') {
    return (
      <div className="upload-control">
        <p className="upload-status">Processing your documents…</p>
        <div className="upload-progress-bar">
          <div className="upload-progress-fill upload-progress-fill--indeterminate" />
        </div>
      </div>
    )
  }

  if (state.phase === 'done') {
    return (
      <div className="upload-control">
        <p className="upload-status upload-status--success">
          {state.set.documentCount} document{state.set.documentCount !== 1 ? 's' : ''} ready
        </p>
        <button className="upload-btn" onClick={reset}>Upload another</button>
      </div>
    )
  }

  if (state.phase === 'error') {
    return (
      <div className="upload-control">
        <p className="upload-status upload-status--error">{state.message}</p>
        <button className="upload-btn" onClick={reset}>Try again</button>
      </div>
    )
  }

  return (
    <div
      className="upload-control upload-dropzone"
      onDrop={handleDrop}
      onDragOver={handleDragOver}
      onClick={() => inputRef.current?.click()}
    >
      <input
        ref={inputRef}
        type="file"
        accept=".zip"
        className="upload-input-hidden"
        onChange={(e) => handleFiles(e.target.files)}
      />
      <p className="upload-dropzone-label">Drop a zip file here or click to browse</p>
    </div>
  )
}
