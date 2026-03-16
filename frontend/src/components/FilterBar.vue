<script setup lang="ts">
import type { Filters } from '../types/procurement'

const props = defineProps<{
  filters: Filters | null
  kldi: string
  jenisPengadaan: string
  metode: string
  search: string
  pageSize: number
}>()

const emit = defineEmits<{
  'update:kldi': [value: string]
  'update:jenisPengadaan': [value: string]
  'update:metode': [value: string]
  'update:search': [value: string]
  'update:pageSize': [value: number]
}>()
</script>

<template>
  <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-4">
    <div>
      <label class="block text-sm font-medium text-stone-600 mb-1">Wilayah (KLPD)</label>
      <select
        :value="kldi"
        class="w-full rounded-lg border border-stone-300 bg-white px-3 py-2 text-sm"
        @change="emit('update:kldi', ($event.target as HTMLSelectElement).value)"
      >
        <option value="">Semua wilayah</option>
        <option v-for="k in props.filters?.kldi" :key="k" :value="k">{{ k }}</option>
      </select>
    </div>
    <div>
      <label class="block text-sm font-medium text-stone-600 mb-1">Jenis pengadaan</label>
      <select
        :value="jenisPengadaan"
        class="w-full rounded-lg border border-stone-300 bg-white px-3 py-2 text-sm"
        @change="emit('update:jenisPengadaan', ($event.target as HTMLSelectElement).value)"
      >
        <option value="">Semua jenis</option>
        <option v-for="j in props.filters?.jenisPengadaan" :key="j" :value="j">{{ j }}</option>
      </select>
    </div>
    <div>
      <label class="block text-sm font-medium text-stone-600 mb-1">Metode pengadaan</label>
      <select
        :value="metode"
        class="w-full rounded-lg border border-stone-300 bg-white px-3 py-2 text-sm"
        @change="emit('update:metode', ($event.target as HTMLSelectElement).value)"
      >
        <option value="">Semua metode</option>
        <option v-for="m in props.filters?.metode" :key="m" :value="m">{{ m }}</option>
      </select>
    </div>
    <div>
      <label class="block text-sm font-medium text-stone-600 mb-1">Cari paket / satker</label>
      <input
        :value="search"
        type="text"
        placeholder="Ketik kata kunci..."
        class="w-full rounded-lg border border-stone-300 bg-white px-3 py-2 text-sm"
        @input="emit('update:search', ($event.target as HTMLInputElement).value)"
      />
    </div>
    <div>
      <label class="block text-sm font-medium text-stone-600 mb-1">Baris per halaman</label>
      <select
        :value="pageSize"
        class="w-full rounded-lg border border-stone-300 bg-white px-3 py-2 text-sm"
        @change="emit('update:pageSize', Number(($event.target as HTMLSelectElement).value))"
      >
        <option :value="10">10</option>
        <option :value="25">25</option>
        <option :value="50">50</option>
        <option :value="100">100</option>
      </select>
    </div>
  </div>
</template>
