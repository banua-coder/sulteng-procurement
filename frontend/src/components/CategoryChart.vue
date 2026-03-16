<script setup lang="ts">
import { computed } from 'vue'
import { Bar } from 'vue-chartjs'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  BarElement,
  Title,
  Tooltip,
} from 'chart.js'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import type { CategoryTotal } from '../types/procurement'
import { formatRupiah } from '@/utils/format'

ChartJS.register(CategoryScale, LinearScale, BarElement, Title, Tooltip)

const props = defineProps<{
  data: CategoryTotal[]
  title: string
  maxItems?: number
}>()

const chartData = computed(() => {
  const items = props.data.slice(0, props.maxItems ?? 10)
  return {
    labels: items.map((d) => d.name),
    datasets: [
      {
        data: items.map((d) => d.total),
        backgroundColor: [
          'oklch(0.216 0.006 56.043)',
          'oklch(0.268 0.007 34.298)',
          'oklch(0.37 0.008 56)',
          'oklch(0.47 0.01 56)',
          'oklch(0.553 0.013 58.071)',
          'oklch(0.65 0.012 56)',
          'oklch(0.709 0.01 56.259)',
          'oklch(0.8 0.006 56)',
          'oklch(0.87 0.004 56)',
          'oklch(0.923 0.003 48.717)',
        ],
        borderRadius: 4,
      },
    ],
  }
})

const chartOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: { display: false },
    tooltip: {
      callbacks: {
        label: (ctx: { raw: unknown }) => formatRupiah(ctx.raw as number),
      },
    },
  },
  scales: {
    y: {
      ticks: {
        callback: (val: number | string) => formatRupiah(Number(val)),
      },
    },
    x: {
      ticks: { maxRotation: 45, font: { size: 10 } },
    },
  },
}))
</script>

<template>
  <Card class="h-full">
    <CardHeader>
      <CardTitle>{{ title }}</CardTitle>
      <CardDescription>
        Menampilkan {{ maxItems ?? 10 }} teratas berdasarkan total pagu.
      </CardDescription>
    </CardHeader>
    <CardContent>
      <div class="h-72">
        <Bar :data="chartData" :options="(chartOptions as any)" />
      </div>
    </CardContent>
  </Card>
</template>
