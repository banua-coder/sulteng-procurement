export function formatRupiah(value: number): string {
  if (value >= 1e12) return `Rp${(value / 1e12).toFixed(2)} T`
  if (value >= 1e9) return `Rp${(value / 1e9).toFixed(2)} M`
  if (value >= 1e6) return `Rp${(value / 1e6).toFixed(1)} Jt`
  return `Rp${value.toLocaleString('id-ID')}`
}
