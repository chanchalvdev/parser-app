import { apiGet } from './api'
import type {
  FileListResponse,
  FileItem,
  FileChildrenResponse,
  FileTreeNode,
  FileRecordsResponse,
} from '@/types/domain'

export const listFiles = async (
  args: {
    tenantId?: string
    status?: string
    extension?: string
    detectedFileType?: string
    page?: number
    pageSize?: number
  },
): Promise<FileListResponse> => {
  return apiGet<FileListResponse>('/files', {
    tenant_id: args.tenantId,
    status: args.status,
    extension: args.extension,
    detected_file_type: args.detectedFileType,
    page: args.page,
    page_size: args.pageSize,
  })
}

export const getFile = async (id: string): Promise<FileItem> => {
  return apiGet<FileItem>(`/files/${id}`)
}

export const getFileChildren = async (
  id: string,
  page = 1,
  pageSize = 25,
): Promise<FileChildrenResponse> => {
  return apiGet<FileChildrenResponse>(`/files/${id}/children`, {
    page,
    page_size: pageSize,
  })
}

export const getFileTree = async (id: string): Promise<FileTreeNode> => {
  return apiGet<FileTreeNode>(`/files/${id}/tree`)
}

export const getFileRecords = async (
  id: string,
  page = 1,
  pageSize = 25,
): Promise<FileRecordsResponse> => {
  return apiGet<FileRecordsResponse>(`/files/${id}/records`, {
    page,
    page_size: pageSize,
  })
}
