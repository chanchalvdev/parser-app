import { Card } from '@/components/ui/Card'

type UploadProgressStage = 'idle' | 'initiating' | 'uploading' | 'completing' | 'done' | 'error'

type UploadProgressProps = {
  stage: UploadProgressStage
  percentage: number
  fileName?: string
}

const statusByStage: Record<UploadProgressStage, string> = {
  idle: 'Ready',
  initiating: 'Preparing upload session',
  uploading: 'Uploading to object storage',
  completing: 'Finalizing ingest job',
  done: 'Completed',
  error: 'Failed',
}

export const UploadProgress = ({ stage, percentage, fileName }: UploadProgressProps) => {
  if (stage === 'idle') {
    return null
  }

  const clamped = Math.max(0, Math.min(100, percentage))

  return (
    <Card title="Upload progress" subtitle={fileName ? `File: ${fileName}` : ''}>
      <div className="space-y-3">
        <p className="text-sm text-slate-200">{statusByStage[stage]}</p>
        <div className="h-2 w-full rounded-full bg-slate-800">
          <div className="h-2 rounded-full bg-blue-500 transition-all duration-300" style={{ width: `${clamped}%` }} />
        </div>
        <p className="text-xs text-slate-300">{clamped}% complete</p>
      </div>
    </Card>
  )
}
