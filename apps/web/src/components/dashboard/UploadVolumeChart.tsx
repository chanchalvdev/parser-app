import {
  Area,
  AreaChart,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts'
import { Card } from '@/components/ui/Card'
import type { DashboardUploadVolume } from '@/types/dashboard'

type UploadVolumeChartProps = {
  data?: DashboardUploadVolume
  isLoading?: boolean
}

const makeSeries = (data?: DashboardUploadVolume) =>
  (data?.buckets || []).map((bucket) => ({
    bucket: bucket.bucket,
    uploads: bucket.uploads,
    files: bucket.files,
    parsed_records: bucket.parsed_records,
  }))

export const UploadVolumeChart = ({ data, isLoading }: UploadVolumeChartProps) => {
  const series = makeSeries(data)

  return (
    <Card title="Upload volume" subtitle={
      data ? `Uploads/files/parsed records (${data.from} → ${data.to})` : 'Upload volume over time'
    }>
      {isLoading ? (
        <p className="text-sm text-slate-300">Loading upload volume…</p>
      ) : series.length === 0 ? (
        <p className="text-sm text-slate-300">No volume data available yet.</p>
      ) : (
        <div className="h-72">
          <ResponsiveContainer width="100%" height="100%">
            <AreaChart data={series}>
              <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
              <XAxis dataKey="bucket" stroke="#94a3b8" />
              <YAxis stroke="#94a3b8" />
              <Tooltip
                contentStyle={{ background: '#0f172a', borderColor: '#475569', color: '#f8fafc' }}
              />
              <Area type="monotone" dataKey="uploads" stroke="#38bdf8" fill="#0369a1" />
              <Area type="monotone" dataKey="files" stroke="#34d399" fill="#166534" />
              <Area type="monotone" dataKey="parsed_records" stroke="#f472b6" fill="#be185d" />
            </AreaChart>
          </ResponsiveContainer>
        </div>
      )}
    </Card>
  )
}
