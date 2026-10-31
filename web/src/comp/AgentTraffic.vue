<template>
  <div id="traffic" class="scroll-mt-6">
    <ChartCard
      :title="title || $t('agents_page.traffic_title')"
      :subtitle="subtitle || $t('agents_page.traffic_desc')"
      :table="hasRows ? tableData : null"
    >
      <template #actions>
        <slot name="actions" />
      </template>
      <div v-if="loading && !hasRows" class="p-8 text-center text-sm text-theme-textLight">
        {{ $t('common.loading') }}
      </div>
      <p v-else-if="!hasRows" class="py-8 text-center text-sm text-theme-textLight">
        {{ $t('agents_page.traffic_empty') }}
      </p>
      <BarChart
        v-else
        :data="chartData"
        :labels="chartLabels"
        :colors="['var(--chart-3)', 'var(--chart-2)']"
        :series-names="[t('agents_page.traffic_completed'), t('agents_page.traffic_other')]"
        :height="220"
        :aria-label="$t('agents_page.traffic_title')"
      />
      <p class="mt-2 text-xs text-theme-textLight">
        {{ $t('agents_page.traffic_scope', { window: TRAFFIC_WINDOW }) }}
      </p>
    </ChartCard>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BarChart from '@/comp/charts/BarChart.vue'
import ChartCard from '@/comp/charts/ChartCard.vue'
import { traffic_api } from '@/api'

const props = defineProps({
  workspaceId: { type: String, required: true },
  agentName: { type: String, required: true },
  title: { type: String, default: '' },
  subtitle: { type: String, default: '' },
})

const TRAFFIC_WINDOW = 200
const CHART_DAYS = 14

const { t } = useI18n()
const loading = ref(false)
const rows = ref([])

function isProblem(call) {
  return call.result !== 'completed'
}

async function loadTraffic() {
  loading.value = true

  try {
    const res = await traffic_api.list(props.workspaceId, {
      limit: TRAFFIC_WINDOW,
      offset: 0,
    })
    rows.value = (res.data.calls || []).filter(
      (call) => call.sourceAgent === props.agentName || call.destinationAgent === props.agentName
    )
  } catch (_) {
    rows.value = []
  } finally {
    loading.value = false
  }
}

const days = computed(() =>
  [...Array(CHART_DAYS)].map((_, i) => {
    const day = new Date()
    day.setDate(day.getDate() - (CHART_DAYS - 1 - i))

    return day
  })
)

const dateLabels = computed(() =>
  days.value.map((day) => day.toLocaleDateString(undefined, { day: 'numeric', month: 'short' }))
)

const chartLabels = computed(() => days.value.map((day) => day.getDate()))

const chartData = computed(() => {
  const byDay = Object.fromEntries(
    days.value.map((day) => [day.toISOString().slice(0, 10), { ok: 0, bad: 0 }])
  )

  for (const call of rows.value) {
    const bucket = byDay[new Date(call.createdAt).toISOString().slice(0, 10)]

    if (bucket) {
      if (isProblem(call)) bucket.bad++
      else bucket.ok++
    }
  }

  return days.value.map((day) => {
    const bucket = byDay[day.toISOString().slice(0, 10)]

    return [bucket.ok, bucket.bad]
  })
})

const hasRows = computed(() => rows.value.length > 0)

const tableData = computed(() => ({
  columns: [
    t('agents_page.traffic_day'),
    t('agents_page.traffic_completed'),
    t('agents_page.traffic_other'),
  ],
  rows: dateLabels.value.map((label, i) => [label, chartData.value[i][0], chartData.value[i][1]]),
}))

watch(
  () => props.agentName,
  () => loadTraffic()
)

onMounted(loadTraffic)
</script>
