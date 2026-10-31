<template>
  <div class="min-h-screen bg-theme-bg">
    <AppNav />

    <main class="w-full py-8 px-4 sm:px-6 lg:px-8">
      <div class="mb-8">
        <h1 class="text-3xl font-semibold text-theme-text">{{ $t('traffic_page.title') }}</h1>
        <p class="text-theme-textLight mt-2">{{ $t('traffic_page.subtitle') }}</p>
      </div>

      <section class="rounded-xl border border-theme-border bg-white shadow-sm overflow-hidden">
        <div class="border-b border-theme-border px-6 py-4">
          <h2 class="text-lg font-semibold text-theme-text">{{ $t('traffic_page.list_title') }}</h2>
          <p class="mt-1 text-sm text-theme-textLight">{{ $t('traffic_page.list_desc') }}</p>
        </div>

        <div
          v-if="errorMessage"
          class="m-6 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800"
        >
          {{ errorMessage }}
        </div>

        <div class="overflow-x-auto">
          <table class="min-w-full divide-y divide-theme-border">
            <thead class="bg-theme-hover">
              <tr>
                <th
                  class="px-4 sm:px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider"
                >
                  {{ $t('traffic_page.call') }}
                </th>
                <th
                  class="px-4 sm:px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider"
                >
                  {{ $t('traffic_page.status') }}
                </th>
                <th
                  class="px-4 sm:px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider"
                >
                  {{ $t('traffic_page.latency') }}
                </th>
                <th
                  class="px-4 sm:px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider"
                >
                  {{ $t('traffic_page.when') }}
                </th>
                <th class="px-4 sm:px-6 py-3 text-right">
                  <span class="sr-only">{{ $t('traffic_page.details') }}</span>
                </th>
              </tr>
            </thead>
            <tbody class="bg-white divide-y divide-theme-border">
              <tr v-if="loading && !calls.length">
                <td colspan="5" class="px-6 py-4 text-center text-sm text-theme-textLight">
                  {{ $t('common.loading') }}
                </td>
              </tr>
              <tr v-else-if="!calls.length">
                <td colspan="5" class="px-6 py-4 text-center text-sm text-theme-textLight">
                  {{ $t('traffic_page.empty') }}
                </td>
              </tr>
              <template v-else v-for="call in calls" :key="call.id">
                <tr
                  class="cursor-pointer hover:bg-theme-hover"
                  :class="{ 'bg-red-50/50 hover:bg-red-50/80': isBad(call) }"
                  @click="toggle(call.id)"
                >
                  <td class="px-4 sm:px-6 py-4 whitespace-nowrap text-sm font-mono">
                    <router-link
                      :to="{ name: 'Agent', params: { name: call.sourceAgent } }"
                      class="font-medium text-primary-800 hover:underline"
                      @click.stop
                    >
                      {{ call.sourceAgent }}
                    </router-link>
                    <span class="mx-1 text-theme-textLight" aria-hidden="true">→</span>
                    <router-link
                      v-if="call.destinationAgent"
                      :to="{ name: 'Agent', params: { name: call.destinationAgent } }"
                      class="font-medium text-primary-800 hover:underline"
                      @click.stop
                    >
                      {{ call.destinationAgent }}
                    </router-link>
                    <span v-else class="text-theme-textLight">—</span>
                  </td>
                  <td class="px-4 sm:px-6 py-4 whitespace-nowrap text-sm">
                    <span
                      class="inline-flex rounded-full px-2 py-0.5 text-xs font-medium"
                      :class="resultClass(statusOf(call))"
                    >
                      {{ $t(`traffic_page.result_${statusOf(call)}`) }}
                    </span>
                  </td>
                  <td
                    class="px-4 sm:px-6 py-4 whitespace-nowrap text-sm tabular-nums text-theme-text"
                    :class="latencyClass(call.latencyMs)"
                  >
                    {{ call.latencyMs != null ? `${call.latencyMs} ms` : '—' }}
                  </td>
                  <td class="px-4 sm:px-6 py-4 whitespace-nowrap text-sm text-theme-textLight">
                    <time :datetime="call.createdAt" :title="formatTimestamp(call.createdAt)">{{
                      formatRelative(call.createdAt)
                    }}</time>
                  </td>
                  <td class="px-4 sm:px-6 py-4 text-right">
                    <svg
                      class="ml-auto h-4 w-4 text-theme-textLight transition-transform"
                      :class="{ 'rotate-180': expandedId === call.id }"
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke="currentColor"
                      aria-hidden="true"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M19 9l-7 7-7-7"
                      />
                    </svg>
                  </td>
                </tr>
                <tr v-if="expandedId === call.id" class="bg-theme-bg/50">
                  <td colspan="5" class="px-4 sm:px-6 py-4">
                    <dl class="grid grid-cols-2 gap-x-6 gap-y-3 sm:grid-cols-3">
                      <div class="min-w-0">
                        <dt class="text-xs text-theme-textLight">
                          {{ $t('traffic_page.source_instance') }}
                        </dt>
                        <dd class="mt-0.5 truncate font-mono text-sm text-theme-text">
                          {{ call.sourceInstance || '—' }}
                        </dd>
                      </div>
                      <div class="min-w-0">
                        <dt class="text-xs text-theme-textLight">{{ $t('traffic_page.verb') }}</dt>
                        <dd class="mt-0.5 truncate font-mono text-sm text-theme-text">
                          {{ call.verb }}
                        </dd>
                      </div>
                      <div class="min-w-0">
                        <dt class="text-xs text-theme-textLight">{{ $t('traffic_page.skill') }}</dt>
                        <dd class="mt-0.5 truncate font-mono text-sm text-theme-text">
                          {{ call.skill || '—' }}
                        </dd>
                      </div>
                      <div class="min-w-0">
                        <dt class="text-xs text-theme-textLight">{{ $t('traffic_page.task') }}</dt>
                        <dd class="mt-0.5 truncate font-mono text-sm text-theme-text">
                          {{ call.taskId || '—' }}
                        </dd>
                      </div>
                      <div class="min-w-0">
                        <dt class="text-xs text-theme-textLight">
                          {{ $t('traffic_page.target_instance') }}
                        </dt>
                        <dd class="mt-0.5 truncate font-mono text-sm text-theme-text">
                          {{ call.destinationInstance || '—' }}
                        </dd>
                      </div>
                      <div class="min-w-0">
                        <dt class="text-xs text-theme-textLight">
                          {{ $t('traffic_page.recorded_at') }}
                        </dt>
                        <dd class="mt-0.5 text-sm text-theme-text">
                          {{ formatTimestamp(call.createdAt) }}
                        </dd>
                      </div>
                    </dl>
                  </td>
                </tr>
              </template>
            </tbody>
          </table>
        </div>

        <div
          v-if="total > 0"
          class="bg-white px-6 py-4 border-t border-theme-border flex items-center justify-between"
        >
          <div class="text-sm text-theme-textLight">
            {{
              $t('traffic_page.showing', {
                from: offset + 1,
                to: Math.min(offset + limit, total),
                total,
              })
            }}
          </div>
          <div class="flex items-center gap-2">
            <button
              type="button"
              class="btn-secondary text-sm disabled:opacity-50 disabled:cursor-not-allowed"
              :disabled="offset === 0"
              @click="goToPage(offset - limit)"
            >
              {{ $t('traffic_page.previous') }}
            </button>
            <button
              type="button"
              class="btn-secondary text-sm disabled:opacity-50 disabled:cursor-not-allowed"
              :disabled="offset + limit >= total"
              @click="goToPage(offset + limit)"
            >
              {{ $t('traffic_page.next') }}
            </button>
          </div>
        </div>
      </section>
    </main>
  </div>
</template>

<script setup>
import { onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import AppNav from '@/comp/AppNav.vue'
import { traffic_api } from '@/api'
import { useUrlPagination } from '@/lib/pagination'
import { usePolling } from '@/lib/polling'
import { useWorkspaceContext } from '@/lib/permission'
import { errorText } from '@/lib/error'
import { formatTimestamp, formatRelative } from '@/utils/date'

const { t } = useI18n()
const route = useRoute()
const { currentWorkspace } = useWorkspaceContext()
const { offset, limit, goToOffset } = useUrlPagination({ limit: 50 })

const calls = ref([])
const total = ref(0)
const loading = ref(false)
const errorMessage = ref(null)
const expandedId = ref(null)

function resultClass(result) {
  return (
    {
      completed: 'bg-emerald-50 text-emerald-800',
      failed: 'bg-red-50 text-red-800',
      denied: 'bg-red-50 text-red-800',
      timeout: 'bg-amber-50 text-amber-800',
      no_instance: 'bg-amber-50 text-amber-800',
    }[result] || 'bg-theme-hover text-theme-textLight'
  )
}

// A denied call never completes, so decision and result fold into one status.
function statusOf(call) {
  return call.decision === 'deny' ? 'denied' : call.result
}

function isBad(call) {
  return call.decision === 'deny' || call.result === 'failed'
}

function latencyClass(ms) {
  if (ms == null) return ''
  if (ms >= 1000) return 'text-red-700 font-medium'
  if (ms >= 200) return 'text-amber-700'
  return ''
}

function toggle(id) {
  expandedId.value = expandedId.value === id ? null : id
}

async function loadTraffic() {
  loading.value = true
  errorMessage.value = null

  try {
    const res = await traffic_api.list(currentWorkspace.id, { limit: limit, offset: offset.value })
    calls.value = res.data.calls || []
    total.value = res.data._meta?.total || 0
  } catch (err) {
    errorMessage.value = errorText(err, t('traffic_page.failed_load'))
    calls.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

function goToPage(newOffset) {
  if (newOffset < 0 || newOffset >= total.value) return
  goToOffset(newOffset)
}

watch(() => route.query.page, loadTraffic)
onMounted(loadTraffic)

// New rows land on the first page, so only that page follows them live
usePolling(() => {
  if (offset.value === 0) loadTraffic()
})
</script>
