import { apiGet, apiPost } from './api'
import type { SearchRequest, SearchResponse, SearchSuggestions } from '@/types/domain'

export const searchParsedRecords = async (request: SearchRequest): Promise<SearchResponse> => {
  const safe = { ...request }
  if (safe.page_size === 0) {
    safe.page_size = 25
  }
  if (safe.page === 0) {
    safe.page = 1
  }

  if (safe.page_size && safe.page_size > 100) {
    safe.page_size = 100
  }

  return apiPost<SearchResponse>('/search', safe)
}

export const searchParsedRecordsGet = async (request: SearchRequest): Promise<SearchResponse> => {
  return apiGet<SearchResponse>('/search', request)
}

export const getSearchSuggestions = async (request: SearchRequest): Promise<SearchSuggestions> => {
  return apiGet<SearchSuggestions>('/search/suggestions', request)
}
