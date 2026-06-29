import { useRef, useState, type ChangeEvent, type DragEvent } from 'react'
import { Card } from '@/components/ui/Card'

type ValidationErrorHandler = (message: string | null) => void

type FileUploaderProps = {
  file: File | null
  onFileChange: (file: File | null) => void
  onValidationError?: ValidationErrorHandler
  disabled?: boolean
}

export const FileUploader = ({
  file,
  onFileChange,
  onValidationError,
  disabled,
}: FileUploaderProps) => {
  const fileInputRef = useRef<HTMLInputElement>(null)
  const [isDragging, setIsDragging] = useState(false)

  const validateAndSet = (next: File | null) => {
    if (!next) {
      onFileChange(null)
      onValidationError?.(null)
      return
    }

    if (next.size <= 0) {
      onFileChange(null)
      onValidationError?.('File must be greater than 0 bytes.')
      return
    }

    onValidationError?.(null)
    onFileChange(next)
  }

  const onInputChange = (event: ChangeEvent<HTMLInputElement>) => {
    const selected = event.target.files?.[0] || null
    validateAndSet(selected)
  }

  const onDragOver = (event: DragEvent<HTMLDivElement>) => {
    event.preventDefault()
    if (!disabled) {
      setIsDragging(true)
    }
  }

  const onDragLeave = (event: DragEvent<HTMLDivElement>) => {
    event.preventDefault()
    setIsDragging(false)
  }

  const onDrop = (event: DragEvent<HTMLDivElement>) => {
    event.preventDefault()
    setIsDragging(false)

    if (disabled) return
    const dropped = event.dataTransfer.files?.[0] || null
    validateAndSet(dropped)
  }

  return (
    <Card title="Select file" subtitle="Drag a file here or use the browse button">
      <div className="space-y-4">
        <div
          className={`rounded-xl border border-dashed p-6 text-center transition ${
            isDragging ? 'border-blue-400 bg-blue-900/20' : 'border-slate-600 bg-slate-900/30'
          } ${disabled ? 'cursor-not-allowed opacity-60' : 'cursor-pointer'}`}
          onDragOver={onDragOver}
          onDragLeave={onDragLeave}
          onDrop={onDrop}
          onClick={() => !disabled && fileInputRef.current?.click()}
          role="button"
          tabIndex={0}
          onKeyDown={(event) => {
            if (event.key === 'Enter' || event.key === ' ') {
              event.preventDefault()
              !disabled && fileInputRef.current?.click()
            }
          }}
        >
          <p className="text-sm text-slate-200">
            Drop your file here, then we will validate and send it to the uploader.
          </p>
          <p className="mt-1 text-xs text-slate-300">or click to browse</p>
          <input
            ref={fileInputRef}
            className="hidden"
            type="file"
            onChange={onInputChange}
            disabled={disabled}
          />
        </div>

        {file ? <p className="text-sm text-slate-200">Selected: {file.name}</p> : null}
      </div>
    </Card>
  )
}
