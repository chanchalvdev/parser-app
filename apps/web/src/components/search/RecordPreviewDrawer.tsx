import { Link } from 'react-router-dom'
import { Card } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { parseTime } from '@/utils/query'
import { EntityBadges } from '@/components/search/EntityBadges'
import type { SearchResult } from '@/types/search'

const renderHighlightedContent = (value: string | null | undefined) => {
  if (!value) return null

  const hasMarkup = /<[^>]+>/.test(value)
  if (!hasMarkup) {
    return <pre className="max-h-72 overflow-auto whitespace-pre-wrap break-words text-sm">{value}</pre>
  }

  return <div className="max-h-72 overflow-auto text-sm" dangerouslySetInnerHTML={{ __html: value }} />
}

type RecordPreviewDrawerProps = {
  record: SearchResult | null
  onClose: () => void
}

export const RecordPreviewDrawer = ({ record, onClose }: RecordPreviewDrawerProps) => {
  if (!record) return null

  return (
    <div className="fixed inset-0 z-50 flex items-start justify-end">
      <button
        aria-label="Close preview"
        onClick={onClose}
        className="absolute inset-0 bg-slate-950/70"
        type="button"
      />
      <aside className="relative z-10 h-full w-full max-w-xl overflow-auto border-l border-slate-700 bg-slate-950/95 p-4 backdrop-blur">
        <Card title="Record preview" subtitle={`Record ${record.record_id}`}>
          <div className="space-y-3 text-sm">
            <div>
              <div className="text-slate-300">Source file</div>
              <Link className="text-blue-300 hover:underline" to={`/files/${record.file_id}`}>
                {record.source_file_name || record.file_id}
              </Link>
            </div>

            <div>
              <div className="text-slate-300">Job</div>
              <Link className="text-blue-300 hover:underline" to={`/jobs/${record.job_id}`}>
                {record.job_id}
              </Link>
            </div>

            <div>
              <div className="text-slate-300">Type</div>
              <div>{record.record_type || 'Unknown'}</div>
            </div>

            <div>
              <div className="text-slate-300">Created</div>
              <div>{parseTime(record.created_at)}</div>
            </div>

            <div>
              <div className="text-slate-300">Content</div>
              {renderHighlightedContent(record.highlight || record.content_preview)}
            </div>

            <div>
              <div className="text-slate-300">Entities</div>
              <EntityBadges entities={record.entities} />
            </div>
          </div>

          <div className="mt-4">
            <Button tone="ghost" onClick={onClose}>
              Close
            </Button>
          </div>
        </Card>
      </aside>
    </div>
  )
}
