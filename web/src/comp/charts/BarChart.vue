<template>
  <div ref="el" class="relative w-full" :style="{ height: `${height + legendOffset}px` }">
    <ChartLegend v-if="legend" :items="legendItems" class="mb-1" />
    <svg
      :viewBox="`0 0 ${width} ${height}`"
      class="block w-full"
      role="img"
      :style="{ height: `${height}px` }"
      :aria-label="ariaLabel"
    >
      <g v-for="tick in yTicks" :key="tick">
        <text :x="width - RIGHT" :y="scaleY(tick) - 4" text-anchor="end" class="chart-axis-label">
          {{ yFormat(tick) }}
        </text>
        <line
          :x1="0"
          :x2="width - RIGHT"
          :y1="scaleY(tick)"
          :y2="scaleY(tick)"
          :style="{ stroke: 'var(--chart-grid)' }"
          stroke-width="1"
        />
      </g>
      <g v-for="(band, bandIndex) in bands" :key="bandIndex">
        <path
          v-for="(point, seriesIndex) in band"
          :key="`${bandIndex}-${seriesIndex}`"
          :d="point.d"
          :style="{ fill: point.color }"
          @mouseenter="hover = { band: bandIndex, series: seriesIndex, x: $event }"
          @mousemove="hover = { band: bandIndex, series: seriesIndex, x: $event }"
          @mouseleave="hover = null"
        />
        <text
          v-if="bandIndex % labelEvery === 0"
          :x="bandCenter(bandIndex)"
          :y="height - 6"
          text-anchor="middle"
          class="chart-axis-label"
        >
          {{ labels[bandIndex] }}
        </text>
      </g>
    </svg>
    <div
      v-if="hover && tooltipText"
      class="chart-tooltip"
      :style="{ left: `${Math.min(tooltipX, tooltipMaxX)}px`, top: '8px' }"
    >
      <p class="mb-1 font-semibold">
        {{ labels[hover.band] }}: {{ tooltipFootnote ?? tooltipTotal }}
      </p>
      <ul class="space-y-0.5">
        <li v-for="(part, i) in tooltipRows" :key="i" class="flex items-center gap-1.5">
          <span
            class="inline-block h-2.5 w-2.5 rounded-sm"
            :style="{ backgroundColor: part.color }"
            aria-hidden="true"
          />
          {{ part.name }}: {{ valueFormat(part.value) }}
        </li>
      </ul>
    </div>
  </div>
</template>

<script setup>
import { computed, ref } from 'vue'
import ChartLegend from './ChartLegend.vue'
import { useElementWidth } from './useElementWidth'
import { niceTicks, formatCompact, seriesColor, columnPath } from '@/lib/chart'

const props = defineProps({
  // One entry per x band, each with one value per series.
  data: { type: Array, required: true },
  labels: { type: Array, required: true },
  colors: { type: Array, default: () => [] },
  height: { type: Number, default: 260 },
  legend: { type: Boolean, default: true },
  legendOffset: { type: Number, default: 20 },
  labelEvery: { type: Number, default: 1 },
  seriesNames: { type: Array, default: () => [] },
  valueFormat: { type: Function, default: formatCompact },
  yFormat: { type: Function, default: formatCompact },
  ariaLabel: { type: String, default: '' },
  tooltipFootnote: { type: String, default: null },
})

const TOP = 8
const BOTTOM = 24
const RIGHT = 44
const GAP = 6

const { el, width } = useElementWidth()
const hover = ref(null)

const innerH = computed(() => props.height - TOP - BOTTOM)
const bandCount = computed(() => props.data.length)
const seriesCount = computed(() => props.data[0]?.length ?? 0)
const columnW = computed(
  () => (width.value - RIGHT - GAP * (bandCount.value - 1)) / bandCount.value
)
const maxTotal = computed(() =>
  Math.max(0, ...props.data.map((row) => row.reduce((a, b) => a + b, 0)))
)
const yTicks = computed(() => niceTicks(maxTotal.value))

function colorAt(index) {
  return props.colors[index] || seriesColor(index)
}

function scaleY(value) {
  const max = Math.max(...yTicks.value, 1)
  return TOP + innerH.value - (value / max) * innerH.value
}

function bandCenter(bandIndex) {
  return bandIndex * (columnW.value + GAP) + columnW.value / 2
}

const bands = computed(() =>
  props.data.map((row, bandIndex) => {
    let stacked = 0
    return row.map((value, seriesIndex) => {
      const x = bandIndex * (columnW.value + GAP)
      const y = scaleY(stacked + value)
      const d = columnPath(x, y, columnW.value, scaleY(stacked) - y)
      stacked += value
      return { d, color: colorAt(seriesIndex) }
    })
  })
)

const legendItems = computed(() =>
  Array.from({ length: seriesCount.value }, (_, i) => ({
    name: props.seriesNames[i] ?? `Series ${i + 1}`,
    color: colorAt(i),
  }))
)

const tooltipTotal = computed(() => {
  if (!hover.value) return ''
  return props.valueFormat(props.data[hover.value.band].reduce((a, b) => a + b, 0))
})

const tooltipRows = computed(() => {
  if (!hover.value) return []
  return props.data[hover.value.band].map((value, i) => ({
    value,
    name: props.seriesNames[i] ?? `Series ${i + 1}`,
    color: colorAt(i),
  }))
})

const tooltipText = computed(() => tooltipRows.value.some((row) => row.value > 0))

const tooltipX = computed(() => {
  if (!hover.value || !el.value) return 0
  const rect = el.value.getBoundingClientRect()
  return hover.value.x.clientX - rect.left + 12
})

const tooltipMaxX = computed(() => Math.max(0, width.value - 180))
</script>
