import type { Summary, PaginatedResult, Filters, QueryParams, RealisasiSummary, JoinedRecord } from '../types/procurement'

const BASE = '/api/v1'

async function fetchJSON<T>(url: string): Promise<T> {
  const res = await fetch(url)
  if (!res.ok) throw new Error(`API error: ${res.status}`)
  return res.json()
}

export function getSummary(params?: Partial<Pick<QueryParams, 'kldi' | 'jenisPengadaan' | 'metode' | 'search'>>): Promise<Summary> {
  const qs = new URLSearchParams()
  if (params) {
    for (const [key, value] of Object.entries(params)) {
      if (value) qs.set(key, value)
    }
  }
  const query = qs.toString()
  return fetchJSON(`${BASE}/summary${query ? `?${query}` : ''}`)
}

export function getProcurements(params: Partial<QueryParams>): Promise<PaginatedResult> {
  const qs = new URLSearchParams()
  for (const [key, value] of Object.entries(params)) {
    if (value !== undefined && value !== '') {
      qs.set(key, String(value))
    }
  }
  return fetchJSON(`${BASE}/procurements?${qs}`)
}

export function getFilters(): Promise<Filters> {
  return fetchJSON(`${BASE}/filters`)
}

export function getRealisasiSummary(): Promise<RealisasiSummary> {
  return fetchJSON(`${BASE}/realisasi/summary`)
}

export function getRealisasi(): Promise<JoinedRecord[]> {
  return fetchJSON(`${BASE}/realisasi`)
}
