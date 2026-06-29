import { Link } from 'react-router-dom'
import { Card } from '@/components/ui/Card'
import type { UploadCompleteResponse } from '@/types/file'

type UploadResultCardProps = {
  fileName: string
  result: UploadCompleteResponse
}

export const UploadResultCard = ({ fileName, result }: UploadResultCardProps) => {
  return (
    <Card title="Upload completed" subtitle="Your file is now queued for processing">
      <div className="space-y-2 text-sm">
        <div>
          <span className="text-slate-400">File:</span> {fileName}
        </div>
        <div>
          <span className="text-slate-400">Upload ID:</span> {result.upload_id}
        </div>
        <div>
          <span className="text-slate-400">File ID:</span> {result.file_id}
        </div>
        <div>
          <span className="text-slate-400">Job ID:</span> {result.job_id}
        </div>
        <div>
          <span className="text-slate-400">Status:</span> {result.status}
        </div>
        <Link to={`/jobs/${result.job_id}`} className="inline-block text-blue-300 underline">
          Open job detail
        </Link>
      </div>
    </Card>
  )
}
