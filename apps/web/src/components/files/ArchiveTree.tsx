import { Link } from 'react-router-dom'
import { useEffect, useMemo, useState } from 'react'
import { FileStatusBadge } from '@/components/files/FileStatusBadge'
import type { FileTreeNode } from '@/types/file'

type ArchiveTreeProps = {
  tree: FileTreeNode | null
  rootName?: string
}

type NodeLineProps = {
  node: FileTreeNode
  level: number
  expanded: Set<string>
  onToggle: (id: string) => void
}

const NodeLine = ({ node, level, expanded, onToggle }: NodeLineProps) => {
  const hasChildren = node.children.length > 0
  const isOpen = expanded.has(node.file.id)
  const indent = level * 16

  return (
    <li className="text-sm">
      <div className="flex items-center gap-2" style={{ marginLeft: indent }}>
        {hasChildren ? (
          <button
            className="h-5 w-5 rounded border border-slate-600/80 text-xs text-slate-300"
            onClick={() => onToggle(node.file.id)}
            aria-label={`${isOpen ? 'Collapse' : 'Expand'} ${node.file.original_name}`}
          >
            {isOpen ? '▾' : '▸'}
          </button>
        ) : (
          <span className="inline-block h-5 w-5" />
        )}

        <Link to={`/files/${node.file.id}`} className="text-blue-300 hover:underline">
          {node.file.original_name}
        </Link>
        <span className="text-xs text-slate-300">· {node.file.size_bytes} B</span>
        <FileStatusBadge status={node.file.processing_status} />
      </div>

      {hasChildren && isOpen ? (
        <ul className="mt-1 space-y-1">
          {node.children.map((child) => (
            <NodeLine key={child.file.id} node={child} level={level + 1} expanded={expanded} onToggle={onToggle} />
          ))}
        </ul>
      ) : null}
    </li>
  )
}

export const ArchiveTree = ({ tree, rootName }: ArchiveTreeProps) => {
  const [expanded, setExpanded] = useState<Set<string>>(new Set())

  const initialExpanded = useMemo(() => {
    if (!tree) return new Set<string>()
    return new Set([tree.file.id])
  }, [tree])

  if (!tree) {
    return <div className="text-sm text-slate-400">No tree data.</div>
  }

  useEffect(() => {
    setExpanded(initialExpanded)
  }, [initialExpanded])

  const toggle = (fileId: string) => {
    setExpanded((previous) => {
      const next = new Set(previous)
      if (next.has(fileId)) {
        next.delete(fileId)
      } else {
        next.add(fileId)
      }
      return next
    })
  }

  return (
    <div className="space-y-3">
      {rootName ? <div className="text-sm text-slate-300">Root: {rootName}</div> : null}
      <ul className="space-y-1">
        <NodeLine node={tree} level={0} expanded={expanded} onToggle={toggle} />
      </ul>
    </div>
  )
}
