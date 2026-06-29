import { useSearchParams } from 'react-router-dom'

export const usePagination = () => {
  const [searchParams, setSearchParams] = useSearchParams()

  const page = Number(searchParams.get('page') || '1')
  const pageSize = Number(searchParams.get('page_size') || '25')

  const setPage = (nextPage: number) => {
    const params = new URLSearchParams(searchParams)
    params.set('page', String(nextPage))
    setSearchParams(params)
  }

  const setPageSize = (nextSize: number) => {
    const params = new URLSearchParams(searchParams)
    params.set('page_size', String(nextSize))
    params.set('page', '1')
    setSearchParams(params)
  }

  return {
    page,
    pageSize,
    setPage,
    setPageSize,
  }
}
