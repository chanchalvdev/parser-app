import { beforeEach, describe, expect, it, vi } from 'vitest'

import { ApiError } from '@/types/api'
import { apiClient } from '../apiClient'

const mockFetch = vi.fn<Parameters<typeof fetch>, ReturnType<typeof fetch>>()

describe('apiClient', () => {
  beforeEach(() => {
    vi.stubEnv('VITE_API_BASE_URL', 'https://api.example.com/api/v1')
    vi.stubGlobal('fetch', mockFetch)
    mockFetch.mockReset()
  })

  it('performs GET requests and parses JSON responses', async () => {
    mockFetch.mockResolvedValueOnce(
      new Response(JSON.stringify({ status: 'ok' }), {
        headers: { 'content-type': 'application/json' },
        status: 200,
      }),
    )

    const data = await apiClient.get<{ status: string }>('/health')

    expect(data).toEqual({ status: 'ok' })
    expect(mockFetch).toHaveBeenCalledTimes(1)
    const calledWith = mockFetch.mock.calls[0]?.[0]
    expect(calledWith).toBe('https://api.example.com/api/v1/health')
  })

  it('applies query params on GET requests', async () => {
    mockFetch.mockResolvedValueOnce(
      new Response(JSON.stringify({ total: 1 }), { status: 200 }),
    )

    await apiClient.get('/files', { query: { page: 1, page_size: 25, tenant_id: 't1' } })

    const calledWith = mockFetch.mock.calls[0]?.[0] as string
    expect(calledWith).toContain('/files?page=1&page_size=25&tenant_id=t1')
  })

  it('performs POST requests with JSON body', async () => {
    mockFetch.mockResolvedValueOnce(
      new Response(JSON.stringify({ ok: true }), { status: 201 }),
    )

    const data = await apiClient.post('/jobs', { limit: 10 })

    const calledInit = mockFetch.mock.calls[0]?.[1] as RequestInit
    expect(calledInit.method).toBe('POST')
    expect(calledInit.body).toBe(JSON.stringify({ limit: 10 }))
    expect(data).toEqual({ ok: true })
  })

  it('throws ApiError for non-success responses', async () => {
    mockFetch.mockResolvedValueOnce(
      new Response(JSON.stringify({ message: 'invalid payload' }), { status: 400, statusText: 'Bad Request' }),
    )

    try {
      await apiClient.get('/errors')
      throw new Error('expected failure')
    } catch (error) {
      if (!(error instanceof ApiError)) {
        throw error
      }
      expect(error.status).toBe(400)
      expect(error.statusText).toBe('Bad Request')
      expect(error.payload?.message).toBe('invalid payload')
    }
  })

  it('wraps network failures as ApiError', async () => {
    mockFetch.mockRejectedValueOnce(new Error('network down'))

    await expect(apiClient.get('/network')).rejects.toMatchObject({
      status: 500,
      statusText: 'Network error',
    })
  })
})
