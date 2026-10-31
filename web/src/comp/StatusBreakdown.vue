<template>
  <span class="inline-flex flex-wrap items-center gap-x-1.5 gap-y-1 text-sm">
    <template v-for="(item, index) in items" :key="item.key">
      <span v-if="index > 0" class="text-theme-textLight" aria-hidden="true">·</span>
      <span class="font-medium" :class="item.count ? item.class : 'text-theme-textLight'">
        {{ item.count }} {{ $t(`agents_page.health_${item.key}`).toLowerCase() }}
      </span>
    </template>
  </span>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  counts: {
    type: Object,
    required: true,
  },
  // Show only the statuses that have at least one item.
  hideZero: {
    type: Boolean,
    default: false,
  },
})

const STATUSES = [
  { key: 'critical', class: 'text-red-700' },
  { key: 'unknown', class: 'text-slate-600' },
  { key: 'warning', class: 'text-amber-700' },
  { key: 'passing', class: 'text-emerald-700' },
]

const items = computed(() =>
  STATUSES.map((status) => ({ ...status, count: props.counts[status.key] || 0 })).filter(
    (item) => !props.hideZero || item.count
  )
)
</script>
