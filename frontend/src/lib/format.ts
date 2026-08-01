const UNITS = ['B', 'KB', 'MB', 'GB', 'TB', 'PB']

export function formatBytes(n: number): string {
  if (n <= 0) return '0 B'
  const exp = Math.min(Math.floor(Math.log(n) / Math.log(1024)), UNITS.length - 1)
  const val = n / Math.pow(1024, exp)
  return `${val.toFixed(exp === 0 ? 0 : 2)} ${UNITS[exp]}`
}

export function formatPercent(part: number, whole: number): string {
  if (whole <= 0) return '0.0%'
  return `${((part / whole) * 100).toFixed(1)}%`
}

export function formatDate(iso: string): string {
  const d = new Date(iso)
  if (isNaN(d.getTime())) return '—'
  return d.toLocaleString()
}

export function formatDateFromUnix(sec: number): string {
  if (!sec) return '—'
  return new Date(sec * 1000).toLocaleString()
}

export function formatDuration(ms: number): string {
  if (ms < 1000) return `${ms} ms`
  const s = ms / 1000
  if (s < 60) return `${s.toFixed(1)} s`
  const m = Math.floor(s / 60)
  return `${m}m ${Math.round(s - m * 60)}s`
}
