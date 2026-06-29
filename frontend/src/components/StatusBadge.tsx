import { Badge, type BadgeVariant } from './Badge'

export type JobStatus =
  | 'queued'
  | 'running'
  | 'completed'
  | 'failed'
  | 'retrying'
  | 'cancelled'
  | string

export type UploadStatus =
  | 'uploading'
  | 'completed'
  | 'failed'
  | 'processing'
  | 'pending'
  | string

export type StatusKind = 'job' | 'upload'

export interface StatusBadgeProps {
  status: string
  kind?: StatusKind
  className?: string
}

interface StatusMeta {
  variant: BadgeVariant
  label: string
}

const JOB_META: Record<string, StatusMeta> = {
  queued:     { variant: 'neutral', label: 'Queued' },
  running:    { variant: 'info',    label: 'Running' },
  completed:  { variant: 'success', label: 'Completed' },
  failed:     { variant: 'danger',  label: 'Failed' },
  retrying:   { variant: 'warning', label: 'Retrying' },
  cancelled:  { variant: 'neutral', label: 'Cancelled' },
  pending:    { variant: 'neutral', label: 'Pending' },
  processing: { variant: 'info',    label: 'Processing' },
}

const UPLOAD_META: Record<string, StatusMeta> = {
  uploading:  { variant: 'info',    label: 'Uploading' },
  completed:  { variant: 'success', label: 'Completed' },
  failed:     { variant: 'danger',  label: 'Failed' },
  processing: { variant: 'info',    label: 'Processing' },
  pending:    { variant: 'neutral', label: 'Pending' },
  queued:     { variant: 'neutral', label: 'Queued' },
}

function humanize(value: string): string {
  if (!value) return 'Unknown'
  return value
    .replace(/[_-]+/g, ' ')
    .replace(/\s+/g, ' ')
    .trim()
    .replace(/\b\w/g, (c) => c.toUpperCase())
}

export function StatusBadge({ status, kind = 'job', className }: StatusBadgeProps) {
  const lookup = kind === 'upload' ? UPLOAD_META : JOB_META
  const normalized = (status || '').toLowerCase()
  const meta = lookup[normalized] ?? { variant: 'neutral' as BadgeVariant, label: humanize(status) }

  return (
    <Badge variant={meta.variant} size="md" className={className}>
      {meta.label}
    </Badge>
  )
}
