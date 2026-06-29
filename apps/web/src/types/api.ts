export type ApiErrorPayload = {
  message?: string
  error?: string
  code?: string
  details?: unknown
  request_id?: string
  [key: string]: unknown
}

export type ApiPagination = {
  total: number
  page: number
  page_size: number
}

export class ApiError extends Error {
  public readonly status: number
  public readonly statusText: string
  public readonly payload?: ApiErrorPayload
  public readonly url: string

  constructor(status: number, statusText: string, payload: ApiErrorPayload | undefined, url: string) {
    super(`Request failed with ${status} ${statusText}`)
    this.name = 'ApiError'
    this.status = status
    this.statusText = statusText
    this.payload = payload
    this.url = url
  }
}

export type TimeoutOptions = {
  timeoutMs?: number
}

export type QueryValue = string | number | boolean | null | undefined

export type QueryParams = Record<string, QueryValue | QueryValue[]>

export const DEFAULT_TIMEOUT_MS = 15_000

export const buildQueryString = (params?: QueryParams): string => {
  if (!params) return ''

  const chunks: string[] = []

  Object.entries(params).forEach(([key, value]) => {
    if (value === undefined || value === null || value === '') {
      return
    }

    if (Array.isArray(value)) {
      value
        .filter((entry) => entry !== undefined && entry !== null && entry !== '')
        .forEach((entry) => {
          chunks.push(`${encodeURIComponent(key)}=${encodeURIComponent(String(entry))}`)
        })
      return
    }

    chunks.push(`${encodeURIComponent(key)}=${encodeURIComponent(String(value))}`)
  })

  if (chunks.length === 0) return ''
  return `?${chunks.join('&')}`
}

export const parseJsonResponse = async <T>(response: Response): Promise<T> => {
  const text = await response.text()
  if (!text) {
    return undefined as unknown as T
  }

  try {
    return JSON.parse(text) as T
  } catch {
    // Keep parser tolerance for API responses that are not JSON for non-application/json cases.
    return text as unknown as T
  }
}
