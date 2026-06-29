import { Badge } from '@/components/ui/Badge'

type FileStatusBadgeProps = {
  status: string
}

const statusTone = (status: string): 'blue' | 'amber' | 'green' | 'red' | 'gray' => {
  const normalized = status.toLowerCase()

  if (normalized.includes('complete')) return 'green'
  if (normalized.includes('done')) return 'green'
  if (normalized.includes('wrong_password')) return 'amber'
  if (normalized.includes('fail')) return 'red'
  if (normalized.includes('error')) return 'red'
  if (normalized.includes('queued')) return 'amber'
  if (normalized.includes('processing')) return 'blue'
  if (normalized.includes('password_required')) return 'amber'

  return 'gray'
}

export const FileStatusBadge = ({ status }: FileStatusBadgeProps) => {
  return <Badge tone={statusTone(status)}>{status}</Badge>
}
