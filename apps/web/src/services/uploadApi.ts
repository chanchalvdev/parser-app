import { apiClient } from './apiClient'
import type {
  UploadCompleteRequest,
  UploadCompleteResponse,
  UploadInitiateRequest,
  UploadInitiateResponse,
} from '@/types/file'

export type UploadProgress = {
  loaded: number
  total: number
  percentage: number
}

export const initiateUpload = async (
  payload: UploadInitiateRequest,
): Promise<UploadInitiateResponse> => {
  return apiClient.post('/uploads/initiate', payload)
}

export const uploadFileToPresignedUrl = async (
  url: string,
  file: File,
  onProgress?: (progress: UploadProgress) => void,
): Promise<void> => {
  await new Promise<void>((resolve, reject) => {
    const request = new XMLHttpRequest()

    request.open('PUT', url, true)
    request.timeout = 600_000

    request.upload.onprogress = (event) => {
      if (!event.lengthComputable || event.total === 0) {
        return
      }

      onProgress?.({
        loaded: event.loaded,
        total: event.total,
        percentage: Math.round((event.loaded / event.total) * 100),
      })
    }

    request.onload = () => {
      if (request.status >= 200 && request.status < 300) {
        resolve()
        return
      }

      reject(new Error(`upload to storage failed: ${request.status} ${request.statusText}`))
    }

    request.onerror = () => {
      reject(new Error('upload to storage failed: network error'))
    }

    request.onabort = () => {
      reject(new Error('upload to storage was canceled'))
    }

    request.ontimeout = () => {
      reject(new Error('upload to storage timed out'))
    }

    request.send(file)
  })
}

export const getUpload = async (uploadId: string): Promise<unknown> => {
  return apiClient.get<unknown>(`/uploads/${uploadId}`)
}

export const completeUpload = async (
  payload: UploadCompleteRequest,
): Promise<UploadCompleteResponse> => {
  return apiClient.post('/uploads/complete', payload)
}

export type { UploadInitiateRequest, UploadInitiateResponse, UploadCompleteRequest, UploadCompleteResponse } from '@/types/file'
