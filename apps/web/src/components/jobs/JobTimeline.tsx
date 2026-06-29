import { useMemo } from 'react'
import type { JobEvent } from '@/types/job'
import { parseTime } from '@/utils/query'

type JobTimelineProps = {
  events: JobEvent[]
  isLoading?: boolean
}

const isEmptyObject = (value: unknown) =>
  typeof value === 'object' && value !== null && !Array.isArray(value) && Object.keys(value as Record<string, unknown>).length === 0

const formatEventDetails = (details: Record<string, unknown>) => {
  if (!details || isEmptyObject(details)) return null
  return JSON.stringify(details, null, 2)
}

export const JobTimeline = ({ events, isLoading }: JobTimelineProps) => {
  const timelineEvents = useMemo(() => {
    return [...events].sort((a, b) => {
      const aTime = Date.parse(a.created_at || '')
      const bTime = Date.parse(b.created_at || '')
      return aTime - bTime
    })
  }, [events])

  if (isLoading) {
    return <div className="text-slate-400">Loading timeline…</div>
  }

  if (timelineEvents.length === 0) {
    return <div className="text-slate-400">No job events yet.</div>
  }

  return (
    <ol className="space-y-3">
      {timelineEvents.map((event) => (
        <li key={event.id} className="relative pl-4">
          <span className="absolute left-0 top-1 h-2.5 w-2.5 rounded-full bg-sky-400 ring-2 ring-sky-300/40" />
          <div className="flex flex-wrap items-start justify-between gap-2">
            <div>
              <div className="text-sm font-semibold text-slate-100">{event.event_type}</div>
              {event.event_message ? <div className="text-sm text-slate-300">{event.event_message}</div> : null}
            </div>
            <div className="text-xs text-slate-400">{parseTime(event.created_at)}</div>
          </div>
          {event.event_details ? <pre className="mt-2 max-h-44 overflow-auto text-xs text-slate-300">{formatEventDetails(event.event_details)}</pre> : null}
        </li>
      ))}
    </ol>
  )
}
