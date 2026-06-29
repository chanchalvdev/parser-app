import { apiClient } from './apiClient'
import type { SearchRequest, SearchResponse, SearchSuggestions } from '@/types/search'

export const searchRecords = async (request: SearchRequest): Promise<SearchResponse> => {
  const safe: SearchRequest = {
    ...request,
    page_size: request.page_size ?? 25,
    page: request.page ?? 1,
  }

  if (safe.page_size > 100) {
    safe.page_size = 100
  }

  return apiClient.post('/search', safe)
}

export const getSearch = async (request: SearchRequest): Promise<SearchResponse> => {
  return apiClient.get<SearchResponse>('/search', {
    query: {
      ...request,
      page_size: request.page_size,
      page: request.page,
    },
  })
}

export const getSearchSuggestions = async (request: SearchRequest): Promise<SearchSuggestions> => {
  return apiClient.get<SearchSuggestions>('/search/suggestions', {
    query: request,
  })
}
