<script setup lang="ts">
import type { Procurement } from '../types/procurement'

defineProps<{ items: Procurement[] }>()

function formatRupiah(value: number): string {
  if (value >= 1e12) return `Rp${(value / 1e12).toFixed(2)} T`
  if (value >= 1e9) return `Rp${(value / 1e9).toFixed(1)} M`
  if (value >= 1e6) return `Rp${(value / 1e6).toFixed(1)} Jt`
  return `Rp${value.toLocaleString('id-ID')}`
}
</script>

<template>
  <div>
    <h3 class="font-semibold text-stone-800">Paket pengadaan terbesar</h3>
    <p class="text-sm text-stone-500 mb-4">5 paket dengan pagu tertinggi.</p>
    <div class="space-y-4">
      <div v-for="item in items.slice(0, 5)" :key="item.id" class="border-b border-stone-200 pb-3">
        <div class="font-medium text-stone-800 text-sm leading-snug">{{ item.paket }}</div>
        <span class="inline-block mt-1 text-xs px-2 py-0.5 rounded-full bg-amber-100 text-amber-800 font-medium">
          {{ item.jenisPengadaan }}
        </span>
        <div class="text-xs text-stone-500 mt-1">{{ item.satuanKerja }}</div>
        <div class="text-sm font-semibold text-stone-700 mt-1">{{ formatRupiah(item.pagu) }}</div>
      </div>
    </div>
  </div>
</template>
