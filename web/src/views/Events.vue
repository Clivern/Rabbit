<template>
  <div class="min-h-screen bg-theme-bg">
    <AppNav />

    <main class="w-full py-8 px-4 sm:px-6 lg:px-8">
      <div class="mb-8">
        <h1 class="text-3xl font-semibold text-theme-text">{{ $t('events_page.title') }}</h1>
        <p class="text-theme-textLight mt-2">{{ $t('events_page.subtitle') }}</p>
      </div>

      <section class="rounded-xl border border-theme-border bg-white shadow-sm overflow-hidden">
        <div class="border-b border-theme-border px-6 py-4">
          <h2 class="text-lg font-semibold text-theme-text">{{ $t('events_page.list_title') }}</h2>
          <p class="mt-1 text-sm text-theme-textLight">{{ $t('events_page.list_desc') }}</p>
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
                  class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider"
                >
                  {{ $t('events_page.event') }}
                </th>
                <th
                  class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider"
                >
                  {{ $t('events_page.actor') }}
                </th>
                <th
                  class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider"
                >
                  {{ $t('events_page.resource') }}
                </th>
                <th
                  class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider"
                >
                  {{ $t('events_page.source') }}
                </th>
                <th
                  class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider"
                >
                  {{ $t('events_page.time') }}
                </th>
                <th class="px-6 py-3 text-right">
                  <span class="sr-only">{{ $t('events_page.details') }}</span>
                </th>
              </tr>
            </thead>
            <tbody class="bg-white divide-y divide-theme-border">
              <tr v-if="loading && !events.length">
                <td colspan="6" class="px-6 py-4 text-center text-sm text-theme-textLight">
                  {{ $t('common.loading') }}
                </td>
              </tr>
              <tr v-else-if="!events.length">
                <td colspan="6" class="px-6 py-4 text-center text-sm text-theme-textLight">
                  {{ $t('events_page.empty') }}
                </td>
              </tr>
              <template v-else>
                <template v-for="item in events" :key="item.id">
                  <tr class="cursor-pointer hover:bg-theme-hover" @click="toggle(item.id)">
                    <td class="px-6 py-4 whitespace-nowrap text-sm">
                      <span
                        class="inline-flex rounded-full bg-theme-hover px-2 py-0.5 text-xs font-medium text-theme-text"
                      >
                        {{ $t(`integrations_page.event_${item.type.replaceAll('.', '_')}`) }}
                      </span>
                    </td>
                    <td class="px-6 py-4 whitespace-nowrap text-sm text-theme-text">
                      <span v-if="item.actorName || item.actorType">
                        {{ item.actorName || '—' }}
                        <span v-if="item.actorType" class="ml-1 text-xs text-theme-textLight">{{
                          $t(`events_page.actor_${item.actorType}`)
                        }}</span>
                      </span>
                      <span v-else class="text-theme-textLight">—</span>
                    </td>
                    <td class="px-6 py-4 whitespace-nowrap text-sm text-theme-text">
                      <span v-if="item.resourceType">
                        {{ item.resourceType }}
                        <span
                          v-if="item.resourceId"
                          class="ml-1 font-mono text-xs text-theme-textLight"
                          >{{ item.resourceId.slice(0, 8) }}</span
                        >
                      </span>
                      <span v-else class="text-theme-textLight">—</span>
                    </td>
                    <td class="px-6 py-4 whitespace-nowrap text-sm text-theme-text font-mono">
                      {{ item.ipAddress || '—' }}
                    </td>
                    <td class="px-6 py-4 whitespace-nowrap text-sm text-theme-textLight">
                      <time :datetime="item.createdAt" :title="formatTimestamp(item.createdAt)">{{
                        formatRelative(item.createdAt)
                      }}</time>
                    </td>
                    <td class="px-6 py-4 text-right">
                      <svg
                        class="ml-auto h-4 w-4 text-theme-textLight transition-transform"
                        :class="{ 'rotate-180': expandedId === item.id }"
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
                  <tr v-if="expandedId === item.id" class="bg-theme-hover">
                    <td colspan="6" class="px-6 py-4">
                      <dl class="grid gap-x-6 gap-y-3 text-sm sm:grid-cols-2">
                        <div>
                          <dt
                            class="text-xs font-medium uppercase tracking-wider text-theme-textLight"
                          >
                            {{ $t('events_page.event_id') }}
                          </dt>
                          <dd class="mt-1 font-mono text-theme-text break-all">{{ item.id }}</dd>
                        </div>
                        <div v-if="item.actorId">
                          <dt
                            class="text-xs font-medium uppercase tracking-wider text-theme-textLight"
                          >
                            {{ $t('events_page.actor_id') }}
                          </dt>
                          <dd class="mt-1 font-mono text-theme-text break-all">
                            {{ item.actorId }}
                          </dd>
                        </div>
                        <div v-if="item.resourceId">
                          <dt
                            class="text-xs font-medium uppercase tracking-wider text-theme-textLight"
                          >
                            {{ $t('events_page.resource_id') }}
                          </dt>
                          <dd class="mt-1 font-mono text-theme-text break-all">
                            {{ item.resourceId }}
                          </dd>
                        </div>
                        <div v-if="item.userAgent">
                          <dt
                            class="text-xs font-medium uppercase tracking-wider text-theme-textLight"
                          >
                            {{ $t('events_page.user_agent') }}
                          </dt>
                          <dd class="mt-1 text-theme-text break-all">{{ item.userAgent }}</dd>
                        </div>
                        <div class="sm:col-span-2">
                          <dt
                            class="text-xs font-medium uppercase tracking-wider text-theme-textLight"
                          >
                            {{ $t('events_page.data') }}
                          </dt>
                          <dd class="mt-1">
                            <pre
                              class="overflow-x-auto rounded-md border border-theme-border bg-white p-3 text-xs font-mono text-theme-text"
                              >{{ JSON.stringify(item.data || {}, null, 2) }}</pre>
                          </dd>
                        </div>
                      </dl>
                    </td>
                  </tr>
                </template>
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
              $t('events_page.showing', {
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
              {{ $t('events_page.previous') }}
            </button>
            <button
              type="button"
              class="btn-secondary text-sm disabled:opacity-50 disabled:cursor-not-allowed"
              :disabled="offset + limit >= total"
              @click="goToPage(offset + limit)"
            >
              {{ $t('events_page.next') }}
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
import { event_api } from '@/api'
import { useUrlPagination } from '@/lib/pagination'
import { usePolling } from '@/lib/polling'
import { useWorkspaceContext } from '@/lib/permission'
import { errorText } from '@/lib/error'
import { formatTimestamp, formatRelative } from '@/utils/date'

const { t } = useI18n()
const route = useRoute()
const { currentWorkspace } = useWorkspaceContext()
const { offset, limit, goToOffset } = useUrlPagination({ limit: 50 })

const events = ref([])
const total = ref(0)
const loading = ref(false)
const errorMessage = ref(null)
const expandedId = ref(null)

async function loadEvents() {
  loading.value = true
  errorMessage.value = null
  expandedId.value = null

  try {
    const res = await event_api.list(currentWorkspace.id, { limit: limit, offset: offset.value })
    events.value = res.data.events || []
    total.value = res.data._meta?.total || 0
  } catch (err) {
    errorMessage.value = errorText(err, t('events_page.failed_load'))
    events.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

function toggle(id) {
  expandedId.value = expandedId.value === id ? null : id
}

function goToPage(newOffset) {
  if (newOffset < 0 || newOffset >= total.value) return
  goToOffset(newOffset)
}

watch(() => route.query.page, loadEvents)
onMounted(loadEvents)

// New rows land on the first page, so only that page follows them live
usePolling(() => {
  if (offset.value === 0) loadEvents()
})
</script>
