import type { Summary, PaginatedResult, Filters, QueryParams } from '../types/procurement'

const BASE = '/api/v1'

async function fetchJSON<T>(url: string): Promise<T> {
  const res = await fetch(url)
  if (!res.ok) throw new Error(`API error: ${res.status}`)
  return res.json()
}

export function getSummary(): Promise<Summary> {
  return fetchJSON(`${BASE}/summary`)
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
