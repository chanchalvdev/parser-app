import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { Button } from '@/components/ui/Button'
import { PageHeader } from '@/components/layout/PageHeader'
import { ApiError } from '@/types/api'
import { completeUpload, initiateUpload, uploadFileToPresignedUrl } from '@/services/uploadApi'
import type { UploadCompleteResponse, UploadInitiateRequest } from '@/types/file'
import { FileUploader } from '@/components/upload/FileUploader'
import { OptionalPasswordInput } from '@/components/upload/OptionalPasswordInput'
import { UploadProgress } from '@/components/upload/UploadProgress'
import { UploadResultCard } from '@/components/upload/UploadResultCard'
import { UploadHistoryMiniList, type UploadHistoryEntry } from '@/components/upload/UploadHistoryMiniList'

type UploadStage = 'idle' | 'initiating' | 'uploading' | 'completing' | 'done' | 'error'

const UPLOAD_HISTORY_KEY = 'upload-history-mini'
const UPLOAD_HISTORY_LIMIT = 10

const readUploadHistory = (): UploadHistoryEntry[] => {
  if (typeof window === 'undefined') return []

  try {
    const raw = window.localStorage.getItem(UPLOAD_HISTORY_KEY)
    if (!raw) return []
    const parsed = JSON.parse(raw)
    if (!Array.isArray(parsed)) return []
    return parsed.slice(0, UPLOAD_HISTORY_LIMIT).map((entry) => ({
      upload_id: String(entry.upload_id || ''),
      file_id: String(entry.file_id || ''),
      job_id: String(entry.job_id || ''),
      file_name: String(entry.file_name || 'uploaded file'),
      status: String(entry.status || 'unknown'),
      created_at: String(entry.created_at || new Date().toISOString()),
    }))
  } catch {
    return []
  }
}

const persistUploadHistory = (items: UploadHistoryEntry[]) => {
  if (typeof window === 'undefined') return

  try {
    window.localStorage.setItem(UPLOAD_HISTORY_KEY, JSON.stringify(items))
  } catch {
    // Ignore storage failures to avoid blocking uploads.
  }
}

const getErrorMessage = (error: unknown): string => {
  if (error instanceof ApiError) {
    return error.payload?.message || error.message
  }
  if (error instanceof Error) {
    return error.message
  }
  return 'Upload failed. Please try again.'
}

export const UploadPage = () => {
  const [selectedFile, setSelectedFile] = useState<File | null>(null)
  const [isPasswordEnabled, setIsPasswordEnabled] = useState(false)
  const [password, setPassword] = useState('')
  const [uploadProgress, setUploadProgress] = useState(0)
  const [uploadStage, setUploadStage] = useState<UploadStage>('idle')
  const [validationError, setValidationError] = useState<string | null>(null)
  const [apiError, setApiError] = useState<string | null>(null)
  const [result, setResult] = useState<UploadCompleteResponse | null>(null)
  const [history, setHistory] = useState<UploadHistoryEntry[]>(() => readUploadHistory())

  const mutateUpload = useMutation({
    mutationFn: async () => {
      if (!selectedFile) {
        throw new Error('Please select a file before uploading.')
      }
      if (selectedFile.size <= 0) {
        throw new Error('File size must be greater than zero.')
      }

      const request: UploadInitiateRequest = {
        file_name: selectedFile.name,
        content_type: selectedFile.type || 'application/octet-stream',
        size_bytes: selectedFile.size,
        password_provided: isPasswordEnabled,
      }

      setUploadStage('initiating')
      setUploadProgress(2)

      const initiated = await initiateUpload(request)

      setUploadStage('uploading')
      setUploadProgress(10)
      await uploadFileToPresignedUrl(initiated.upload_url, selectedFile, ({ percentage }) =>
        setUploadProgress(Math.max(percentage, 10)),
      )

      setUploadStage('completing')
      setUploadProgress(99)
      return completeUpload({ upload_id: initiated.upload_id })
    },
    onMutate: () => {
      setResult(null)
      setApiError(null)
      setValidationError(null)
    },
    onSuccess: (next) => {
      setUploadStage('done')
      setUploadProgress(100)
      setResult(next)
      setIsPasswordEnabled(false)
      setPassword('')

      if (!selectedFile) {
        return
      }

      const nextEntry: UploadHistoryEntry = {
        upload_id: next.upload_id,
        file_id: next.file_id,
        job_id: next.job_id,
        file_name: selectedFile.name,
        status: next.status,
        created_at: new Date().toISOString(),
      }

      setHistory((current) => {
        const deduped = current.filter((item) => item.job_id !== next.job_id)
        const updated = [nextEntry, ...deduped].slice(0, UPLOAD_HISTORY_LIMIT)
        persistUploadHistory(updated)
        return updated
      })
    },
    onError: (error) => {
      setUploadStage('error')
      setApiError(getErrorMessage(error))
      setUploadProgress(0)
    },
  })

  const handleSubmit = () => {
    mutateUpload.mutate()
  }

  return (
    <div className="space-y-4">
      <PageHeader title="Upload" subtitle="Start ingestion by uploading a file and monitoring job progress." />

      <div className="grid gap-4 lg:grid-cols-[2fr,1fr]">
        <div className="space-y-4">
          <FileUploader
            file={selectedFile}
            disabled={mutateUpload.isPending}
            onFileChange={(nextFile) => {
              setSelectedFile(nextFile)
              setUploadStage('idle')
              setUploadProgress(0)
              if (nextFile) {
                setResult(null)
                setApiError(null)
              }
            }}
            onValidationError={(message) => {
              setValidationError(message)
              setUploadStage('idle')
              setUploadProgress(0)
              if (message) {
                setResult(null)
                setApiError(null)
              }
            }}
          />

          <OptionalPasswordInput
            enabled={isPasswordEnabled}
            password={password}
            onEnabledChange={setIsPasswordEnabled}
            onPasswordChange={(value) => setPassword(value)}
            disabled={mutateUpload.isPending}
          />

          <Button disabled={!selectedFile || mutateUpload.isPending} onClick={handleSubmit}>
            {mutateUpload.isPending ? 'Uploading…' : 'Start upload'}
          </Button>

          {validationError || apiError ? (
            <p className="text-sm text-rose-300">{validationError || apiError}</p>
          ) : null}

          <UploadProgress stage={uploadStage} percentage={uploadProgress} fileName={selectedFile?.name} />
          {result ? <UploadResultCard fileName={selectedFile?.name || 'uploaded file'} result={result} /> : null}
        </div>

        <div>
          <UploadHistoryMiniList items={history} />
        </div>
      </div>
    </div>
  )
}
