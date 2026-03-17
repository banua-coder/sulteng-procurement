<script setup lang="ts">
import type { PaginatedResult, TenderResult } from '../types/procurement'
import { Button } from '@/components/ui/button'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
  TableEmpty,
} from '@/components/ui/table'
import { formatRupiah } from '@/utils/format'

const props = withDefaults(defineProps<{
  result: PaginatedResult | null
  loading: boolean
  sortBy: string
  sortDir: string
  tenderMap?: Map<number, TenderResult>
}>(), {
  tenderMap: undefined
})

const emit = defineEmits<{
  sort: [field: string]
  page: [page: number]
}>()

const columns = [
  { key: 'kldi', label: 'Wilayah' },
  { key: 'satuanKerja', label: 'Satuan Kerja' },
  { key: 'paket', label: 'Paket' },
  { key: 'jenisPengadaan', label: 'Jenis' },
  { key: 'metode', label: 'Metode' },
  { key: 'pagu', label: 'Pagu' },
]

function sortIcon(field: string, currentSort: string, dir: string): string {
  if (field !== currentSort) return ''
  return dir === 'desc' ? ' ▼' : ' ▲'
}
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-3">
      <h3 class="font-semibold">Tabel detail</h3>
      <div class="text-sm text-muted-foreground" v-if="result">
        Menampilkan {{ result.data.length }} dari {{ result.total.toLocaleString('id-ID') }} data
        <span class="ml-4">Halaman {{ result.page }} dari {{ result.totalPages }}</span>
      </div>
    </div>

    <div class="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead
              v-for="col in columns"
              :key="col.key"
              class="cursor-pointer hover:bg-muted/50 whitespace-nowrap"
              @click="emit('sort', col.key)"
            >
              {{ col.label }}{{ sortIcon(col.key, sortBy, sortDir) }}
            </TableHead>
            <TableHead v-if="props.tenderMap">Nilai Kontrak</TableHead>
            <TableHead v-if="props.tenderMap">Pemenang</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow v-if="loading">
            <TableEmpty :colspan="columns.length" class="py-8 text-center text-muted-foreground">
              Memuat data...
            </TableEmpty>
          </TableRow>
          <TableRow v-else-if="!result?.data.length">
            <TableEmpty :colspan="columns.length" class="py-8 text-center text-muted-foreground">
              Tidak ada data
            </TableEmpty>
          </TableRow>
          <template v-else>
            <TableRow v-for="item in result?.data" :key="item.id">
              <TableCell>{{ item.kldi }}</TableCell>
              <TableCell>{{ item.satuanKerja }}</TableCell>
              <TableCell class="max-w-xs truncate">{{ item.paket }}</TableCell>
              <TableCell class="whitespace-nowrap">{{ item.jenisPengadaan }}</TableCell>
              <TableCell class="whitespace-nowrap">{{ item.metode }}</TableCell>
              <TableCell class="text-right whitespace-nowrap font-mono">{{ formatRupiah(item.pagu) }}</TableCell>
              <TableCell v-if="props.tenderMap">
                {{ props.tenderMap.get(item.id)?.nilaiKontrak != null
                  ? formatRupiah(props.tenderMap.get(item.id)!.nilaiKontrak)
                  : '—' }}
              </TableCell>
              <TableCell v-if="props.tenderMap" class="text-xs max-w-32 truncate" :title="props.tenderMap.get(item.id)?.pemenang">
                {{ props.tenderMap.get(item.id)?.pemenang ?? '—' }}
              </TableCell>
            </TableRow>
          </template>
        </TableBody>
      </Table>
    </div>

    <div class="flex justify-end gap-2 mt-3" v-if="result && result.totalPages > 1">
      <Button
        variant="outline"
        size="sm"
        :disabled="result.page <= 1"
        @click="emit('page', result!.page - 1)"
      >
        Sebelumnya
      </Button>
      <Button
        variant="outline"
        size="sm"
        :disabled="result.page >= result.totalPages"
        @click="emit('page', result!.page + 1)"
      >
        Berikutnya
      </Button>
    </div>
  </div>
</template>
