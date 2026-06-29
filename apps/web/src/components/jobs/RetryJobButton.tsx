import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { retryJob } from '@/services/jobsApi'
import { Button } from '@/components/ui/Button'

type RetryJobButtonProps = {
  jobId: string
  disabled?: boolean
  label?: string
  onRetrySuccess?: () => void
}

export const RetryJobButton = ({ jobId, disabled, label = 'Retry', onRetrySuccess }: RetryJobButtonProps) => {
  const queryClient = useQueryClient()
  const [errorMessage, setErrorMessage] = useState('')

  const mutation = useMutation({
    mutationFn: () => retryJob(jobId),
    onSuccess: () => {
      setErrorMessage('')
      queryClient.invalidateQueries({ queryKey: ['job', jobId] })
      queryClient.invalidateQueries({ queryKey: ['jobs'] })
      queryClient.invalidateQueries({ queryKey: ['jobs', jobId, 'events'] })
      onRetrySuccess?.()
    },
    onError: (error: unknown) => {
      setErrorMessage(error instanceof Error ? error.message : 'Failed to retry job.')
    },
  })

  return (
    <div className="space-y-1">
      <Button tone="danger" disabled={disabled || mutation.isPending} onClick={() => mutation.mutate()}>
        {mutation.isPending ? 'Retrying…' : label}
      </Button>
      {errorMessage ? <p className="text-xs text-rose-300">{errorMessage}</p> : null}
    </div>
  )
}
