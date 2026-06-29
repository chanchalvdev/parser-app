import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { submitFilePassword } from '@/services/filesApi'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import type { SubmitFilePasswordRequest } from '@/types/file'

type PasswordRequiredModalProps = {
  fileId: string
  open: boolean
  onClose: () => void
  onSuccess?: () => void
}

const extractMessage = (error: unknown): string => {
  if (typeof error === 'object' && error !== null && 'message' in error) {
    const maybeError = error as { message?: string }
    return maybeError.message || 'Unable to submit password.'
  }
  if (error instanceof Error) {
    return error.message
  }
  return 'Unable to submit password.'
}

export const PasswordRequiredModal = ({ fileId, open, onClose, onSuccess }: PasswordRequiredModalProps) => {
  const [password, setPassword] = useState('')
  const [errorMessage, setErrorMessage] = useState('')

  const mutation = useMutation({
    mutationFn: (request: SubmitFilePasswordRequest) => submitFilePassword(fileId, request),
    onSuccess: () => {
      setPassword('')
      setErrorMessage('')
      onSuccess?.()
      onClose()
    },
    onError: (error: unknown) => {
      setErrorMessage(extractMessage(error))
    },
  })

  const submit = () => {
    if (!password) {
      setErrorMessage('Password is required.')
      return
    }

    mutation.mutate({ password })
  }

  if (!open) {
    return null
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/70 p-4">
      <div className="w-full max-w-lg rounded-xl border border-slate-700 bg-slate-900 p-5">
        <h3 className="mb-3 text-lg font-semibold text-white">Archive password required</h3>
        <p className="mb-4 text-sm text-slate-300">
          This file is protected by a password. Enter it below to retry extraction.
        </p>
        <Input
          type="password"
          value={password}
          placeholder="Archive password"
          onChange={(event) => setPassword(event.target.value)}
          disabled={mutation.isPending}
        />
        <div className="mt-4 flex justify-end gap-2">
          <Button tone="ghost" onClick={onClose} disabled={mutation.isPending}>
            Cancel
          </Button>
          <Button tone="primary" onClick={submit} disabled={mutation.isPending || !password}>
            {mutation.isPending ? 'Submitting…' : 'Submit password'}
          </Button>
        </div>
        {errorMessage ? <p className="mt-3 text-sm text-rose-300">{errorMessage}</p> : null}
      </div>
    </div>
  )
}
