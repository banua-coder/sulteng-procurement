import { ref, onMounted } from 'vue'
import { getRealisasiSummary, getRealisasi } from '../api/procurement'
import type { RealisasiSummary, JoinedRecord } from '../types/procurement'

export function useRealisasi() {
  const summary = ref<RealisasiSummary | null>(null)
  const records = ref<JoinedRecord[]>([])
  const loading = ref(false)
  const available = ref(false) // false when API returns 503 (SPSE not loaded)

  async function load() {
    loading.value = true
    try {
      const [s, r] = await Promise.all([getRealisasiSummary(), getRealisasi()])
      summary.value = s
      records.value = r
      available.value = true
    } catch {
      available.value = false
    } finally {
      loading.value = false
    }
  }

  onMounted(load)

  return { summary, records, loading, available, load }
}
