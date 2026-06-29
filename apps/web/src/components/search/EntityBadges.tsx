import { Badge } from '@/components/ui/Badge'

type EntityBadgesProps = {
  entities?: Record<string, unknown>
  compact?: boolean
}

const collectValues = (value: unknown): string[] => {
  if (value === undefined || value === null) return []

  if (Array.isArray(value)) {
    return value
      .flatMap((entry) => collectValues(entry))
      .filter((entry): entry is string => typeof entry === 'string' && entry.length > 0)
  }

  if (typeof value === 'string') return value.trim() ? [value.trim()] : []
  if (typeof value === 'number' || typeof value === 'boolean') return [String(value)]

  return []
}

const getBucketGroups = (entities: Record<string, unknown>) => {
  const orderedKeys = ['ip_addresses', 'emails', 'domains', 'urls', 'hashes']
  const known = orderedKeys.flatMap((key) => collectValues(entities[key]).map((entry) => ({ key, value: entry })))

  return known
}

export const EntityBadges = ({ entities, compact = false }: EntityBadgesProps) => {
  if (!entities || Object.keys(entities).length === 0) {
    return <span className="text-xs text-slate-500">No entities</span>
  }

  const flattened = getBucketGroups(entities)

  if (flattened.length === 0) {
    return <span className="text-xs text-slate-500">No entities</span>
  }

  return (
    <div className={`flex flex-wrap gap-1 ${compact ? 'max-h-20 overflow-hidden' : ''}`}>
      {flattened.map((item, index) => (
        <Badge tone="gray" key={`${item.key}-${item.value}-${index}`}>
          {item.value}
        </Badge>
      ))}
    </div>
  )
}
