import { Badge } from '@/components/ui/Badge'

type FileTypeBadgeProps = {
  fileType?: string | null
}

const typeTone = (value?: string | null): 'blue' | 'gray' | 'green' | 'amber' | 'red' => {
  if (!value) return 'gray'

  const normalized = value.toLowerCase()
  if (normalized.includes('archive') || normalized.includes('zip') || normalized.includes('rar') || normalized.includes('7z')) {
    return 'amber'
  }

  if (normalized.includes('text') || normalized.includes('json') || normalized.includes('log')) {
    return 'blue'
  }

  return 'gray'
}

export const FileTypeBadge = ({ fileType }: FileTypeBadgeProps) => {
  return <Badge tone={typeTone(fileType)}>{fileType || 'unknown'}</Badge>
}
