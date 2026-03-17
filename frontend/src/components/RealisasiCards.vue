<script setup lang="ts">
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import type { RealisasiSummary } from '../types/procurement'
import { formatRupiah } from '@/utils/format'

defineProps<{ summary: RealisasiSummary }>()
</script>

<template>
  <div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
    <Card>
      <CardHeader class="pb-1">
        <CardTitle class="text-sm font-medium text-muted-foreground">Total kontrak</CardTitle>
      </CardHeader>
      <CardContent>
        <p class="text-2xl font-bold">{{ formatRupiah(summary.totalKontrak) }}</p>
        <p class="text-xs text-muted-foreground mt-1">dari {{ formatRupiah(summary.totalPagu) }} pagu</p>
      </CardContent>
    </Card>

    <Card>
      <CardHeader class="pb-1">
        <CardTitle class="text-sm font-medium text-muted-foreground">Utilisasi anggaran</CardTitle>
      </CardHeader>
      <CardContent>
        <p class="text-2xl font-bold">{{ summary.utilisasiRate.toFixed(1) }}%</p>
        <div class="w-full bg-stone-200 rounded-full h-1.5 mt-2">
          <div class="bg-stone-700 h-1.5 rounded-full" :style="{ width: Math.min(summary.utilisasiRate, 100) + '%' }" />
        </div>
      </CardContent>
    </Card>

    <Card>
      <CardHeader class="pb-1">
        <CardTitle class="text-sm font-medium text-muted-foreground">Paket selesai</CardTitle>
      </CardHeader>
      <CardContent>
        <p class="text-2xl font-bold">{{ summary.totalSelesai.toLocaleString('id-ID') }}</p>
        <Badge variant="secondary" class="mt-1 text-xs">Kontrak ditandatangani</Badge>
      </CardContent>
    </Card>

    <Card>
      <CardHeader class="pb-1">
        <CardTitle class="text-sm font-medium text-muted-foreground">Belum ditenderkan</CardTitle>
      </CardHeader>
      <CardContent>
        <p class="text-2xl font-bold">{{ summary.belumTender.toLocaleString('id-ID') }}</p>
        <p class="text-xs text-muted-foreground mt-1">Masih di tahap perencanaan</p>
      </CardContent>
    </Card>
  </div>
</template>
