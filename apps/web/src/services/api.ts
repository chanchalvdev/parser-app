import { toQueryString } from '@/utils/query'

export class ApiError extends Error {
  status: number
  details?: unknown

  constructor(status: number, message: string, details?: unknown) {
    super(message)
    this.status = status
    this.name = 'ApiError'
    this.details = details
  }
}

const BASE_URL = (import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1').replace(/\/$/, '')

async function readErrorBody(response: Response) {
  const text = await response.text()
  try {
    if (!text) return undefined
    return JSON.parse(text)
  } catch (error) {
    return { message: text }
  }
}

export async function apiRequest<T>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  const response = await fetch(`${BASE_URL}${path.startsWith('/') ? path : `/${path}`}`, {
    headers: {
      'Content-Type': 'application/json',
      ...(options.headers || {}),
    },
    ...options,
  })

  if (!response.ok) {
    const details = await readErrorBody(response)
    throw new ApiError(response.status, `Request failed: ${response.statusText}`, details)
  }

  if (response.status === 204) return null as T
  const text = await response.text()
  if (!text) return null as T
  return JSON.parse(text) as T
}

export const apiGet = <T>(path: string, params?: Record<string, string | number | boolean | undefined>): Promise<T> => {
  const query = params ? toQueryString(params) : ''
  return apiRequest<T>(`${path}${query}`)
}

export const apiPost = <T>(path: string, payload: unknown): Promise<T> => {
  return apiRequest<T>(path, {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}
