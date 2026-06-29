import { Bar, BarChart, CartesianGrid, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts'
import { Card } from '@/components/ui/Card'
import type { DashboardErrorBreakdown } from '@/types/dashboard'

type ErrorBreakdownChartProps = {
  data?: DashboardErrorBreakdown
  isLoading?: boolean
}

export const ErrorBreakdownChart = ({ data, isLoading }: ErrorBreakdownChartProps) => {
  const points = data?.errors || []

  return (
    <Card title="Error breakdown" subtitle="Top error categories">
      {isLoading ? (
        <p className="text-sm text-slate-300">Loading error breakdown…</p>
      ) : points.length === 0 ? (
        <p className="text-sm text-slate-300">No error data available yet.</p>
      ) : (
        <>
          <div className="h-72">
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={points} layout="vertical" margin={{ left: 30 }}>
                <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
                <XAxis type="number" stroke="#94a3b8" />
                <YAxis type="category" dataKey="error_code" width={140} stroke="#94a3b8" />
                <Tooltip
                  formatter={(v: number) => [String(v), 'Count']}
                  contentStyle={{ background: '#0f172a', borderColor: '#475569', color: '#f8fafc' }}
                />
                <Bar dataKey="count" fill="#f97316" radius={[0, 4, 4, 0]} />
              </BarChart>
            </ResponsiveContainer>
          </div>
          <p className="mt-2 text-xs text-slate-400">Total errors: {data?.total || 0}</p>
        </>
      )}
    </Card>
  )
}
