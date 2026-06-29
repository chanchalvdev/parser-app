export type SearchResult = {
  record_id: string
  file_id: string
  job_id: string
  source_file_name: string
  record_type: string
  content_preview: string
  highlight: string
  entities: Record<string, unknown>
  created_at: string
}

export type SearchFacetItem = {
  value: string
  count: number
}

export type SearchFacets = {
  record_type: SearchFacetItem[]
  detected_file_type: SearchFacetItem[]
  entities: {
    ip_addresses: SearchFacetItem[]
    emails: SearchFacetItem[]
    domains: SearchFacetItem[]
    urls: SearchFacetItem[]
    hashes: SearchFacetItem[]
  }
}

export type SearchResponse = {
  total: number
  page: number
  page_size: number
  results: SearchResult[]
  facets: SearchFacets
}

export type SearchSuggestions = {
  suggestions: {
    record_type: SearchFacetItem[]
    detected_file_type: SearchFacetItem[]
    entities: {
      ip_addresses: SearchFacetItem[]
      emails: SearchFacetItem[]
      domains: SearchFacetItem[]
      urls: SearchFacetItem[]
      hashes: SearchFacetItem[]
    }
  }
}

export type SearchRequest = {
  q?: string
  file_id?: string
  extension?: string
  detected_file_type?: string
  record_type?: string
  date_from?: string
  date_to?: string
  ip?: string
  email?: string
  domain?: string
  job_id?: string
  page?: number
  page_size?: number
  sort?: 'relevance' | 'created_at'
  tenant_id?: string
}
