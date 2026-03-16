import { ref, reactive, watch } from 'vue'
import { getSummary, getProcurements, getFilters } from '../api/procurement'
import type { Summary, PaginatedResult, Filters, QueryParams } from '../types/procurement'

export function useProcurement() {
  const summary = ref<Summary | null>(null)
  const result = ref<PaginatedResult | null>(null)
  const filters = ref<Filters | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  const query = reactive<QueryParams>({
    page: 1,
    pageSize: 25,
    search: '',
    kldi: '',
    jenisPengadaan: '',
    metode: '',
    sortBy: 'pagu',
    sortDir: 'desc',
  })

  async function loadSummary() {
    try {
      summary.value = await getSummary()
    } catch (e: unknown) {
      error.value = e instanceof Error ? e.message : String(e)
    }
  }

  async function loadFilters() {
    try {
      filters.value = await getFilters()
    } catch (e: unknown) {
      error.value = e instanceof Error ? e.message : String(e)
    }
  }

  async function loadData() {
    loading.value = true
    error.value = null
    try {
      result.value = await getProcurements({ ...query })
    } catch (e: unknown) {
      error.value = e instanceof Error ? e.message : String(e)
    } finally {
      loading.value = false
    }
  }

  function setPage(page: number) {
    query.page = page
  }

  function setSort(sortBy: string) {
    if (query.sortBy === sortBy) {
      query.sortDir = query.sortDir === 'desc' ? 'asc' : 'desc'
    } else {
      query.sortBy = sortBy
      query.sortDir = 'desc'
    }
  }

  function resetFilters() {
    query.page = 1
    query.search = ''
    query.kldi = ''
    query.jenisPengadaan = ''
    query.metode = ''
  }

  watch(query, () => {
    loadData()
  })

  return {
    summary,
    result,
    filters,
    loading,
    error,
    query,
    loadSummary,
    loadFilters,
    loadData,
    setPage,
    setSort,
    resetFilters,
  }
}
