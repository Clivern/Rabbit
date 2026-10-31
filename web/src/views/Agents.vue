<template>
  <div class="min-h-screen bg-theme-bg">
    <AppNav />

    <main class="w-full py-8 px-4 sm:px-6 lg:px-8">
      <div class="mb-8">
        <h1 class="text-3xl font-semibold text-theme-text">{{ $t('agents_page.title') }}</h1>
        <p class="text-theme-textLight mt-2">{{ $t('agents_page.subtitle') }}</p>
      </div>

      <section class="rounded-xl border border-theme-border bg-white shadow-sm overflow-hidden">
        <div
          class="flex flex-wrap items-start justify-between gap-4 border-b border-theme-border px-6 py-4"
        >
          <div>
            <h2 class="text-lg font-semibold text-theme-text">
              {{ $t('agents_page.list_title') }}
            </h2>
            <p class="mt-1 text-sm text-theme-textLight">{{ $t('agents_page.list_desc') }}</p>
          </div>
          <button
            v-if="canManage"
            type="button"
            class="btn-primary inline-flex items-center justify-center gap-2"
            @click="showRegisterModal = true"
          >
            <svg
              class="h-4 w-4"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              aria-hidden="true"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M12 4v16m8-8H4"
              />
            </svg>
            {{ $t('agents_page.register') }}
          </button>
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
                  {{ $t('agents_page.name') }}
                </th>
                <th
                  class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider"
                >
                  {{ $t('agents_page.version') }}
                </th>
                <th
                  class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider"
                >
                  {{ $t('agents_page.skills') }}
                </th>
                <th
                  class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider"
                >
                  {{ $t('agents_page.instances') }}
                </th>
                <th
                  class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider"
                >
                  {{ $t('agents_page.health') }}
                </th>
                <th
                  class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider"
                >
                  {{ $t('agents_page.updated') }}
                </th>
              </tr>
            </thead>
            <tbody class="bg-white divide-y divide-theme-border">
              <tr v-if="loading && !agents.length">
                <td colspan="6" class="px-6 py-4 text-center text-sm text-theme-textLight">
                  {{ $t('common.loading') }}
                </td>
              </tr>
              <tr v-else-if="!agents.length">
                <td colspan="6" class="px-6 py-4 text-center text-sm text-theme-textLight">
                  {{ $t('agents_page.no_agents') }}
                </td>
              </tr>
              <template v-else>
                <tr v-for="agent in rows" :key="agent.id" class="hover:bg-theme-hover">
                  <td class="px-6 py-4">
                    <router-link
                      :to="{ name: 'Agent', params: { name: agent.name } }"
                      class="text-sm font-medium text-theme-text font-mono hover:underline"
                    >
                      {{ agent.name }}
                    </router-link>
                    <p v-if="agent.description" class="mt-0.5 text-sm text-theme-textLight">
                      {{ agent.description }}
                    </p>
                    <span
                      v-if="agent.cardPending"
                      class="mt-1 inline-flex items-center rounded-full bg-amber-50 px-2 py-0.5 text-xs font-medium text-amber-800"
                    >
                      {{ $t('agents_page.card_pending') }}
                    </span>
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap text-sm text-theme-text font-mono">
                    {{ agent.version }}
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap text-sm text-theme-text">
                    {{ $t('agents_page.skills_count', agent.skillCount) }}
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap text-sm text-theme-text">
                    <StatusBreakdown :counts="agent.statusCounts" />
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap text-sm">
                    <span
                      class="inline-flex items-center gap-1.5 rounded-full px-2 py-0.5 text-xs font-medium"
                      :class="healthClass(agent.health)"
                    >
                      <span
                        class="h-1.5 w-1.5 rounded-full"
                        :class="healthDotClass(agent.health)"
                        aria-hidden="true"
                      />
                      {{ $t(`agents_page.health_${agent.health}`) }}
                    </span>
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap text-sm text-theme-textLight">
                    {{ formatRelative(agent.updatedAt) }}
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
              $t('agents_page.showing', {
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
              {{ $t('agents_page.previous') }}
            </button>
            <button
              type="button"
              class="btn-secondary text-sm disabled:opacity-50 disabled:cursor-not-allowed"
              :disabled="offset + limit >= total"
              @click="goToPage(offset + limit)"
            >
              {{ $t('agents_page.next') }}
            </button>
          </div>
        </div>
      </section>
    </main>

    <AgentCardModal
      :open="showRegisterModal"
      :workspace-id="currentWorkspace.id"
      @close="showRegisterModal = false"
      @saved="onRegistered"
    />
  </div>
</template>

<script setup>
import StatusBreakdown from '@/comp/StatusBreakdown.vue'
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import AppNav from '@/comp/AppNav.vue'
import AgentCardModal from '@/comp/AgentCardModal.vue'
import { agent_api } from '@/api'
import { showFlash } from '@/lib/flash'
import { useUrlPagination } from '@/lib/pagination'
import { usePolling } from '@/lib/polling'
import { useWorkspaceContext } from '@/lib/permission'
import { healthClass, healthDotClass, parseCard } from '@/lib/agent'
import { formatRelative } from '@/utils/date'
import { errorText } from '@/lib/error'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const { currentWorkspace, canManage } = useWorkspaceContext()
const { offset, limit, goToOffset } = useUrlPagination({ limit: 50 })

const agents = ref([])
const total = ref(0)
const loading = ref(false)
const errorMessage = ref(null)
const showRegisterModal = ref(false)

const rows = computed(() =>
  agents.value.map((agent) => {
    const card = parseCard(agent.card)
    return {
      id: agent.id,
      name: agent.name,
      version: agent.version,
      description: card.description,
      cardPending: !agent.card,
      skillCount: card.skills.length,
      statusCounts: agent.statusCounts,
      health: agent.health,
      updatedAt: agent.updatedAt,
    }
  })
)

async function loadPage() {
  loading.value = true
  errorMessage.value = null

  try {
    const res = await agent_api.list(currentWorkspace.id, { limit, offset: offset.value })
    agents.value = res.data.agents || []
    total.value = res.data._meta?.total || 0
  } catch (err) {
    errorMessage.value = errorText(err, t('agents_page.failed_load'))
    agents.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

function goToPage(newOffset) {
  if (newOffset < 0 || newOffset >= total.value) return
  goToOffset(newOffset)
}

function onRegistered(agent) {
  showRegisterModal.value = false
  showFlash(t('common.created'))
  router.push({ name: 'Agent', params: { name: agent.name } })
}

watch(() => route.query.page, loadPage)

onMounted(loadPage)

// Agent health changes as instances report in, so keep the list live
usePolling(loadPage)
</script>
