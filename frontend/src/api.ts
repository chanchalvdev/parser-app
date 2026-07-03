import type { DashboardSummary, FileNode, Job, SearchItem, Settings, Upload } from './types'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'
const API_KEY = import.meta.env.VITE_API_KEY || ''

const normalizeHeaders = (headers?: HeadersInit): Record<string, string> => {
  if (!headers) return {}
  if (headers instanceof Headers) {
    return Object.fromEntries(headers.entries())
  }
  if (Array.isArray(headers)) {
    return Object.fromEntries(headers)
  }
  return headers as Record<string, string>
}

const headersWithAuth = (headers: HeadersInit | undefined = undefined) => {
  const next = normalizeHeaders(headers)
  return API_KEY ? { ...next, 'X-API-Key': API_KEY } : next
}

async function callApi<T>(path: string, init: RequestInit = {}): Promise<T> {
  const requestHeaders = headersWithAuth(init.headers)
  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...init,
    headers: {
      ...requestHeaders,
    },
  })
  if (!response.ok) {
    const text = await response.text()
    throw new Error(text || `HTTP ${response.status}`)
  }
  if (response.status === 204) {
    return undefined as T
  }
  return (await response.json()) as T
}

export const api = {
  health: () => callApi<{ status: string }>('/health'),
  listUploads: (query = '') => callApi<Upload[]>(`/api/v1/uploads${query}`),
  getUpload: (id: string) => callApi<Upload>(`/api/v1/uploads/${id}`),
  uploadFile: async (file: File, archivePassword?: string): Promise<{ upload_id: string; job_id: string; status: string }> => {
    const form = new FormData()
    form.append('file', file)
    if (archivePassword) {
      form.append('archive_password', archivePassword)
    }
    return callApi<{ upload_id: string; job_id: string; status: string }>('/api/v1/uploads', {
      method: 'POST',
      body: form,
    })
  },
  listJobs: (query = '') => callApi<Job[]>(`/api/v1/jobs${query}`),
  getJob: (id: string) => callApi<Job>(`/api/v1/jobs/${id}`),
  jobEvents: (jobId: string) => callApi<unknown[]>(`/api/v1/jobs/${jobId}/events`),
  retryWithPassword: (jobId: string, password: string) =>
    callApi(`/api/v1/jobs/${jobId}/password`, {
      method: 'POST',
      body: JSON.stringify({ password }),
      headers: { 'Content-Type': 'application/json' },
    }),
  search: (q: string, limit = 25) =>
    callApi<SearchItem[]>(`/api/v1/search?q=${encodeURIComponent(q)}&limit=${limit}`),
  fileTree: (uploadId: string) => callApi<FileNode[]>(`/api/v1/files/${uploadId}/tree`),
  dashboardSummary: () => callApi<DashboardSummary>('/api/v1/dashboard/summary'),
  auditLogs: () => callApi<unknown[]>('/api/v1/audit'),
  settings: () => callApi<Settings>('/api/v1/admin/settings'),
  updateSettings: (payload: Record<string, any>) =>
    callApi<Settings>('/api/v1/admin/settings', {
      method: 'PUT',
      body: JSON.stringify(payload),
      headers: { 'Content-Type': 'application/json' },
    }),
}
