import { useQueries, useQuery } from '@tanstack/react-query'
import type { Query, UseQueryOptions } from '@tanstack/react-query'
import type { FileItem, FileRecordsResponse, FileTreeNode } from '@/types/file'
import { getDashboardErrorBreakdown, getDashboardEntities, getDashboardFileTypes, getDashboardProcessingDuration, getDashboardProcessingStatus, getDashboardSummary, getDashboardUploadVolume } from '@/services/dashboardApi'
import { getFile, getFileRecords, getFileTree, listFiles } from '@/services/filesApi'
import { getJob, getJobEvents, listJobs } from '@/services/jobsApi'
import { searchRecords } from '@/services/searchApi'
import type { SearchRequest, SearchResponse } from '@/types/search'
import type {
  DashboardDistribution,
  DashboardEntities,
  DashboardErrorBreakdown,
  DashboardProcessingDuration,
  DashboardSummary,
  DashboardUploadVolume,
} from '@/types/dashboard'
import type { JobEventsResponse, JobListResponse, IngestionJob } from '@/types/job'
import type { FileListResponse } from '@/types/file'

/** Statuses for which the job is no longer progressing on its own. */
export const TERMINAL_STATUSES = new Set<string>([
  'completed',
  'failed',
  'cancelled',
  'canceled',
])

/** Whether the job is still progressing and the UI should keep polling for live updates. */
export const isJobLive = (job: { status?: string | null } | null | undefined): boolean => {
  if (!job?.status) {
    return true
  }
  return !TERMINAL_STATUSES.has(job.status.toLowerCase())
}

/**
 * React Query refetchInterval that polls at `intervalMs` while the job is live
 * and stops as soon as it hits a terminal status.
 */
const liveRefetchInterval = <TData>(
  intervalMs: number,
) => (query: Query<TData, Error, TData, readonly unknown[]>): number | false => {
  const data = query.state.data as { status?: string | null } | undefined
  if (data && !isJobLive(data)) {
    return false
  }
  return intervalMs
}

export const useFiles = (
  params: {
    tenant_id?: string
    tenantId?: string
    status?: string
    extension?: string
    detected_file_type?: string
    detectedFileType?: string
    page?: number
    page_size?: number
    pageSize?: number
  },
  options: Omit<UseQueryOptions<FileListResponse>, 'queryKey' | 'queryFn'> = {},
) => {
  const tenantId = params.tenant_id ?? params.tenantId
  const status = params.status
  const extension = params.extension
  const detectedFileType = params.detected_file_type ?? params.detectedFileType
  const page = params.page ?? 1
  const pageSize = params.page_size ?? params.pageSize ?? 25

  return useQuery({
    queryKey: ['files', tenantId, status, extension, detectedFileType, page, pageSize],
    queryFn: () =>
      listFiles({
        tenant_id: tenantId,
        status,
        extension,
        detected_file_type: detectedFileType,
        page,
        page_size: pageSize,
      }),
    ...options,
  })
}

export const useFile = (
  fileId: string,
  options: Omit<UseQueryOptions<FileItem>, 'queryKey' | 'queryFn'> = {},
) =>
  useQuery({
    queryKey: ['file', fileId],
    enabled: !!fileId,
    queryFn: () => getFile(fileId),
    ...options,
  })

export const useFileTree = (
  fileId: string,
  options: Omit<UseQueryOptions<FileTreeNode>, 'queryKey' | 'queryFn'> = {},
) =>
  useQuery({
    queryKey: ['file-tree', fileId],
    enabled: !!fileId,
    queryFn: () => getFileTree(fileId),
    ...options,
  })

export const useFileRecords = (
  fileId: string,
  page = 1,
  pageSize = 25,
  options: Omit<UseQueryOptions<FileRecordsResponse>, 'queryKey' | 'queryFn'> = {},
) =>
  useQuery({
    queryKey: ['file-records', fileId, page, pageSize],
    enabled: !!fileId,
    queryFn: () => getFileRecords(fileId, page, pageSize),
    ...options,
  })

export const useJobs = (
  params: {
    tenant_id?: string
    tenantId?: string
    status?: string
    page?: number
    page_size?: number
    pageSize?: number
  },
  options: Omit<UseQueryOptions<JobListResponse>, 'queryKey' | 'queryFn'> = {},
) =>
  useQuery({
    queryKey: [
      'jobs',
      params.tenant_id ?? params.tenantId,
      params.status,
      params.page ?? 1,
      params.page_size ?? params.pageSize ?? 25,
    ],
    queryFn: () =>
      listJobs({
        tenant_id: params.tenant_id,
        tenantId: params.tenantId,
        status: params.status,
        page: params.page,
        page_size: params.page_size ?? params.pageSize,
      }),
    ...options,
  })

export const useJob = (
  jobId: string,
  options: Omit<UseQueryOptions<IngestionJob>, 'queryKey' | 'queryFn'> = {},
) =>
  useQuery({
    queryKey: ['job', jobId],
    enabled: !!jobId,
    queryFn: () => getJob(jobId),
    refetchInterval: liveRefetchInterval<IngestionJob>(1500),
    refetchIntervalInBackground: false,
    ...options,
  })

export const useJobEvents = (
  jobId: string,
  page = 1,
  pageSize = 25,
  options: Omit<UseQueryOptions<JobEventsResponse>, 'queryKey' | 'queryFn'> = {},
) =>
  useQuery({
    queryKey: ['jobs', jobId, 'events', page, pageSize],
    enabled: !!jobId,
    queryFn: () => getJobEvents(jobId, page, pageSize),
    refetchInterval: liveRefetchInterval<JobEventsResponse>(2000),
    refetchIntervalInBackground: false,
    ...options,
  })

export const useSearch = (
  request: SearchRequest,
  options: Omit<UseQueryOptions<SearchResponse>, 'queryKey' | 'queryFn'> = {},
) =>
  useQuery({
    queryKey: ['search', request],
    queryFn: () => searchRecords(request),
    ...options,
  })

export const useDashboardSummary = (
  tenantId?: string,
  options: Omit<UseQueryOptions<DashboardSummary>, 'queryKey' | 'queryFn'> = {},
) =>
  useQuery({
    queryKey: ['dashboard-summary', tenantId],
    queryFn: () => getDashboardSummary(tenantId),
    ...options,
  })

export type DashboardChartsData = {
  fileTypes?: DashboardDistribution
  processingStatus?: DashboardDistribution
  uploadVolume?: DashboardUploadVolume
  errorBreakdown?: DashboardErrorBreakdown
  entities?: DashboardEntities
  duration?: DashboardProcessingDuration
  isLoading: boolean
  isError: boolean
  refetch: () => Promise<{
    fileTypes: Awaited<ReturnType<typeof getDashboardFileTypes>>
    processingStatus: Awaited<ReturnType<typeof getDashboardProcessingStatus>>
    uploadVolume: Awaited<ReturnType<typeof getDashboardUploadVolume>>
    errorBreakdown: Awaited<ReturnType<typeof getDashboardErrorBreakdown>>
    entities: Awaited<ReturnType<typeof getDashboardEntities>>
    duration: Awaited<ReturnType<typeof getDashboardProcessingDuration>>
  }>
}

export const useDashboardCharts = (
  tenantId?: string,
  params: {
    fileTypeLimit?: number
    statusLimit?: number
    volumeDays?: number
    errorLimit?: number
    entitiesLimit?: number
  } = {},
): DashboardChartsData => {
  const fileTypeLimit = params.fileTypeLimit ?? 25
  const statusLimit = params.statusLimit ?? 25
  const volumeDays = params.volumeDays ?? 14
  const errorLimit = params.errorLimit ?? 10
  const entitiesLimit = params.entitiesLimit ?? 10

  const [fileTypesQ, statusQ, volumeQ, errorQ, entitiesQ, durationQ] = useQueries({
    queries: [
      {
        queryKey: ['dashboard-file-types', tenantId, fileTypeLimit],
        queryFn: () => getDashboardFileTypes(tenantId, fileTypeLimit),
      },
      {
        queryKey: ['dashboard-status', tenantId, statusLimit],
        queryFn: () => getDashboardProcessingStatus(tenantId, statusLimit),
      },
      {
        queryKey: ['dashboard-volume', tenantId, volumeDays],
        queryFn: () => getDashboardUploadVolume(tenantId, volumeDays),
      },
      {
        queryKey: ['dashboard-errors', tenantId, errorLimit],
        queryFn: () => getDashboardErrorBreakdown(tenantId, errorLimit),
      },
      {
        queryKey: ['dashboard-entities', tenantId, entitiesLimit],
        queryFn: () => getDashboardEntities(tenantId, entitiesLimit),
      },
      {
        queryKey: ['dashboard-duration', tenantId],
        queryFn: () => getDashboardProcessingDuration(tenantId),
      },
    ],
  })

  const isLoading =
    fileTypesQ.isLoading ||
    statusQ.isLoading ||
    volumeQ.isLoading ||
    errorQ.isLoading ||
    entitiesQ.isLoading ||
    durationQ.isLoading

  const isError =
    fileTypesQ.isError ||
    statusQ.isError ||
    volumeQ.isError ||
    errorQ.isError ||
    entitiesQ.isError ||
    durationQ.isError

  return {
    fileTypes: fileTypesQ.data,
    processingStatus: statusQ.data,
    uploadVolume: volumeQ.data,
    errorBreakdown: errorQ.data,
    entities: entitiesQ.data,
    duration: durationQ.data,
    isLoading,
    isError,
    refetch: async () => {
      const [fileTypes, processingStatus, uploadVolume, errorBreakdown, entities, duration] = await Promise.all([
        fileTypesQ.refetch(),
        statusQ.refetch(),
        volumeQ.refetch(),
        errorQ.refetch(),
        entitiesQ.refetch(),
        durationQ.refetch(),
      ])

      return {
        fileTypes: fileTypes.data as any,
        processingStatus: processingStatus.data as any,
        uploadVolume: uploadVolume.data as any,
        errorBreakdown: errorBreakdown.data as any,
        entities: entities.data as any,
        duration: duration.data as any,
      }
    },
  }
}
