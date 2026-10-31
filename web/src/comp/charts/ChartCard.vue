<template>
  <section class="section !p-0 overflow-hidden">
    <header
      class="flex flex-wrap items-start justify-between gap-3 border-b border-theme-border bg-primary-50/40 px-5 py-4"
    >
      <div class="min-w-0">
        <h2 class="text-sm font-semibold text-theme-text">{{ title }}</h2>
        <p v-if="subtitle" class="mt-0.5 text-xs text-theme-textLight">{{ subtitle }}</p>
      </div>
      <div class="flex items-center gap-2">
        <slot name="actions" />
        <button
          v-if="table"
          type="button"
          class="rounded-md border border-theme-border px-2 py-1 text-xs font-medium text-theme-textLight hover:bg-theme-hover hover:text-theme-text"
          :aria-pressed="showTable"
          @click="showTable = !showTable"
        >
          {{ showTable ? $t('charts.chart') : $t('charts.table') }}
        </button>
      </div>
    </header>

    <div class="px-5 py-4">
      <div v-if="showTable && table" class="max-h-72 overflow-auto">
        <table class="w-full text-xs">
          <thead class="sticky top-0 bg-white">
            <tr class="border-b border-theme-border">
              <th
                v-for="(column, i) in table.columns"
                :key="column"
                class="py-2 pr-4 font-semibold text-theme-text"
                :class="i === 0 ? 'text-left' : 'text-right'"
              >
                {{ column }}
              </th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="(row, r) in table.rows"
              :key="r"
              class="border-b border-theme-border last:border-0"
            >
              <td
                v-for="(cell, i) in row"
                :key="i"
                class="py-1.5 pr-4 tabular-nums"
                :class="i === 0 ? 'text-left text-theme-text' : 'text-right text-theme-textLight'"
              >
                {{ cell }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <slot v-else />
    </div>
  </section>
</template>

<script setup>
import { ref } from 'vue'

defineProps({
  title: { type: String, required: true },
  subtitle: { type: String, default: '' },
  // { columns: string[], rows: any[][] } - enables the table view toggle
  table: { type: Object, default: null },
})

const showTable = ref(false)
</script>
