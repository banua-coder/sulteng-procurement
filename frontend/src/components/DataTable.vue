<script setup lang="ts">
import type { PaginatedResult } from '../types/procurement'

defineProps<{
  result: PaginatedResult | null
  loading: boolean
  sortBy: string
  sortDir: string
}>()

const emit = defineEmits<{
  sort: [field: string]
  page: [page: number]
}>()

function formatRupiah(value: number): string {
  return `Rp${value.toLocaleString('id-ID')}`
}

function sortIcon(field: string, currentSort: string, dir: string): string {
  if (field !== currentSort) return ''
  return dir === 'desc' ? ' ▼' : ' ▲'
}
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-2">
      <h3 class="font-semibold text-stone-800">Tabel detail</h3>
      <div class="text-sm text-stone-500" v-if="result">
        Menampilkan {{ result.data.length }} dari {{ result.total.toLocaleString('id-ID') }} data
        <span class="ml-4">Halaman {{ result.page }} dari {{ result.totalPages }}</span>
      </div>
    </div>

    <div class="overflow-x-auto rounded-lg border border-stone-200">
      <table class="min-w-full text-sm">
        <thead class="bg-stone-50 text-stone-600">
          <tr>
            <th
              v-for="col in [
                { key: 'kldi', label: 'Wilayah' },
                { key: 'satuanKerja', label: 'Satuan Kerja' },
                { key: 'paket', label: 'Paket' },
                { key: 'jenisPengadaan', label: 'Jenis' },
                { key: 'metode', label: 'Metode' },
                { key: 'pagu', label: 'Pagu' },
              ]"
              :key="col.key"
              class="px-3 py-2 text-left cursor-pointer hover:bg-stone-100 whitespace-nowrap"
              @click="emit('sort', col.key)"
            >
              {{ col.label }}{{ sortIcon(col.key, sortBy, sortDir) }}
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="loading">
            <td colspan="6" class="px-3 py-8 text-center text-stone-400">Memuat data...</td>
          </tr>
          <tr v-else-if="!result?.data.length">
            <td colspan="6" class="px-3 py-8 text-center text-stone-400">Tidak ada data</td>
          </tr>
          <tr
            v-for="item in result?.data"
            :key="item.id"
            class="border-t border-stone-100 hover:bg-stone-50"
          >
            <td class="px-3 py-2">{{ item.kldi }}</td>
            <td class="px-3 py-2">{{ item.satuanKerja }}</td>
            <td class="px-3 py-2 max-w-xs truncate">{{ item.paket }}</td>
            <td class="px-3 py-2 whitespace-nowrap">{{ item.jenisPengadaan }}</td>
            <td class="px-3 py-2 whitespace-nowrap">{{ item.metode }}</td>
            <td class="px-3 py-2 text-right whitespace-nowrap font-mono">{{ formatRupiah(item.pagu) }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <div class="flex justify-end gap-2 mt-3" v-if="result && result.totalPages > 1">
      <button
        class="px-3 py-1.5 text-sm rounded-lg border border-stone-300 hover:bg-stone-100 disabled:opacity-40"
        :disabled="result.page <= 1"
        @click="emit('page', result!.page - 1)"
      >
        Sebelumnya
      </button>
      <button
        class="px-3 py-1.5 text-sm rounded-lg border border-stone-300 hover:bg-stone-100 disabled:opacity-40"
        :disabled="result.page >= result.totalPages"
        @click="emit('page', result!.page + 1)"
      >
        Berikutnya
      </button>
    </div>
  </div>
</template>
