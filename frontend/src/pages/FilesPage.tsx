import { useCallback, useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import {
  AlertCircle,
  ChevronDown,
  ChevronRight,
  ExternalLink,
  Folder,
  FolderOpen,
  FolderTree,
  RefreshCw,
  Search,
  UploadCloud,
} from '../components/Icon'
import { Card } from '../components/Card'
import { Button } from '../components/Button'
import { Skeleton } from '../components/Skeleton'
import { EmptyState } from '../components/EmptyState'
import { StatusBadge } from '../components/StatusBadge'
import { CopyButton } from '../components/CopyButton'
import { FileTypeIcon } from '../components/FileTypeIcon'
import { api } from '../api'
import type { FileNode, Upload } from '../types'

function formatBytes(n: number): string {
  if (!Number.isFinite(n) || n < 0) return '—'
  if (n === 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  const idx = Math.min(Math.floor(Math.log(n) / Math.log(1024)), units.length - 1)
  const value = n / Math.pow(1024, idx)
  return `${value.toFixed(value >= 100 || idx === 0 ? 0 : 1)} ${units[idx]}`
}

function formatRelative(iso: string | undefined | null): string {
  if (!iso) return '—'
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return '—'
  const diffMs = Date.now() - d.getTime()
  if (diffMs < 0) return 'just now'
  const seconds = Math.floor(diffMs / 1000)
  if (seconds < 60) return 'just now'
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m ago`
  if (seconds < 86400) return `${Math.floor(seconds / 3600)}h ago`
  if (seconds < 86400 * 30) return `${Math.floor(seconds / 86400)}d ago`
  return d.toLocaleDateString()
}

function truncate(s: string, n: number): string {
  if (!s) return ''
  return s.length > n ? `${s.slice(0, n - 1)}…` : s
}

function isDirectory(node: FileNode): boolean {
  if (node.kind === 'directory' || node.kind === 'folder' || node.kind === 'dir') return true
  if (node.mime_type) return false
  return (node.children?.length ?? 0) > 0
}

function filterTree(nodes: FileNode[], query: string): FileNode[] {
  if (!query) return nodes
  const q = query.toLowerCase()
  const result: FileNode[] = []
  for (const node of nodes) {
    const selfMatch = node.filename.toLowerCase().includes(q)
    const filteredChildren = node.children ? filterTree(node.children, query) : undefined
    if (selfMatch || (filteredChildren && filteredChildren.length > 0)) {
      result.push({ ...node, children: filteredChildren })
    }
  }
  return result
}

function countNodes(nodes: FileNode[]): number {
  let total = 0
  for (const node of nodes) {
    total += 1
    if (node.children) total += countNodes(node.children)
  }
  return total
}

interface TreeNodeProps {
  node: FileNode
  depth: number
  expanded: Set<string>
  selectedId: string | null
  onToggle: (id: string) => void
  onSelect: (node: FileNode) => void
}

function TreeNode({ node, depth, expanded, selectedId, onToggle, onSelect }: TreeNodeProps) {
  const isDir = isDirectory(node)
  const hasChildren = (node.children?.length ?? 0) > 0
  const isExpanded = expanded.has(node.id)
  const isSelected = selectedId === node.id

  return (
    <li>
      <div
        className={`tree-row${isSelected ? ' active' : ''}`}
        style={{ paddingLeft: `${depth * 20 + 8}px` }}
      >
        {hasChildren ? (
          <button
            type="button"
            className="tree-chevron"
            aria-label={isExpanded ? 'Collapse' : 'Expand'}
            aria-expanded={isExpanded}
            onClick={(e) => {
              e.stopPropagation()
              onToggle(node.id)
            }}
          >
            {isExpanded ? <ChevronDown size={14} /> : <ChevronRight size={14} />}
          </button>
        ) : (
          <span className="tree-chevron-spacer" aria-hidden="true" />
        )}
        <button
          type="button"
          className="tree-row-main"
          onClick={() => onSelect(node)}
          aria-current={isSelected ? 'true' : undefined}
        >
          {isDir ? (
            isExpanded && hasChildren ? (
              <FolderOpen size={16} className="muted" />
            ) : (
              <Folder size={16} className="muted" />
            )
          ) : (
            <FileTypeIcon filename={node.filename} mimeType={node.mime_type} size={16} />
          )}
          <span className="tree-row-label truncate">{node.filename}</span>
        </button>
      </div>
      {hasChildren && isExpanded ? (
        <ul className="tree-children">
          {(node.children ?? []).map((child) => (
            <TreeNode
              key={child.id}
              node={child}
              depth={depth + 1}
              expanded={expanded}
              selectedId={selectedId}
              onToggle={onToggle}
              onSelect={onSelect}
            />
          ))}
        </ul>
      ) : null}
    </li>
  )
}

export default function FilesPage() {
  const [uploads, setUploads] = useState<Upload[] | null>(null)
  const [uploadsError, setUploadsError] = useState<string | null>(null)
  const [uploadFilter, setUploadFilter] = useState('')
  const [selectedUploadId, setSelectedUploadId] = useState<string | null>(null)
  const [tree, setTree] = useState<FileNode[] | null>(null)
  const [treeLoading, setTreeLoading] = useState(false)
  const [treeError, setTreeError] = useState<string | null>(null)
  const [expanded, setExpanded] = useState<Set<string>>(new Set())
  const [selectedNode, setSelectedNode] = useState<FileNode | null>(null)
  const [treeFilter, setTreeFilter] = useState('')

  const reload = useCallback(async () => {
    setUploadsError(null)
    try {
      const data = await api.listUploads('?limit=100')
      setUploads(data)
    } catch (e) {
      const msg = e instanceof Error ? e.message : 'Failed to fetch uploads'
      setUploadsError(msg)
      setUploads([])
    }
  }, [])

  useEffect(() => {
    void reload()
  }, [reload])

  const filteredUploads = useMemo(() => {
    const list = uploads ?? []
    const q = uploadFilter.toLowerCase()
    if (!q) return list
    return list.filter(
      (u) =>
        u.filename.toLowerCase().includes(q) ||
        (u.original_name ?? '').toLowerCase().includes(q),
    )
  }, [uploads, uploadFilter])

  const selectedUpload = useMemo(
    () => (uploads ?? []).find((u) => u.id === selectedUploadId) ?? null,
    [uploads, selectedUploadId],
  )

  const loadTree = useCallback(async (uploadId: string) => {
    setTreeError(null)
    setTreeLoading(true)
    setTree(null)
    setSelectedNode(null)
    try {
      const data = await api.fileTree(uploadId)
      setTree(data)
      setExpanded(new Set(data.map((n) => n.id)))
    } catch (e) {
      const msg = e instanceof Error ? e.message : 'Failed to fetch file tree'
      setTreeError(msg)
      setTree([])
    } finally {
      setTreeLoading(false)
    }
  }, [])

  useEffect(() => {
    if (!selectedUploadId) {
      setTree(null)
      setTreeError(null)
      setExpanded(new Set())
      setSelectedNode(null)
      setTreeFilter('')
      return
    }
    setTreeFilter('')
    setSelectedNode(null)
    void loadTree(selectedUploadId)
  }, [selectedUploadId, loadTree])

  const filteredTree = useMemo(() => {
    if (!tree) return []
    return filterTree(tree, treeFilter)
  }, [tree, treeFilter])

  const totalNodes = useMemo(() => (tree ? countNodes(tree) : 0), [tree])
  const visibleNodes = useMemo(() => countNodes(filteredTree), [filteredTree])

  const toggleExpanded = useCallback((id: string) => {
    setExpanded((prev) => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }, [])

  const handleSelectUpload = useCallback((id: string) => {
    setSelectedUploadId(id)
  }, [])

  const reloadTree = useCallback(() => {
    if (selectedUploadId) void loadTree(selectedUploadId)
  }, [selectedUploadId, loadTree])

  const uploadsLoading = !uploads && !uploadsError

  return (
    <div className="stack-lg">
      <header>
        <h2>Files</h2>
        <p className="muted">Browse the parsed file tree from any upload.</p>
      </header>

      <div className="files-layout">
        <Card
          padding="none"
          title="Uploads"
          actions={
            <Button
              variant="ghost"
              size="sm"
              iconLeft={<RefreshCw size={14} />}
              onClick={() => void reload()}
            >
              Refresh
            </Button>
          }
        >
          <div className="upload-toolbar">
            <label className="upload-search">
              <span className="visually-hidden">Filter uploads</span>
              <Search size={14} className="muted" aria-hidden="true" />
              <input
                type="search"
                value={uploadFilter}
                onChange={(e) => setUploadFilter(e.target.value)}
                placeholder="Filter uploads…"
              />
            </label>
          </div>

          <div className="upload-list">
            {uploadsLoading ? (
              <div className="stack-sm">
                <Skeleton height={48} />
                <Skeleton height={48} />
                <Skeleton height={48} />
                <Skeleton height={48} />
                <Skeleton height={48} />
                <Skeleton height={48} />
              </div>
            ) : uploadsError ? (
              <EmptyState
                icon={AlertCircle}
                title="Couldn't load uploads"
                description={uploadsError}
                action={
                  <Button
                    onClick={() => void reload()}
                    variant="secondary"
                    iconLeft={<RefreshCw size={14} />}
                  >
                    Retry
                  </Button>
                }
              />
            ) : filteredUploads.length === 0 ? (
              uploads && uploads.length > 0 ? (
                <EmptyState
                  icon={Search}
                  title="No uploads match your filter"
                  description="Try clearing the filter to see all uploads."
                />
              ) : (
                <EmptyState
                  icon={UploadCloud}
                  title="No uploads yet"
                  description="Your parsed files will appear here once you upload."
                  action={
                    <Link to="/upload">
                      <Button variant="primary" iconLeft={<UploadCloud size={16} />}>
                        Upload a file
                      </Button>
                    </Link>
                  }
                />
              )
            ) : (
              filteredUploads.map((u) => (
                <button
                  key={u.id}
                  type="button"
                  className={`upload-row${selectedUploadId === u.id ? ' active' : ''}`}
                  onClick={() => handleSelectUpload(u.id)}
                  aria-current={selectedUploadId === u.id ? 'true' : undefined}
                >
                  <FileTypeIcon
                    filename={u.original_name ?? u.filename}
                    mimeType={u.content_type}
                    size={20}
                  />
                  <div className="stack-sm upload-row-text">
                    <span className="truncate" style={{ fontWeight: 600 }}>
                      {u.original_name || u.filename}
                    </span>
                    <span className="muted text-sm">
                      {formatBytes(u.size_bytes)} · {formatRelative(u.created_at)}
                    </span>
                  </div>
                  <StatusBadge status={u.status} kind="upload" />
                </button>
              ))
            )}
          </div>
        </Card>

        <Card padding="none">
          {!selectedUploadId ? (
            <EmptyState
              icon={FolderTree}
              title="Select an upload"
              description="Pick an upload from the left to see its file tree."
            />
          ) : treeLoading ? (
            <div className="card-pad-md">
              <div className="stack-sm">
                <Skeleton height={28} />
                <Skeleton height={28} />
                <Skeleton height={28} />
                <Skeleton height={28} />
                <Skeleton height={28} />
              </div>
            </div>
          ) : treeError ? (
            <div className="card-pad-md">
              <EmptyState
                icon={AlertCircle}
                title="Couldn't load tree"
                description={treeError}
                action={
                  <Button
                    onClick={reloadTree}
                    variant="secondary"
                    iconLeft={<RefreshCw size={14} />}
                  >
                    Retry
                  </Button>
                }
              />
            </div>
          ) : !tree || tree.length === 0 ? (
            <div className="card-pad-md">
              <EmptyState
                icon={FolderTree}
                title="No files in this upload"
                description="The upload didn't produce any parsed files."
              />
            </div>
          ) : (
            <>
              <div className="tree-toolbar">
                <label className="tree-search">
                  <span className="visually-hidden">Filter tree</span>
                  <Search size={14} className="muted" aria-hidden="true" />
                  <input
                    type="search"
                    value={treeFilter}
                    onChange={(e) => setTreeFilter(e.target.value)}
                    placeholder="Filter files…"
                  />
                </label>
                <span className="muted text-sm">
                  {visibleNodes} of {totalNodes}
                </span>
              </div>

              <div className="tree-scroll">
                {filteredTree.length === 0 ? (
                  <div className="card-pad-md">
                    <EmptyState
                      icon={Search}
                      title="No files match"
                      description="Try a different filter."
                    />
                  </div>
                ) : (
                  <ul className="tree-root">
                    {filteredTree.map((node) => (
                      <TreeNode
                        key={node.id}
                        node={node}
                        depth={0}
                        expanded={expanded}
                        selectedId={selectedNode?.id ?? null}
                        onToggle={toggleExpanded}
                        onSelect={setSelectedNode}
                      />
                    ))}
                  </ul>
                )}
              </div>

              {selectedNode ? (
                <div className="tree-details">
                  <h4 className="tree-details-title">Selected file</h4>
                  <div className="grid-2 tree-details-grid">
                    <div className="tree-detail-item">
                      <span className="muted text-sm">Filename</span>
                      <span className="truncate">{selectedNode.filename}</span>
                    </div>
                    <div className="tree-detail-item">
                      <span className="muted text-sm">Kind</span>
                      <span className="truncate">{selectedNode.kind || '—'}</span>
                    </div>
                    <div className="tree-detail-item">
                      <span className="muted text-sm">MIME</span>
                      <span className="truncate">{selectedNode.mime_type || '—'}</span>
                    </div>
                    <div className="tree-detail-item">
                      <span className="muted text-sm">Size</span>
                      <span>{formatBytes(selectedNode.size_bytes)}</span>
                    </div>
                    <div className="tree-detail-item">
                      <span className="muted text-sm">SHA-256</span>
                      <span className="sha-row">
                        <code className="font-mono text-sm truncate">
                          {selectedNode.sha256 || '—'}
                        </code>
                        {selectedNode.sha256 ? (
                          <CopyButton text={selectedNode.sha256} aria-label="Copy SHA-256" />
                        ) : null}
                      </span>
                    </div>
                    <div className="tree-detail-item">
                      <span className="muted text-sm">Path</span>
                      <code className="font-mono text-sm truncate">
                        {selectedNode.path || '—'}
                      </code>
                    </div>
                  </div>
                  {!isDirectory(selectedNode) ? (
                    <div className="tree-details-action">
                      <Link to="/search">
                        <Button
                          variant="secondary"
                          size="sm"
                          iconRight={<ExternalLink size={12} />}
                        >
                          Search for "{truncate(selectedNode.filename, 30)}"
                        </Button>
                      </Link>
                    </div>
                  ) : null}
                </div>
              ) : null}
            </>
          )}
        </Card>
      </div>
    </div>
  )
}
