import {
  Cell,
  Pie,
  PieChart,
  ResponsiveContainer,
  Tooltip,
} from 'recharts'
import { Card } from '@/components/ui/Card'
import type { DashboardDistribution } from '@/types/dashboard'

type ProcessingStatusChartProps = {
  data?: DashboardDistribution
  isLoading?: boolean
}

const palette = ['#38bdf8', '#f472b6', '#34d399', '#f59e0b', '#a78bfa', '#f97316', '#fb7185', '#22d3ee']

export const ProcessingStatusChart = ({ data, isLoading }: ProcessingStatusChartProps) => {
  const buckets = data?.buckets || []

  return (
    <Card title="Processing status" subtitle="Current file/job state breakdown">
      {isLoading ? (
        <p className="text-sm text-slate-300">Loading status breakdown…</p>
      ) : buckets.length === 0 ? (
        <p className="text-sm text-slate-300">No status data available yet.</p>
      ) : (
        <div className="h-72">
          <ResponsiveContainer width="100%" height="100%">
            <PieChart>
              <Pie
                data={buckets}
                dataKey="count"
                nameKey="value"
                outerRadius={110}
                label
                labelLine={false}
              >
                {buckets.map((entry, index) => (
                  <Cell key={entry.value} fill={palette[index % palette.length]} />
                ))}
              </Pie>
              <Tooltip
                formatter={(v: number) => [String(v), 'Count']}
                contentStyle={{ background: '#0f172a', borderColor: '#475569', color: '#f8fafc' }}
              />
            </PieChart>
          </ResponsiveContainer>
          <ul className="mt-3 space-y-1 text-xs">
            {buckets.map((item, index) => (
              <li key={item.value} className="flex justify-between gap-2 text-slate-200">
                <span style={{ color: palette[index % palette.length] }}>{item.value || 'unknown'}</span>
                <span>{item.count}</span>
              </li>
            ))}
          </ul>
        </div>
      )}
    </Card>
  )
}
