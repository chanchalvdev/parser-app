import { Link } from 'react-router-dom'
import { Card } from '@/components/ui/Card'

export type UploadHistoryEntry = {
  upload_id: string
  file_id: string
  job_id: string
  file_name: string
  status: string
  created_at: string
}

type UploadHistoryMiniListProps = {
  items: UploadHistoryEntry[]
}

export const UploadHistoryMiniList = ({ items }: UploadHistoryMiniListProps) => {
  if (items.length === 0) {
    return (
      <Card title="Recent uploads" subtitle="Your latest successful uploads">
        <p className="text-sm text-slate-400">No uploads yet.</p>
      </Card>
    )
  }

  return (
    <Card title="Recent uploads" subtitle="Latest upload completions">
      <ul className="space-y-2 text-sm">
        {items.map((item) => (
          <li key={item.upload_id} className="rounded-lg border border-slate-700/60 p-2">
            <div className="text-slate-100">{item.file_name}</div>
            <div className="mt-1 text-xs text-slate-400">
              status: {item.status} · job: {item.job_id} · file: {item.file_id}
            </div>
            <Link to={`/jobs/${item.job_id}`} className="mt-1 inline-block text-xs text-blue-300 underline">
              Open job
            </Link>
          </li>
        ))}
      </ul>
    </Card>
  )
}
