import { useMutation } from '@tanstack/react-query'
import { useMemo, useState } from 'react'
import { Button } from '@/components/ui/Button'
import { completeUpload, initiateUpload, uploadFileToPresignedUrl } from '@/services/uploadService'
import { Input } from '@/components/ui/Input'
import { Card } from '@/components/ui/Card'
import type { UploadCompleteResponse, UploadInitiateResponse } from '@/types/domain'

export const UploadPanel = () => {
  const [file, setFile] = useState<File | null>(null)
  const [isPasswordProtected, setPasswordProtected] = useState(false)
  const [result, setResult] = useState<UploadCompleteResponse | null>(null)

  const mutation = useMutation({
    mutationFn: async () => {
      if (!file) throw new Error('Please pick a file')

      const initiate = await initiateUpload({
        file_name: file.name,
        content_type: file.type || 'application/octet-stream',
        size_bytes: file.size,
        password_provided: isPasswordProtected,
      })

      await uploadFileToPresignedUrl(initiate.upload_url, file)
      return completeUpload({ upload_id: initiate.upload_id })
    },
    onSuccess: (next: UploadCompleteResponse) => setResult(next),
  })

  const canSubmit = useMemo(() => file !== null && !mutation.isPending, [file, mutation.isPending])

  return (
    <Card title="Upload file" subtitle="Upload archives or raw files for processing">
      <div className="space-y-4">
        <Input
          type="file"
          onChange={(event) => {
            setResult(null)
            const selected = event.target.files?.[0] || null
            setFile(selected)
          }}
        />
        <label className="flex items-center gap-2 text-sm text-slate-200">
          <input
            type="checkbox"
            checked={isPasswordProtected}
            onChange={(event) => setPasswordProtected(event.target.checked)}
          />
          Archive is password-protected
        </label>
        <div>
          <Button onClick={() => mutation.mutate()} disabled={!canSubmit}>
            {mutation.isPending ? 'Uploading…' : 'Upload'}
          </Button>
        </div>
        {mutation.isError ? <p className="text-sm text-rose-300">{(mutation.error as Error).message}</p> : null}
        {result ? (
          <div className="rounded-md border border-emerald-400/60 bg-emerald-900/20 p-3 text-sm text-emerald-100">
            Upload complete
            <div>upload_id: {result.upload_id}</div>
            <div>file_id: {result.file_id}</div>
            <div>job_id: {result.job_id}</div>
            <div>status: {result.status}</div>
          </div>
        ) : null}
      </div>
    </Card>
  )
}
