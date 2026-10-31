<template>
  <div class="min-h-screen bg-theme-bg">
    <AppNav />

    <main class="w-full px-4 sm:px-6 lg:px-8 py-8">
      <header class="mb-8">
        <h1 class="text-2xl sm:text-3xl font-semibold text-theme-text tracking-tight">
          {{ $t('dashboard.title') }}
        </h1>
        <p class="mt-1 text-sm text-theme-textLight">
          {{ $t('dashboard.subtitle') }}
        </p>
        <p v-if="user?.email" class="mt-2 text-sm text-theme-text">
          {{ $t('dashboard.welcome_back') }}
          <span class="font-medium">{{ user.name || user.email }}</span>
        </p>
      </header>

      <section class="mb-8 grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-4">
        <div class="section bg-blue-50/50">
          <p class="mb-3 text-sm text-theme-textLight">{{ $t('dashboard.metrics_calls') }}</p>
          <p class="text-3xl font-normal tabular-nums tracking-tight text-theme-text">
            {{ formatCount(callsTotal) }}
          </p>
          <Sparkline v-if="callsSeries.length > 1" :values="callsSeries" class="mt-3" />
          <p class="mt-2 text-xs text-theme-textLight">
            {{ $t('dashboard.metrics_calls_sub', { window: METRICS_WINDOW }) }}
          </p>
        </div>
        <div class="section bg-red-50/50">
          <p class="mb-3 text-sm text-theme-textLight">{{ $t('dashboard.metrics_problems') }}</p>
          <p
            class="text-3xl font-normal tabular-nums tracking-tight text-theme-text"
            :class="{ 'text-red-700': (problemsCount ?? 0) > 0 }"
          >
            {{ formatCount(problemsCount) }}
          </p>
          <Sparkline
            v-if="problemsSeries.length > 1"
            :values="problemsSeries"
            color="var(--chart-2)"
            class="mt-3"
          />
          <p class="mt-2 text-xs text-theme-textLight">
            {{ $t('dashboard.metrics_problems_sub', { window: METRICS_WINDOW }) }}
          </p>
        </div>
        <div class="section bg-amber-50/50">
          <p class="mb-3 text-sm text-theme-textLight">{{ $t('dashboard.metrics_latency') }}</p>
          <p
            class="text-3xl font-normal tabular-nums tracking-tight text-theme-text"
            :class="latencyValueClass"
          >
            {{ avgLatency == null ? '—' : `${avgLatency} ms` }}
          </p>
          <Sparkline
            v-if="latencySeries.length > 1"
            :values="latencySeries"
            color="var(--chart-4)"
            class="mt-3"
          />
          <p class="mt-2 text-xs text-theme-textLight">
            {{ $t('dashboard.metrics_latency_sub', { window: METRICS_WINDOW }) }}
          </p>
        </div>
        <div class="section bg-emerald-50/50">
          <p class="mb-3 text-sm text-theme-textLight">{{ $t('dashboard.metrics_agents') }}</p>
          <p v-if="agentsList === null" class="text-3xl font-normal text-theme-text">—</p>
          <StatusBreakdown v-else :counts="healthCounts" class="mt-1" />
          <p class="mt-2 text-xs text-theme-textLight">
            {{ $t('dashboard.metrics_agents_sub', { total: agentsTotal ?? 0 }) }}
          </p>
        </div>
      </section>

      <section class="mb-8 grid grid-cols-1 gap-6 lg:grid-cols-3">
        <div class="lg:col-span-2">
          <AgentTraffic
            v-if="selectedAgent"
            :key="selectedAgent"
            :title="$t('dashboard.agents_traffic_title')"
            :subtitle="$t('dashboard.agents_traffic_desc')"
            :workspace-id="currentWorkspace.id"
            :agent-name="selectedAgent"
          >
            <template #actions>
              <select
                v-model="selectedAgent"
                class="input-field w-auto"
                :aria-label="$t('dashboard.agents_traffic_title')"
              >
                <option v-for="name in agentNames" :key="name" :value="name">{{ name }}</option>
              </select>
            </template>
          </AgentTraffic>
          <div
            v-else-if="agentsList !== null"
            class="section flex flex-col items-center gap-2 border-dashed px-6 py-12 text-center"
          >
            <p class="text-base font-semibold text-theme-text">
              {{ $t('dashboard.agents_traffic_empty') }}
            </p>
            <p class="max-w-sm text-sm text-theme-textLight">
              {{ $t('dashboard.agents_traffic_empty_desc') }}
            </p>
            <router-link :to="{ name: 'Agents' }" class="btn-primary mt-3">
              {{ $t('dashboard.agents_traffic_action') }}
            </router-link>
          </div>
        </div>
        <div class="section overflow-hidden !p-0">
          <div class="border-b border-theme-border px-6 py-4">
            <h2 class="text-lg font-semibold text-theme-text">
              {{ $t('dashboard.talkers_title') }}
            </h2>
            <p class="mt-0.5 text-sm text-theme-textLight">
              {{ $t('dashboard.talkers_desc', { window: METRICS_WINDOW }) }}
            </p>
          </div>
          <ul v-if="topTalkers.length" class="space-y-3 px-6 py-4">
            <li v-for="row in topTalkers" :key="row.name">
              <div class="mb-1 flex justify-between gap-3 text-xs">
                <router-link
                  :to="{ name: 'Agent', params: { name: row.name } }"
                  class="min-w-0 truncate font-mono text-theme-text hover:underline"
                >
                  {{ row.name }}
                </router-link>
                <span class="shrink-0 tabular-nums text-theme-textLight"
                  >{{ Math.round((row.count / talkersTotal) * 100) }}%</span
                >
              </div>
              <div class="h-2 rounded-r bg-theme-hover">
                <div
                  class="h-2 rounded-r"
                  :style="{
                    width: `${Math.round((row.count / topTalkerMax) * 100)}%`,
                    backgroundColor: 'var(--chart-1)',
                  }"
                />
              </div>
            </li>
          </ul>
          <p v-else class="px-6 py-4 text-sm text-theme-textLight">
            {{ $t('dashboard.talkers_empty') }}
          </p>
        </div>
      </section>

      <section class="mb-8 grid grid-cols-1 gap-6 lg:grid-cols-2">
        <div class="section overflow-hidden !p-0">
          <div class="border-b border-theme-border px-6 py-4">
            <div class="flex flex-wrap items-center justify-between gap-2">
              <div>
                <h2 class="text-lg font-semibold text-theme-text">
                  {{ $t('dashboard.attention_title') }}
                </h2>
                <p class="mt-0.5 text-sm text-theme-textLight">
                  {{ $t('dashboard.attention_subtitle') }}
                </p>
              </div>
            </div>
          </div>

          <ul class="divide-y divide-theme-border max-h-[28rem] overflow-y-auto">
            <li v-for="item in attentionItems" :key="item.id" class="flex gap-4 px-6 py-4">
              <span
                class="mt-1.5 h-2 w-2 shrink-0 rounded-full"
                :class="attentionDotClass(item.level)"
                aria-hidden="true"
              />
              <div class="min-w-0 flex-1">
                <p class="text-sm font-medium text-theme-text">{{ item.title }}</p>
                <p class="mt-0.5 text-sm text-theme-textLight leading-relaxed">
                  {{ item.description }}
                </p>
                <router-link
                  v-if="item.to"
                  :to="item.to"
                  class="mt-2 inline-flex text-sm font-medium text-primary-800 hover:underline"
                >
                  {{ item.action }}
                </router-link>
              </div>
            </li>
          </ul>
        </div>
        <div class="section overflow-hidden !p-0">
          <div class="border-b border-theme-border px-6 py-4">
            <h2 class="text-lg font-semibold text-theme-text">
              {{ $t('dashboard.whats_new_title') }}
            </h2>
            <p class="mt-0.5 text-sm text-theme-textLight">
              {{ $t('dashboard.whats_new_subtitle') }}
            </p>
          </div>
          <div class="px-6 py-5">
            <p class="text-xs font-medium uppercase tracking-wide text-theme-textLight">
              {{ $t('dashboard.whats_new_tip_label') }}
            </p>
            <p class="mt-2 text-base font-medium text-theme-text">
              {{ $t('dashboard.whats_new_tip_title') }}
            </p>
            <p class="mt-1 text-sm leading-relaxed text-theme-textLight">
              {{ $t('dashboard.whats_new_tip_body') }}
            </p>
          </div>
        </div>
      </section>
    </main>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import AppNav from '@/comp/AppNav.vue'
import AgentTraffic from '@/comp/AgentTraffic.vue'
import Sparkline from '@/comp/Sparkline.vue'
import StatusBreakdown from '@/comp/StatusBreakdown.vue'
import {
  agent_api,
  traffic_api,
  workspace_access_key_api,
  workspace_invite_api,
} from '@/api'
import { user } from '@/lib/auth'
import { useWorkspaceContext } from '@/lib/permission'

const { t } = useI18n()

const METRICS_WINDOW = 200
const SPARK_DAYS = 14

const { currentWorkspace, canManage } = useWorkspaceContext()

const callsTotal = ref(null)
const problemsCount = ref(null)
const avgLatency = ref(null)
const healthyAgents = ref(null)
const agentsTotal = ref(null)
const callsSeries = ref([])
const problemsSeries = ref([])
const latencySeries = ref([])
const agentNames = ref([])
const selectedAgent = ref('')
const agentsList = ref(null)
const trafficRows = ref(null)
const accessKeysTotal = ref(null)
const pendingInvites = ref(null)

function formatCount(value) {
  return value == null ? '—' : value.toLocaleString()
}

function lastDays(count) {
  return [...Array(count)].map((_, i) => {
    const day = new Date()
    day.setDate(day.getDate() - (count - 1 - i))

    return day.toISOString().slice(0, 10)
  })
}

async function loadMetrics() {
  try {
    const res = await traffic_api.list(currentWorkspace.id, {
      limit: METRICS_WINDOW,
      offset: 0,
    })
    const rows = res.data.calls || []
    trafficRows.value = rows
    callsTotal.value = res.data._meta?.total ?? rows.length

    const problems = rows.filter((call) => call.decision === 'deny' || call.result === 'failed')
    problemsCount.value = problems.length

    const latencies = rows.map((call) => call.latencyMs).filter((ms) => ms != null)
    avgLatency.value = latencies.length
      ? Math.round(latencies.reduce((sum, ms) => sum + ms, 0) / latencies.length)
      : null

    const days = lastDays(SPARK_DAYS)
    const perDay = Object.fromEntries(
      days.map((day) => [day, { calls: 0, problems: 0, latency: [] }])
    )

    for (const call of rows) {
      const bucket = perDay[new Date(call.createdAt).toISOString().slice(0, 10)]

      if (bucket) {
        bucket.calls++
        if (call.decision === 'deny' || call.result === 'failed') bucket.problems++
        if (call.latencyMs != null) bucket.latency.push(call.latencyMs)
      }
    }

    callsSeries.value = days.map((day) => perDay[day].calls)
    problemsSeries.value = days.map((day) => perDay[day].problems)
    latencySeries.value = days.map((day) =>
      perDay[day].latency.length
        ? Math.round(
            perDay[day].latency.reduce((sum, ms) => sum + ms, 0) / perDay[day].latency.length
          )
        : 0
    )
  } catch (_) {
    // Metrics stay blank when the traffic cannot be read.
  }
}

async function loadAgentHealth() {
  try {
    const loaded = []

    for (;;) {
      const res = await agent_api.list(currentWorkspace.id, {
        limit: 100,
        offset: loaded.length,
      })
      const page = res.data.agents || []
      loaded.push(...page)

      if (page.length < 100 || loaded.length >= (res.data._meta?.total || 0)) break
    }

    agentsTotal.value = loaded.length
    healthyAgents.value = loaded.filter((agent) => agent.health === 'passing').length
    agentsList.value = loaded
    agentNames.value = loaded.map((agent) => agent.name).sort()

    if (!selectedAgent.value && agentNames.value.length) {
      selectedAgent.value = agentNames.value[0]
    }
  } catch (_) {
    // Metrics stay blank when the agents cannot be read.
  }
}

async function loadSetupSignals() {
  if (canManage.value) {
    try {
      const res = await workspace_access_key_api.list(currentWorkspace.id, {
        limit: 1,
        offset: 0,
      })
      accessKeysTotal.value = res.data._meta?.total ?? (res.data.keys || []).length
    } catch (_) {
      // The key check is skipped when the keys cannot be read.
    }

    try {
      const res = await workspace_invite_api.list(currentWorkspace.id, {
        limit: 100,
        offset: 0,
      })
      pendingInvites.value = (res.data.invites || []).filter(
        (invite) => invite.status === 'pending'
      ).length
    } catch (_) {
      // The invite check is skipped when the invites cannot be read.
    }
  }
}

onMounted(() => {
  loadMetrics()
  loadAgentHealth()
  loadSetupSignals()
})

function instanceCount(agent) {
  return Object.values(agent.statusCounts || {}).reduce((sum, count) => sum + count, 0)
}

function shortNames(list) {
  const names = list.map((agent) => agent.name)

  if (names.length <= 3) return names.join(', ')

  return `${names.slice(0, 3).join(', ')} (+${names.length - 3})`
}

function agentsTarget(list) {
  if (list.length === 1) return { name: 'Agent', params: { name: list[0].name } }

  return { name: 'Agents' }
}

const attentionReady = computed(
  () =>
    agentsList.value !== null &&
    trafficRows.value !== null &&
    (!canManage.value || (accessKeysTotal.value !== null && pendingInvites.value !== null))
)

const attentionItems = computed(() => {
  const items = []

  if (agentsList.value) {
    const critical = agentsList.value.filter((agent) => agent.health === 'critical')
    const warning = agentsList.value.filter((agent) => agent.health === 'warning')
    const idle = agentsList.value.filter((agent) => instanceCount(agent) === 0)

    if (critical.length) {
      items.push({
        id: 'critical',
        level: 'error',
        title: t('dashboard.attention_critical_title'),
        description: t('dashboard.attention_critical_desc', { names: shortNames(critical) }),
        action: t('dashboard.attention_critical_action'),
        to: agentsTarget(critical),
      })
    }

    if (warning.length) {
      items.push({
        id: 'warning',
        level: 'warn',
        title: t('dashboard.attention_warning_title'),
        description: t('dashboard.attention_warning_desc', { names: shortNames(warning) }),
        action: t('dashboard.attention_warning_action'),
        to: agentsTarget(warning),
      })
    }

    if (idle.length) {
      items.push({
        id: 'idle',
        level: 'warn',
        title: t('dashboard.attention_no_instances_title'),
        description: t('dashboard.attention_no_instances_desc', { names: shortNames(idle) }),
        action: t('dashboard.attention_no_instances_action'),
        to: agentsTarget(idle),
      })
    }
  }

  if (trafficRows.value) {
    const denied = trafficRows.value.filter((call) => call.decision === 'deny')
    const failed = trafficRows.value.filter((call) => call.result === 'failed')

    if (denied.length) {
      items.push({
        id: 'denied',
        level: 'warn',
        title: t('dashboard.attention_denied_title'),
        description: t('dashboard.attention_denied_desc', {
          count: denied.length,
          window: METRICS_WINDOW,
        }),
        action: t('dashboard.attention_denied_action'),
        to: { name: 'Traffic' },
      })
    }

    if (failed.length) {
      items.push({
        id: 'failed',
        level: 'error',
        title: t('dashboard.attention_failed_title'),
        description: t('dashboard.attention_failed_desc', {
          count: failed.length,
          window: METRICS_WINDOW,
        }),
        action: t('dashboard.attention_failed_action'),
        to: { name: 'Traffic' },
      })
    }
  }

  if (canManage.value) {
    if (accessKeysTotal.value === 0) {
      items.push({
        id: 'key',
        level: 'info',
        title: t('dashboard.attention_no_key_title'),
        description: t('dashboard.attention_no_key_desc'),
        action: t('dashboard.attention_no_key_action'),
        to: { name: 'Settings' },
      })
    }

    if (pendingInvites.value > 0) {
      items.push({
        id: 'invites',
        level: 'info',
        title: t('dashboard.attention_invites_title'),
        description: t('dashboard.attention_invites_desc', { count: pendingInvites.value }),
        action: t('dashboard.attention_invites_action'),
        to: { name: 'Members' },
      })
    }
  }

  if (!items.length && attentionReady.value) {
    items.push({
      id: 'clear',
      level: 'info',
      title: t('dashboard.attention_all_clear_title'),
      description: t('dashboard.attention_all_clear_desc'),
    })
  }

  return items
})

const topTalkers = computed(() => {
  const counts = {}

  for (const call of trafficRows.value || []) {
    for (const name of [call.sourceAgent, call.destinationAgent]) {
      if (name) counts[name] = (counts[name] || 0) + 1
    }
  }

  return Object.entries(counts)
    .map(([name, count]) => ({ name, count }))
    .sort((a, b) => b.count - a.count)
    .slice(0, 6)
})

const latencyValueClass = computed(() => {
  if (avgLatency.value == null) return ''
  if (avgLatency.value >= 1000) return 'text-red-700'
  if (avgLatency.value >= 200) return 'text-amber-700'

  return ''
})

const healthCounts = computed(() => {
  const counts = { passing: 0, warning: 0, critical: 0, unknown: 0 }

  for (const agent of agentsList.value || []) {
    if (counts[agent.health] === undefined) counts.unknown++
    else counts[agent.health]++
  }

  return counts
})

const topTalkerMax = computed(() => topTalkers.value[0]?.count || 1)
const talkersTotal = computed(() => topTalkers.value.reduce((sum, row) => sum + row.count, 0) || 1)

function attentionDotClass(level) {
  return (
    {
      error: 'bg-red-500',
      warn: 'bg-amber-500',
      info: 'bg-sky-500',
    }[level] || 'bg-theme-textLight'
  )
}
</script>
