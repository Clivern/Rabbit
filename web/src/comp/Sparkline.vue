<template>
  <svg
    :viewBox="`0 0 ${W} ${height}`"
    preserveAspectRatio="none"
    class="block w-full"
    :style="{ height: `${height}px` }"
    aria-hidden="true"
  >
    <path :d="area" :style="{ fill: color }" fill-opacity="0.1" />
    <path
      :d="line"
      fill="none"
      :style="{ stroke: color }"
      stroke-width="2"
      stroke-linejoin="round"
      stroke-linecap="round"
      vector-effect="non-scaling-stroke"
    />
  </svg>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  values: { type: Array, required: true },
  color: { type: String, default: 'var(--chart-1)' },
  height: { type: Number, default: 32 },
})

const W = 100

const points = computed(() => {
  const min = Math.min(...props.values)
  const max = Math.max(...props.values)
  const span = max - min || 1
  const n = Math.max(1, props.values.length - 1)
  return props.values.map((v, i) => [(i / n) * W, 2 + (1 - (v - min) / span) * (props.height - 4)])
})

const line = computed(() =>
  points.value.map(([x, y], i) => `${i ? 'L' : 'M'}${x.toFixed(2)},${y.toFixed(2)}`).join('')
)
const area = computed(() => `${line.value}L${W},${props.height}L0,${props.height}Z`)
</script>
