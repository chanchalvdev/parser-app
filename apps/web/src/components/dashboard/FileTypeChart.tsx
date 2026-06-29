import {
  Bar,
  BarChart,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts'
import { Card } from '@/components/ui/Card'
import type { DashboardDistribution } from '@/types/dashboard'

type FileTypeChartProps = {
  data?: DashboardDistribution
  isLoading?: boolean
}

export const FileTypeChart = ({ data, isLoading }: FileTypeChartProps) => {
  const buckets = data?.buckets || []

  return (
    <Card title="File types" subtitle="Records by detected file type">
      {isLoading ? (
        <p className="text-sm text-slate-300">Loading file type stats…</p>
      ) : buckets.length === 0 ? (
        <p className="text-sm text-slate-300">No file-type data available yet.</p>
      ) : (
        <div className="h-72">
          <ResponsiveContainer width="100%" height="100%">
            <BarChart data={buckets}>
              <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
              <XAxis dataKey="value" stroke="#94a3b8" />
              <YAxis stroke="#94a3b8" />
              <Tooltip
                formatter={(v: number) => [`${v}`, 'Count']}
                contentStyle={{ background: '#0f172a', borderColor: '#475569', color: '#f8fafc' }}
              />
              <Bar dataKey="count" fill="#38bdf8" radius={4} />
            </BarChart>
          </ResponsiveContainer>
        </div>
      )}
    </Card>
  )
}
