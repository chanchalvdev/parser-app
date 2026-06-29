import {
  ApiError,
  ApiErrorPayload,
  buildQueryString,
  DEFAULT_TIMEOUT_MS,
  parseJsonResponse,
  QueryParams,
  TimeoutOptions,
} from '@/types/api'

export type ApiClientRequestOptions = {
  method?: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE'
  query?: QueryParams
  body?: unknown
  headers?: HeadersInit
  timeoutMs?: number
  credentials?: RequestCredentials
  mode?: RequestMode
  cache?: RequestCache
}

const resolveApiBaseUrl = (): string => {
  return (import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1').replace(/\/$/, '')
}

const createHeaders = (body: unknown, headers?: HeadersInit): HeadersInit => {
  const next = new Headers(headers)

  if (body !== undefined && body !== null && !(body instanceof FormData) && !(body instanceof Blob)) {
    if (!next.has('Content-Type')) {
      next.set('Content-Type', 'application/json')
    }
  }

  if (!next.has('Accept')) {
    next.set('Accept', 'application/json')
  }

  return next
}

const buildUrl = (path: string, query?: QueryParams): string => {
  const normalizedPath = path.startsWith('/') ? path : `/${path}`
  const baseUrl = resolveApiBaseUrl()

  return `${baseUrl}${normalizedPath}${buildQueryString(query)}`
}

const createTimeoutController = (timeoutMs: number): { signal: AbortSignal; clear: () => void } => {
  const controller = new AbortController()
  const id = setTimeout(() => {
    controller.abort()
  }, timeoutMs)

  return {
    signal: controller.signal,
    clear: () => clearTimeout(id),
  }
}

async function request<T>(path: string, options: ApiClientRequestOptions = {}): Promise<T> {
  const {
    method = 'GET',
    query,
    body,
    headers,
    timeoutMs = DEFAULT_TIMEOUT_MS,
    credentials,
    mode,
    cache,
  } = options

  const url = buildUrl(path, query)

  const timeout = createTimeoutController(timeoutMs)
  const init: RequestInit = {
    method,
    headers: createHeaders(body, headers),
    credentials,
    mode,
    cache,
    signal: timeout.signal,
  }

  if (body !== undefined && body !== null) {
    init.body = body instanceof FormData || body instanceof Blob ? (body as BodyInit) : JSON.stringify(body)
  }

  try {
    const response = await fetch(url, init)
    timeout.clear()

    if (!response.ok) {
      const errorPayload = (await parseJsonResponse<ApiErrorPayload>(response)) || {
        message: response.statusText,
      }

      throw new ApiError(response.status, response.statusText, errorPayload, url)
    }

    if (response.status === 204) {
      return undefined as T
    }

    return parseJsonResponse<T>(response)
  } catch (error) {
    timeout.clear()

    if (error instanceof ApiError) {
      throw error
    }

    if (error instanceof DOMException && error.name === 'AbortError') {
      const reason: ApiErrorPayload = {
        message: `Request timeout after ${timeoutMs}ms`,
        code: 'request_timeout',
      }
      throw new ApiError(408, 'Request Timeout', reason, url)
    }

    if (error instanceof Error) {
      const wrapped: ApiErrorPayload = {
        message: error.message,
      }
      throw new ApiError(500, 'Network error', wrapped, url)
    }

    throw new ApiError(500, 'Unknown error', undefined, url)
  }
}

export const apiClient = {
  get: <T>(path: string, options: ApiClientRequestOptions & TimeoutOptions = {}): Promise<T> =>
    request<T>(path, { method: 'GET', ...options }),

  post: <T>(path: string, body?: unknown, options: ApiClientRequestOptions & TimeoutOptions = {}): Promise<T> =>
    request<T>(path, { method: 'POST', body, ...options }),

  put: <T>(path: string, body?: unknown, options: ApiClientRequestOptions & TimeoutOptions = {}): Promise<T> =>
    request<T>(path, { method: 'PUT', body, ...options }),

  patch: <T>(path: string, body?: unknown, options: ApiClientRequestOptions & TimeoutOptions = {}): Promise<T> =>
    request<T>(path, { method: 'PATCH', body, ...options }),

  del: <T>(path: string, options: ApiClientRequestOptions & TimeoutOptions = {}): Promise<T> =>
    request<T>(path, { method: 'DELETE', ...options }),
}

export const BASE_URL = resolveApiBaseUrl()
