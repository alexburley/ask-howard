import { useCallback, useMemo } from 'react'
import { ReactFlow, Background, Controls, type Node, type NodeMouseHandler } from '@xyflow/react'
import '@xyflow/react/dist/style.css'
import { DocumentResponse } from '../documents/types'

type DocumentNodeData = {
  doc: DocumentResponse
}

type Props = {
  documents: DocumentResponse[]
  onNodeClick?: (doc: DocumentResponse) => void
}

function scatter(index: number): { x: number; y: number } {
  const cols = 4
  const col = index % cols
  const row = Math.floor(index / cols)
  const jitterX = (Math.random() - 0.5) * 40
  const jitterY = (Math.random() - 0.5) * 40
  return { x: col * 220 + jitterX, y: row * 180 + jitterY }
}

function isImage(contentType: string) {
  return contentType.startsWith('image/')
}

function fileIcon(contentType: string): string {
  if (contentType === 'application/pdf') return '📄'
  if (contentType.startsWith('text/')) return '📝'
  return '📁'
}

function DocumentNode({ data }: { data: DocumentNodeData }) {
  const { doc } = data
  return (
    <div className="doc-node">
      {isImage(doc.content_type) ? (
        <img src={doc.presigned_url} alt={doc.filename} className="doc-node-thumb" />
      ) : (
        <div className="doc-node-icon">{fileIcon(doc.content_type)}</div>
      )}
      <p className="doc-node-name">{doc.filename}</p>
    </div>
  )
}

const nodeTypes = { document: DocumentNode }

export function DocumentCanvas({ documents, onNodeClick }: Props) {
  const nodes: Node<DocumentNodeData>[] = useMemo(
    () =>
      documents.map((doc, i) => {
        const pos = scatter(i)
        return {
          id: doc.id,
          type: 'document',
          position: pos,
          data: { doc },
          draggable: true,
        }
      }),
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [documents.map((d) => d.id).join(',')],
  )

  const handleNodeClick: NodeMouseHandler<Node<DocumentNodeData>> = useCallback(
    (_event, node) => {
      onNodeClick?.(node.data.doc)
    },
    [onNodeClick],
  )

  return (
    <div className="canvas-container">
      <ReactFlow
        nodes={nodes}
        edges={[]}
        nodeTypes={nodeTypes}
        onNodeClick={handleNodeClick}
        fitView
        fitViewOptions={{ padding: 0.2 }}
        minZoom={0.2}
        maxZoom={3}
      >
        <Background gap={24} color="#2a2a2a" />
        <Controls />
      </ReactFlow>
    </div>
  )
}
