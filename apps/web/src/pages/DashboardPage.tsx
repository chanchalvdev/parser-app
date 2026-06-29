import { useUiStore } from '@/stores/uiStore'
import { useDashboardSummary, useDashboardCharts } from '@/hooks/apiHooks'
import { PageHeader } from '@/components/layout/PageHeader'
import { SummaryCards } from '@/components/dashboard/SummaryCards'
import { FileTypeChart } from '@/components/dashboard/FileTypeChart'
import { ProcessingStatusChart } from '@/components/dashboard/ProcessingStatusChart'
import { UploadVolumeChart } from '@/components/dashboard/UploadVolumeChart'
import { ErrorBreakdownChart } from '@/components/dashboard/ErrorBreakdownChart'
import { TopEntitiesTable } from '@/components/dashboard/TopEntitiesTable'
import { ProcessingDurationCard } from '@/components/dashboard/ProcessingDurationCard'
import { RecentFailuresTable } from '@/components/dashboard/RecentFailuresTable'

export const DashboardPage = () => {
  const tenantId = useUiStore((state) => state.tenantId)

  const summaryQuery = useDashboardSummary(tenantId)
  const charts = useDashboardCharts(tenantId, {
    fileTypeLimit: 10,
    statusLimit: 10,
    volumeDays: 14,
    errorLimit: 10,
    entitiesLimit: 20,
  })

  const isAnyError =
    summaryQuery.isError || charts.isError

  return (
    <div className="space-y-4">
      <PageHeader title="Dashboard" subtitle="Operational summary and quality metrics" />

      {isAnyError ? <p className="text-rose-300">Some dashboard queries failed to load.</p> : null}

      <SummaryCards summary={summaryQuery.data} isLoading={summaryQuery.isLoading} />

      <div className="grid gap-4 lg:grid-cols-2">
        <FileTypeChart data={charts.fileTypes} isLoading={charts.isLoading} />
        <ProcessingStatusChart data={charts.processingStatus} isLoading={charts.isLoading} />
      </div>

      <UploadVolumeChart data={charts.uploadVolume} isLoading={charts.isLoading} />

      <div className="grid gap-4 lg:grid-cols-2">
        <ErrorBreakdownChart data={charts.errorBreakdown} isLoading={charts.isLoading} />
        <TopEntitiesTable data={charts.entities} isLoading={charts.isLoading} />
      </div>

      <div className="grid gap-4 lg:grid-cols-2">
        <ProcessingDurationCard data={charts.duration} isLoading={charts.isLoading} />
        <RecentFailuresTable breakdown={charts.errorBreakdown} isLoading={charts.isLoading} />
      </div>
    </div>
  )
}
