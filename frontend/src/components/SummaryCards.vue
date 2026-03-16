<script setup lang="ts">
import type { Summary } from '../types/procurement'

defineProps<{ summary: Summary | null }>()

function formatRupiah(value: number): string {
  if (value >= 1e12) return `Rp${(value / 1e12).toFixed(2)} T`
  if (value >= 1e9) return `Rp${(value / 1e9).toFixed(2)} M`
  if (value >= 1e6) return `Rp${(value / 1e6).toFixed(1)} Jt`
  return `Rp${value.toLocaleString('id-ID')}`
}
</script>

<template>
  <div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
    <div class="rounded-xl bg-amber-50 border border-amber-200 p-4">
      <div class="text-sm text-stone-500">Total pagu anggaran</div>
      <div class="text-2xl font-bold text-stone-800 mt-1">
        {{ summary ? formatRupiah(summary.totalPagu) : '-' }}
      </div>
    </div>
    <div class="rounded-xl bg-amber-50 border border-amber-200 p-4">
      <div class="text-sm text-stone-500">Jumlah paket</div>
      <div class="text-2xl font-bold text-stone-800 mt-1">
        {{ summary ? summary.totalPaket.toLocaleString('id-ID') : '-' }}
      </div>
    </div>
    <div class="rounded-xl bg-amber-50 border border-amber-200 p-4">
      <div class="text-sm text-stone-500">Jumlah wilayah</div>
      <div class="text-2xl font-bold text-stone-800 mt-1">
        {{ summary?.kldiCount ?? '-' }}
      </div>
    </div>
    <div class="rounded-xl bg-amber-50 border border-amber-200 p-4">
      <div class="text-sm text-stone-500">Wilayah terbesar</div>
      <div class="text-2xl font-bold text-stone-800 mt-1 truncate">
        {{ summary?.topKldi ?? '-' }}
      </div>
    </div>
  </div>
</template>
