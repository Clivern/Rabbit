export const SERIES_COLORS = [
  'var(--chart-1)',
  'var(--chart-2)',
  'var(--chart-3)',
  'var(--chart-4)',
  'var(--chart-5)',
]

export function seriesColor(index) {
  return SERIES_COLORS[index] || 'var(--primary-400)'
}

// Rounds an axis maximum up to a clean step and returns the ticks.
export function niceTicks(max, count = 4) {
  if (!max || max <= 0) return [0, 1]
  const raw = max / count
  const magnitude = Math.pow(10, Math.floor(Math.log10(raw)))
  const residual = raw / magnitude
  const step = (residual > 5 ? 10 : residual > 2 ? 5 : residual > 1 ? 2 : 1) * magnitude
  const ticks = []
  for (let v = 0; v <= max + step * 0.001; v += step) ticks.push(+v.toFixed(10))
  if (ticks[ticks.length - 1] < max) ticks.push(+(ticks[ticks.length - 1] + step).toFixed(10))
  return ticks
}

export function formatCompact(value) {
  if (value == null || Number.isNaN(value)) return '–'
  const abs = Math.abs(value)
  if (abs >= 1e9) return `${+(value / 1e9).toFixed(1)}B`
  if (abs >= 1e6) return `${+(value / 1e6).toFixed(1)}M`
  if (abs >= 1e4) return `${+(value / 1e3).toFixed(1)}K`
  if (abs >= 100 || Number.isInteger(value)) return Math.round(value).toLocaleString()
  return `${+value.toFixed(2)}`
}

export function formatDuration(ms) {
  if (ms >= 1000) return `${+(ms / 1000).toFixed(2)}s`
  if (ms >= 10) return `${Math.round(ms)}ms`
  return `${+ms.toFixed(1)}ms`
}

// Path for a column with a 4px rounded top and a square base.
export function columnPath(x, y, width, height, radius = 4) {
  if (height <= 0) return ''
  const r = Math.min(radius, width / 2, height)
  return [
    `M${x},${y + height}`,
    `V${y + r}`,
    `Q${x},${y} ${x + r},${y}`,
    `H${x + width - r}`,
    `Q${x + width},${y} ${x + width},${y + r}`,
    `V${y + height}`,
    'Z',
  ].join(' ')
}
