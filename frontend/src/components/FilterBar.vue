<script setup lang="ts">
import type { Filters } from '../types/procurement'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

const ALL = '__all__'

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

function toModel(v: string) {
  return v === '' ? ALL : v
}

function fromModel(v: string | undefined) {
  return v === ALL || v === undefined ? '' : v
}
</script>

<template>
  <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-4">
    <div class="space-y-1">
      <label class="text-sm font-medium text-muted-foreground">Wilayah (KLPD)</label>
      <Select :model-value="toModel(kldi)" @update:model-value="(v) => emit('update:kldi', fromModel(v))">
        <SelectTrigger class="w-full">
          <SelectValue placeholder="Semua wilayah" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem :value="ALL">Semua wilayah</SelectItem>
          <SelectItem v-for="k in props.filters?.kldi" :key="k" :value="k">{{ k }}</SelectItem>
        </SelectContent>
      </Select>
    </div>

    <div class="space-y-1">
      <label class="text-sm font-medium text-muted-foreground">Jenis pengadaan</label>
      <Select :model-value="toModel(jenisPengadaan)" @update:model-value="(v) => emit('update:jenisPengadaan', fromModel(v))">
        <SelectTrigger class="w-full">
          <SelectValue placeholder="Semua jenis" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem :value="ALL">Semua jenis</SelectItem>
          <SelectItem v-for="j in props.filters?.jenisPengadaan" :key="j" :value="j">{{ j }}</SelectItem>
        </SelectContent>
      </Select>
    </div>

    <div class="space-y-1">
      <label class="text-sm font-medium text-muted-foreground">Metode pengadaan</label>
      <Select :model-value="toModel(metode)" @update:model-value="(v) => emit('update:metode', fromModel(v))">
        <SelectTrigger class="w-full">
          <SelectValue placeholder="Semua metode" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem :value="ALL">Semua metode</SelectItem>
          <SelectItem v-for="m in props.filters?.metode" :key="m" :value="m">{{ m }}</SelectItem>
        </SelectContent>
      </Select>
    </div>

    <div class="space-y-1">
      <label class="text-sm font-medium text-muted-foreground">Cari paket / satker</label>
      <Input
        :value="search"
        type="text"
        placeholder="Ketik kata kunci..."
        @input="emit('update:search', ($event.target as HTMLInputElement).value)"
      />
    </div>

    <div class="space-y-1">
      <label class="text-sm font-medium text-muted-foreground">Baris per halaman</label>
      <Select
        :model-value="String(pageSize)"
        @update:model-value="(v) => emit('update:pageSize', Number(v ?? 25))"
      >
        <SelectTrigger class="w-full">
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="10">10</SelectItem>
          <SelectItem value="25">25</SelectItem>
          <SelectItem value="50">50</SelectItem>
          <SelectItem value="100">100</SelectItem>
        </SelectContent>
      </Select>
    </div>
  </div>
</template>
