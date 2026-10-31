<template>
  <ul class="flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-theme-textLight">
    <li v-for="item in items" :key="item.name" class="flex items-center gap-1.5">
      <span
        v-if="shape === 'line'"
        class="inline-block h-0.5 w-3 rounded-full"
        :style="{ backgroundColor: item.color }"
        aria-hidden="true"
      />
      <span
        v-else
        class="inline-block h-2.5 w-2.5 rounded-sm"
        :style="{ backgroundColor: item.color }"
        aria-hidden="true"
      />
      {{ item.name }}
    </li>
  </ul>
</template>

<script setup>
defineProps({
  items: { type: Array, required: true },
  shape: { type: String, default: 'rect' },
})
</script>
