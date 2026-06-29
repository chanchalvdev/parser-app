import { apiGet, apiPost } from './api'
import type {
  UploadCompleteRequest,
  UploadCompleteResponse,
  UploadInitiateRequest,
  UploadInitiateResponse,
} from '@/types/domain'

export const initiateUpload = async (
  payload: UploadInitiateRequest,
): Promise<UploadInitiateResponse> => {
  return apiPost<UploadInitiateResponse>('/uploads/initiate', payload)
}

export const uploadFileToPresignedUrl = async (url: string, file: File): Promise<void> => {
  const response = await fetch(url, {
    method: 'PUT',
    body: file,
  })
  if (!response.ok) {
    throw new Error(`upload to storage failed: ${response.status} ${response.statusText}`)
  }
}

export const completeUpload = async (
  payload: UploadCompleteRequest,
): Promise<UploadCompleteResponse> => {
  return apiPost<UploadCompleteResponse>('/uploads/complete', payload)
}

export const getUpload = async (uploadId: string): Promise<unknown> => {
  return apiGet<unknown>(`/uploads/${uploadId}`)
}
