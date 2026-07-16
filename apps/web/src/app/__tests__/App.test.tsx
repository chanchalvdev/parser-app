import { render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import App from '../App'

describe('App', () => {
  beforeEach(() => {
    // The jobs list fetches on mount; return an empty list.
    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ total: 0, page: 1, page_size: 50, jobs: [] }),
    }) as unknown as typeof fetch
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('renders the upload card and jobs surface', async () => {
    render(
      <MemoryRouter initialEntries={['/']}>
        <App />
      </MemoryRouter>,
    )

    expect(screen.getByText('Upload a file')).toBeInTheDocument()
    expect(screen.getByText('Upload & parse')).toBeInTheDocument()
    await waitFor(() =>
      expect(screen.getByText(/No jobs yet/i)).toBeInTheDocument(),
    )
  })
})
