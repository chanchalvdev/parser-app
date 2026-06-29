import { SearchFacets as SearchFacetsType } from '@/types/search'
import { SearchFiltersValues } from '@/components/search/SearchFilters'
import { Link } from 'react-router-dom'

export type SearchFacetUpdate = {
  recordType?: string
  detectedFileType?: string
  ip?: string
  email?: string
  domain?: string
}

type SearchFacetsProps = {
  facets?: SearchFacetsType
  filters: SearchFiltersValues
  onFacetUpdate: (update: SearchFacetUpdate) => void
}

const FacetSection = ({
  title,
  items,
  selected,
  onSelect,
}: {
  title: string
  items: Array<{ value: string; count: number }>
  selected?: string
  onSelect: (value: string) => void
}) => {
  if (!items || items.length === 0) return null

  return (
    <section>
      <h3 className="mb-2 text-sm font-semibold text-slate-100">{title}</h3>
      <div className="flex flex-wrap gap-2">
        {items.map((item) => {
          const isActive = selected === item.value
          return (
            <button
              type="button"
              key={`${title}-${item.value}`}
              onClick={() => onSelect(item.value)}
              className={`rounded-full border px-3 py-1 text-xs ${isActive ? 'border-blue-300 bg-blue-900/40 text-blue-100' : 'border-slate-500/80 text-slate-200'}`}
            >
              {item.value} · {item.count}
            </button>
          )
        })}
      </div>
    </section>
  )
}

export const SearchFacets = ({ facets, filters, onFacetUpdate }: SearchFacetsProps) => {
  if (!facets) {
    return (
      <section className="panel p-4">
        <h3 className="panel-title mb-1">Facets</h3>
        <p className="text-sm text-slate-400">No facets yet. Run a search to populate.</p>
      </section>
    )
  }

  return (
    <section className="panel p-4 space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="panel-title">Facets</h3>
        <Link to="/search" className="text-xs text-blue-300 hover:underline">
          View all
        </Link>
      </div>

      <FacetSection
        title="Record type"
        items={facets.record_type}
        selected={filters.recordType}
        onSelect={(value) =>
          onFacetUpdate({
            recordType: filters.recordType === value ? '' : value,
          })
        }
      />

      <FacetSection
        title="File type"
        items={facets.detected_file_type}
        selected={filters.detectedFileType}
        onSelect={(value) =>
          onFacetUpdate({
            detectedFileType: filters.detectedFileType === value ? '' : value,
          })
        }
      />

      <FacetSection
        title="IP addresses"
        items={facets.entities.ip_addresses}
        selected={filters.ip}
        onSelect={(value) =>
          onFacetUpdate({
            ip: filters.ip === value ? '' : value,
          })
        }
      />

      <FacetSection
        title="Emails"
        items={facets.entities.emails}
        selected={filters.email}
        onSelect={(value) =>
          onFacetUpdate({
            email: filters.email === value ? '' : value,
          })
        }
      />

      <FacetSection
        title="Domains"
        items={facets.entities.domains}
        selected={filters.domain}
        onSelect={(value) =>
          onFacetUpdate({
            domain: filters.domain === value ? '' : value,
          })
        }
      />

      {facets.entities.urls.length > 0 ? (
        <FacetSection
          title="URLs"
          items={facets.entities.urls}
          onSelect={() => undefined}
        />
      ) : null}

      {facets.entities.hashes.length > 0 ? (
        <FacetSection
          title="Hashes"
          items={facets.entities.hashes}
          onSelect={() => undefined}
        />
      ) : null}

      {facets.record_type.length + facets.detected_file_type.length + facets.entities.ip_addresses.length + facets.entities.emails.length + facets.entities.domains.length + facets.entities.urls.length + facets.entities.hashes.length === 0 ? (
        <p className="text-sm text-slate-400">No facet buckets available.</p>
      ) : null}
    </section>
  )
}
