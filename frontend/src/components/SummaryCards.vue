<script setup lang="ts">
import type { Summary } from '../types/procurement'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { formatRupiah } from '@/utils/format'

defineProps<{ summary: Summary | null }>()
</script>

<template>
  <div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
    <Card>
      <CardHeader class="pb-2">
        <CardTitle class="text-sm font-medium text-muted-foreground">Total pagu anggaran</CardTitle>
      </CardHeader>
      <CardContent>
        <div class="text-2xl font-bold">
          {{ summary ? formatRupiah(summary.totalPagu) : '-' }}
        </div>
      </CardContent>
    </Card>

    <Card>
      <CardHeader class="pb-2">
        <CardTitle class="text-sm font-medium text-muted-foreground">Jumlah paket</CardTitle>
      </CardHeader>
      <CardContent>
        <div class="text-2xl font-bold">
          {{ summary ? summary.totalPaket.toLocaleString('id-ID') : '-' }}
        </div>
      </CardContent>
    </Card>

    <Card>
      <CardHeader class="pb-2">
        <CardTitle class="text-sm font-medium text-muted-foreground">Jumlah wilayah</CardTitle>
      </CardHeader>
      <CardContent>
        <div class="text-2xl font-bold">
          {{ summary?.kldiCount ?? '-' }}
        </div>
      </CardContent>
    </Card>

    <Card>
      <CardHeader class="pb-2">
        <CardTitle class="text-sm font-medium text-muted-foreground">Wilayah terbesar</CardTitle>
      </CardHeader>
      <CardContent>
        <div class="text-2xl font-bold truncate">
          {{ summary?.topKldi ?? '-' }}
        </div>
      </CardContent>
    </Card>
  </div>
</template>
