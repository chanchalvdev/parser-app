export const toQueryString = (params: Record<string, string | number | boolean | undefined>): string => {
  const values = Object.entries(params)
    .filter(([, value]) => value !== undefined)
    .map(([key, value]) => `${encodeURIComponent(key)}=${encodeURIComponent(String(value))}`)

  return values.length === 0 ? '' : `?${values.join('&')}`
}

export const parseTime = (value?: string | null): string => {
  if (!value) return '—'
  const parsed = new Date(value)
  if (Number.isNaN(parsed.getTime())) return value
  return parsed.toLocaleString()
}

export const formatBytes = (size = 0): string => {
  if (size === 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let value = size
  let unit = 0
  while (value >= 1024 && unit < units.length - 1) {
    value /= 1024
    unit += 1
  }
  return `${value.toFixed(1)} ${units[unit]}`
}

export const buildClassName = (...parts: Array<string | undefined | false | null>) =>
  parts.filter(Boolean).join(' ')
