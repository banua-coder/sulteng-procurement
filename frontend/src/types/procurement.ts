export interface Procurement {
  id: number
  paket: string
  pagu: number
  jenisPengadaan: string
  metode: string
  pemilihan: string
  satuanKerja: string
  kldi: string
  lokasi: string
  sumberDana: string
  isPDN: boolean
  isUMK: boolean
}

export interface CategoryTotal {
  name: string
  total: number
  count: number
}

export interface Summary {
  totalPagu: number
  totalPaket: number
  jenisCount: number
  kldiCount: number
  topKldi: string
  byJenis: CategoryTotal[]
  byKldi: CategoryTotal[]
  byMetode: CategoryTotal[]
}

export interface PaginatedResult {
  data: Procurement[]
  total: number
  page: number
  pageSize: number
  totalPages: number
}

export interface Filters {
  kldi: string[]
  jenisPengadaan: string[]
  metode: string[]
}

export interface QueryParams {
  page: number
  pageSize: number
  search: string
  kldi: string
  jenisPengadaan: string
  metode: string
  sortBy: string
  sortDir: string
}
