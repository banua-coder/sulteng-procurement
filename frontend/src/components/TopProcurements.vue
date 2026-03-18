<script setup lang="ts">
import type { Procurement } from '../types/procurement'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { formatRupiah } from '@/utils/format'

defineProps<{ items: Procurement[] }>()
</script>

<template>
  <Card class="h-full">
    <CardHeader>
      <CardTitle>Paket pengadaan terbesar</CardTitle>
      <CardDescription>Paket dengan pagu tertinggi.</CardDescription>
    </CardHeader>
    <CardContent>
      <div class="space-y-4 max-h-72 overflow-y-auto pr-1">
        <div
          v-for="(item, i) in items"
          :key="`${item.id}-${i}`"
          class="border-b border-border pb-3 last:border-0 last:pb-0"
        >
          <div class="font-medium text-sm leading-snug">{{ item.paket }}</div>
          <Badge variant="secondary" class="mt-1">{{ item.jenisPengadaan }}</Badge>
          <div class="text-xs text-muted-foreground mt-1">{{ item.satuanKerja }}</div>
          <div class="text-sm font-semibold mt-1">{{ formatRupiah(item.pagu) }}</div>
        </div>
      </div>
    </CardContent>
  </Card>
</template>
