import { Badge } from '@/components/ui/Badge'

type JobStatusBadgeProps = {
  status: string
}

const statusTone = (status: string): 'blue' | 'amber' | 'green' | 'red' | 'gray' => {
  const normalized = status.toLowerCase()

  if (normalized.includes('complete') || normalized.includes('done') || normalized.includes('success')) return 'green'
  if (normalized.includes('wrong_password') || normalized.includes('password_required')) return 'amber'
  if (
    normalized.includes('fail') ||
    normalized.includes('error') ||
    normalized.includes('abort') ||
    normalized.includes('dead')
  )
    return 'red'
  if (
    normalized.includes('queue') ||
    normalized.includes('pending') ||
    normalized.includes('waiting') ||
    normalized.includes('retry') ||
    normalized.includes('retrying')
  )
    return 'amber'
  if (normalized.includes('run') || normalized.includes('process') || normalized.includes('active')) return 'blue'

  return 'gray'
}

export const JobStatusBadge = ({ status }: JobStatusBadgeProps) => {
  return <Badge tone={statusTone(status)}>{status}</Badge>
}
