import { useEffect, useState } from 'react'
import { api } from '../api'
import type { FileNode, Upload } from '../types'

const renderNode = (node: FileNode, depth = 0) => (
  <li key={node.id} style={{ marginLeft: depth * 16 }}>
    <strong>{node.filename}</strong> <span className="muted">[{node.kind}]</span>
    {node.children?.length ? (
      <ul>
        {node.children.map((child) => renderNode(child, depth + 1))}
      </ul>
    ) : null}
  </li>
)

export default function FilesPage() {
  const [uploads, setUploads] = useState<Upload[]>([])
  const [selectedUpload, setSelectedUpload] = useState('')
  const [tree, setTree] = useState<FileNode[]>([])

  useEffect(() => {
    const load = async () => {
      setUploads(await api.listUploads('?limit=100'))
    }
    load()
  }, [])

  useEffect(() => {
    if (!selectedUpload) return
    const loadTree = async () => {
      setTree(await api.fileTree(selectedUpload))
    }
    loadTree()
  }, [selectedUpload])

  return (
    <section className="card">
      <h2>File hierarchy</h2>
      <label>
        Upload
        <select value={selectedUpload} onChange={(e) => setSelectedUpload(e.target.value)}>
          <option value="">Choose upload</option>
          {uploads.map((u) => (
            <option value={u.id} key={u.id}>
              {u.filename}
            </option>
          ))}
        </select>
      </label>
      {selectedUpload ? (
        <ul>{tree.map((node) => renderNode(node))}</ul>
      ) : (
        <p>Select an upload first.</p>
      )}
    </section>
  )
}
