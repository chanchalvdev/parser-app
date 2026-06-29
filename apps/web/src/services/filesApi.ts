import { apiClient } from './apiClient'
import type {
  FileChildrenResponse,
  FileItem,
  FileListResponse,
  FileRecordsResponse,
  FileTreeNode,
  ListFilesParams,
  SubmitFilePasswordRequest,
  SubmitFilePasswordResponse,
} from '@/types/file'

export const listFiles = async (args: ListFilesParams = {}): Promise<FileListResponse> => {
  const page = args.page
  const page_size = args.page_size ?? args.pageSize

  return apiClient.get<FileListResponse>('/files', {
    query: {
      tenant_id: args.tenant_id ?? args.tenantId,
      status: args.status,
      extension: args.extension,
      detected_file_type: args.detected_file_type ?? args.detectedFileType,
      page,
      page_size,
    },
  })
}

export const getFile = async (fileId: string): Promise<FileItem> => {
  return apiClient.get<FileItem>(`/files/${fileId}`)
}

export const getFileChildren = async (
  fileId: string,
  page = 1,
  pageSize = 25,
): Promise<FileChildrenResponse> => {
  return apiClient.get<FileChildrenResponse>(`/files/${fileId}/children`, {
    query: {
      page,
      page_size: pageSize,
    },
  })
}

export const getFileTree = async (fileId: string): Promise<FileTreeNode> => {
  return apiClient.get<FileTreeNode>(`/files/${fileId}/tree`)
}

export const getFileRecords = async (
  fileId: string,
  page = 1,
  pageSize = 25,
): Promise<FileRecordsResponse> => {
  return apiClient.get<FileRecordsResponse>(`/files/${fileId}/records`, {
    query: {
      page,
      page_size: pageSize,
    },
  })
}

export const submitFilePassword = async (
  fileId: string,
  request: SubmitFilePasswordRequest,
): Promise<SubmitFilePasswordResponse> => {
  return apiClient.post<SubmitFilePasswordResponse>(`/files/${fileId}/password`, request)
}
