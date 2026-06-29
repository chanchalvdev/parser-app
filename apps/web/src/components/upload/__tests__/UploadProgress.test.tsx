import { describe, expect, it } from 'vitest'
import { render, screen } from '@testing-library/react'

import { UploadProgress } from '../UploadProgress'

describe('UploadProgress', () => {
  it('shows idle state as null', () => {
    const { container } = render(<UploadProgress stage="idle" percentage={0} />)
    expect(container.firstChild).toBeNull()
  })

  it('renders status and percent while uploading', () => {
    render(<UploadProgress stage="uploading" percentage={42} fileName="sample.txt" />)

    expect(screen.getByText('Uploading to object storage')).toBeDefined()
    expect(screen.getByText('42% complete')).toBeDefined()
    expect(screen.getByText('File: sample.txt')).toBeDefined()
  })

  it('clamps percent values above 100', () => {
    const { container } = render(<UploadProgress stage="completing" percentage={150} />)
    const bar = container.querySelector('.bg-blue-500') as HTMLElement
    expect(bar.style.width).toBe('100%')
  })
})

