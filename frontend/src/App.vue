<script setup lang="ts">
import { onMounted, computed } from 'vue'
import { useProcurement } from './composables/useProcurement'
import FilterBar from './components/FilterBar.vue'
import SummaryCards from './components/SummaryCards.vue'
import CategoryChart from './components/CategoryChart.vue'
import TopProcurements from './components/TopProcurements.vue'
import DataTable from './components/DataTable.vue'

const {
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
} = useProcurement()

// When a single KLDI is selected, break down by satuanKerja; otherwise by KLDI.
const chartData = computed(() => {
  if (!summary.value) return []
  return query.kldi ? summary.value.bySatker : summary.value.byKldi
})

const chartTitle = computed(() =>
  query.kldi ? `Pagu anggaran per satuan kerja — ${query.kldi}` : 'Pagu anggaran per wilayah',
)

let searchTimeout: ReturnType<typeof setTimeout>
function onSearchUpdate(val: string) {
  clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    query.search = val
    query.page = 1
  }, 300)
}

function onFilterChange(key: 'kldi' | 'jenisPengadaan' | 'metode', val: string) {
  query[key] = val
  query.page = 1
}

onMounted(() => {
  loadSummary()
  loadFilters()
  loadData()
})
</script>

<template>
  <div class="min-h-screen bg-stone-100">
    <div class="max-w-7xl mx-auto px-4 py-8">
      <h1 class="text-3xl font-bold text-stone-800 mb-1">
        Anggaran Pengadaan Sulawesi Tengah 2026
      </h1>
      <p class="text-stone-500 mb-6">
        Dashboard pengadaan barang dan jasa pemerintah provinsi dan kabupaten/kota di Sulawesi Tengah.
      </p>

      <div v-if="error" class="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg text-red-700 text-sm">
        {{ error }}
      </div>

      <FilterBar
        :filters="filters"
        :kldi="query.kldi"
        :jenis-pengadaan="query.jenisPengadaan"
        :metode="query.metode"
        :search="query.search"
        :page-size="query.pageSize"
        @update:kldi="onFilterChange('kldi', $event)"
        @update:jenis-pengadaan="onFilterChange('jenisPengadaan', $event)"
        @update:metode="onFilterChange('metode', $event)"
        @update:search="onSearchUpdate"
        @update:page-size="query.pageSize = $event; query.page = 1"
      />

      <div class="mt-6">
        <SummaryCards :summary="summary" />
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-5 gap-6 mt-6">
        <div class="lg:col-span-3 bg-white rounded-xl border border-stone-200 p-5">
          <CategoryChart
            v-if="chartData.length"
            :data="chartData"
            :title="chartTitle"
            :max-items="10"
          />
        </div>
        <div class="lg:col-span-2 bg-white rounded-xl border border-stone-200 p-5">
          <TopProcurements :items="summary?.topItems ?? []" />
        </div>
      </div>

      <div class="mt-6 bg-white rounded-xl border border-stone-200 p-5">
        <DataTable
          :result="result"
          :loading="loading"
          :sort-by="query.sortBy"
          :sort-dir="query.sortDir"
          @sort="setSort"
          @page="setPage"
        />
      </div>

      <footer class="mt-8 text-xs text-stone-400">
        <p><strong>Sumber:</strong> SIRUP LKPP (sirup.inaproc.id) — Data RUP Sulawesi Tengah 2026</p>
      </footer>
    </div>
  </div>
</template>
