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
import type { CategoryTotal } from '../types/procurement'

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
          '#292524', '#44403c', '#57534e', '#78716c',
          '#a8a29e', '#d6d3d1', '#e7e5e4', '#f5f5f4',
          '#fafaf9', '#b8b2a8',
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
        label: (ctx: { raw: unknown }) => {
          const val = ctx.raw as number
          if (val >= 1e12) return `Rp${(val / 1e12).toFixed(2)} T`
          if (val >= 1e9) return `Rp${(val / 1e9).toFixed(2)} M`
          return `Rp${(val / 1e6).toFixed(1)} Jt`
        },
      },
    },
  },
  scales: {
    y: {
      ticks: {
        callback: (val: number | string) => {
          const n = Number(val)
          if (n >= 1e12) return `Rp${(n / 1e12).toFixed(1)}T`
          if (n >= 1e9) return `Rp${(n / 1e9).toFixed(0)}M`
          return `Rp${(n / 1e6).toFixed(0)}Jt`
        },
      },
    },
    x: {
      ticks: {
        maxRotation: 45,
        font: { size: 10 },
      },
    },
  },
}))
</script>

<template>
  <div>
    <h3 class="font-semibold text-stone-800">{{ title }}</h3>
    <p class="text-sm text-stone-500 mb-4">
      Grafik menampilkan {{ maxItems ?? 10 }} teratas berdasarkan total pagu.
    </p>
    <div class="h-72">
      <Bar :data="chartData" :options="(chartOptions as any)" />
    </div>
  </div>
</template>
