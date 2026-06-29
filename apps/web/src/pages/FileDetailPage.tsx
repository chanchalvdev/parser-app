import { Link, useParams } from 'react-router-dom'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useMemo } from 'react'
import { useState } from 'react'
import { PageHeader } from '@/components/layout/PageHeader'
import { Card } from '@/components/ui/Card'
import { FileStatusBadge } from '@/components/files/FileStatusBadge'
import { FileTypeBadge } from '@/components/files/FileTypeBadge'
import { ArchiveTree } from '@/components/files/ArchiveTree'
import { PasswordRequiredModal } from '@/components/files/PasswordRequiredModal'
import { FileTable } from '@/components/files/FileTable'
import { ParsedRecordsPreview } from '@/components/files/ParsedRecordsPreview'
import { useFile, useFileTree } from '@/hooks/apiHooks'
import { getFileChildren, getFileRecords } from '@/services/filesApi'
import { parseTime, formatBytes } from '@/utils/query'
import type { FileItem } from '@/types/file'

export const FileDetailPage = () => {
  const { fileId = '' } = useParams<{ fileId: string }>()
  const [isPasswordModalOpen, setPasswordModalOpen] = useState(false)
  const queryClient = useQueryClient()

  const fileQuery = useFile(fileId)
  const treeQuery = useFileTree(fileId)

  const childrenQuery = useQuery({
    queryKey: ['file-children', fileId],
    queryFn: () => getFileChildren(fileId, 1, 50),
    enabled: !!fileId,
  })

  const recordsPreviewQuery = useQuery({
    queryKey: ['file-records', fileId, 1, 5],
    queryFn: () => getFileRecords(fileId, 1, 5),
    enabled: !!fileId,
  })

  const file = fileQuery.data as FileItem | undefined
  const isPasswordBlocked =
    file?.processing_status?.toLowerCase() === 'password_required' ||
    file?.processing_status?.toLowerCase() === 'wrong_password'
  const status = file?.processing_status || 'unknown'

  const hasParent = Boolean(file?.parent_file_id)
  const recordCount = recordsPreviewQuery.data?.total ?? 0

  const children = childrenQuery.data?.children || []

  const fileName = useMemo(() => file?.original_name || 'File detail', [file?.original_name])
  const parentLink = file?.parent_file_id ? `/files/${file.parent_file_id}` : null

  if (fileQuery.isError) {
    return <div className="text-rose-300">File not found.</div>
  }

  return (
    <div className="space-y-4">
      <PageHeader title={fileName} subtitle={file?.id ? `File ID: ${file.id}` : ''} />

      <div className="grid gap-4 lg:grid-cols-2">
        <Card title="Metadata" subtitle="File attributes">
          {file ? (
            <div className="grid gap-2 text-sm md:grid-cols-2">
              <div>
                <div className="text-slate-400">Original name</div>
                <div>{file.original_name}</div>
              </div>
              <div>
                <div className="text-slate-400">Normalized name</div>
                <div>{file.normalized_name || '—'}</div>
              </div>
              <div>
                <div className="text-slate-400">Extension</div>
                <div>{file.extension || '—'}</div>
              </div>
              <div>
                <div className="text-slate-400">Detected file type</div>
                <FileTypeBadge fileType={file.detected_file_type || file.detected_mime_type} />
              </div>
              <div>
                <div className="text-slate-400">Size</div>
                <div>{formatBytes(file.size_bytes)}</div>
              </div>
              <div>
                <div className="text-slate-400">Created</div>
                <div>{parseTime(file.created_at)}</div>
              </div>
              <div>
                <div className="text-slate-400">Storage path</div>
                <div className="break-all text-xs text-slate-300">{file.storage_path}</div>
              </div>
              <div>
                <div className="text-slate-400">SHA256</div>
                <div className="break-all text-xs text-slate-300">{file.sha256_hash || '—'}</div>
              </div>
              <div>
                <div className="text-slate-400">Password protected</div>
                <div>{file.is_password_protected ? 'Yes' : 'No'}</div>
              </div>
              {isPasswordBlocked ? (
                <button
                  className="rounded border border-amber-400/60 px-3 py-1.5 text-xs text-amber-100"
                  onClick={() => setPasswordModalOpen(true)}
                >
                  Enter archive password
                </button>
              ) : null}
              <div>
                <div className="text-slate-400">Archive file</div>
                <div>{file.is_archive ? 'Yes' : 'No'}</div>
              </div>
              <div>
                <div className="text-slate-400">Parent</div>
                <div>
                  {hasParent && parentLink ? (
                    <Link className="text-blue-300 underline" to={parentLink}>
                      {file.parent_file_id}
                    </Link>
                  ) : (
                    'Root file'
                  )}
                </div>
              </div>
            </div>
          ) : (
            <div>Loading…</div>
          )}
        </Card>

        <Card title="Processing status">
          <div className="space-y-2 text-sm">
            <div>
              <span className="text-slate-400">Current:</span> <FileStatusBadge status={status} />
            </div>
            <div>
              <span className="text-slate-400">Upload ID:</span> {file?.upload_id || '—'}
            </div>
            <div>
              <span className="text-slate-400">Parsed record count:</span> {recordCount}
            </div>
            <div>
              <span className="text-slate-400">Last updated:</span> {parseTime(file?.updated_at || null)}
            </div>
          </div>
        </Card>
      </div>

      <Card title="Children" subtitle="Direct file children">
        {childrenQuery.isLoading ? (
          <div className="text-slate-400">Loading children…</div>
        ) : children.length === 0 ? (
          <div className="text-slate-400">No child files.</div>
        ) : (
          <FileTable files={children} isLoading={childrenQuery.isLoading} />
        )}
      </Card>

      <Card title="Archive tree">
        {treeQuery.isLoading ? (
          <div className="text-slate-400">Loading archive tree…</div>
        ) : treeQuery.isError ? (
          <div className="text-rose-300">Unable to load archive tree.</div>
        ) : (
          <ArchiveTree tree={treeQuery.data || null} rootName={file?.original_name} />
        )}
      </Card>

      <Card title="Parsed records">
        <ParsedRecordsPreview fileId={fileId} pageSize={5} initialData={recordsPreviewQuery.data || undefined} />
      </Card>

      <PasswordRequiredModal
        fileId={fileId}
        open={isPasswordModalOpen}
        onClose={() => setPasswordModalOpen(false)}
        onSuccess={() => {
          queryClient.invalidateQueries({ queryKey: ['file', fileId] })
          queryClient.invalidateQueries({ queryKey: ['jobs'] })
          queryClient.invalidateQueries({ queryKey: ['files'] })
          queryClient.invalidateQueries({ queryKey: ['file-tree', fileId] })
        }}
      />
    </div>
  )
}
