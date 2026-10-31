<template>
  <div class="min-h-screen bg-theme-bg">
    <AppNav />

    <main class="w-full py-8 px-4 sm:px-6 lg:px-8">
      <div
        v-if="loading"
        class="mt-8 rounded-xl border border-theme-border bg-white p-12 text-center shadow-sm"
      >
        <svg
          class="animate-spin h-8 w-8 mx-auto text-theme-text"
          fill="none"
          viewBox="0 0 24 24"
          aria-hidden="true"
        >
          <circle
            class="opacity-25"
            cx="12"
            cy="12"
            r="10"
            stroke="currentColor"
            stroke-width="4"
          />
          <path
            class="opacity-75"
            fill="currentColor"
            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
          />
        </svg>
        <p class="text-theme-textLight mt-3">{{ $t('agents_page.loading_agent') }}</p>
      </div>

      <div
        v-else-if="notFound"
        class="mt-8 rounded-xl border border-theme-border bg-white p-8 shadow-sm"
      >
        <h1 class="text-2xl font-semibold text-theme-text">{{ $t('agents_page.not_found') }}</h1>
      </div>

      <div
        v-else-if="!agent"
        class="mt-8 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800"
      >
        {{ errorMessage }}
      </div>

      <template v-else>
        <div class="mb-6 flex flex-wrap items-start justify-between gap-4">
          <div class="min-w-0">
            <div class="flex flex-wrap items-center gap-3">
              <h1 class="text-3xl font-semibold text-theme-text font-mono break-all">
                {{ agent.name }}
              </h1>
              <span
                v-if="agent.health"
                class="inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-medium"
                :class="healthClass(agent.health)"
              >
                <span
                  class="h-1.5 w-1.5 rounded-full"
                  :class="healthDotClass(agent.health)"
                  aria-hidden="true"
                />
                {{ $t(`agents_page.health_${agent.health}`) }}
              </span>
              <span
                class="rounded-full border border-theme-border bg-white px-2.5 py-1 text-xs font-mono text-theme-textLight"
              >
                v{{ agent.version }}
              </span>
            </div>
            <p v-if="card.description" class="mt-2 max-w-3xl text-theme-textLight">
              {{ card.description }}
            </p>
          </div>
          <div v-if="canManage" class="flex items-center gap-2">
            <button type="button" class="btn-secondary" @click="showEditModal = true">
              {{ $t('agents_page.edit_card') }}
            </button>
            <button
              type="button"
              class="btn-primary bg-red-600 hover:bg-red-700 focus:ring-red-500"
              @click="showDeleteModal = true"
            >
              {{ $t('agents_page.delete') }}
            </button>
          </div>
        </div>

        <div
          v-if="errorMessage"
          class="mb-6 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800"
        >
          {{ errorMessage }}
        </div>

        <div class="mb-6 grid grid-cols-2 gap-4 lg:grid-cols-4">
          <a
            href="#instances"
            class="rounded-xl border border-theme-border bg-white p-4 shadow-sm transition-colors hover:border-primary-300"
          >
            <p class="text-xs font-medium uppercase tracking-wider text-theme-textLight">
              {{ $t('agents_page.instances') }}
            </p>
            <p class="mt-1 text-2xl font-semibold text-theme-text">{{ instancesTotal }}</p>
            <div class="mt-1 min-h-5">
              <StatusBreakdown v-if="instancesTotal" :counts="agent.statusCounts || {}" hide-zero />
            </div>
          </a>
          <a
            href="#checks"
            class="rounded-xl border border-theme-border bg-white p-4 shadow-sm transition-colors hover:border-primary-300"
          >
            <p class="text-xs font-medium uppercase tracking-wider text-theme-textLight">
              {{ $t('agents_page.checks_title') }}
            </p>
            <p class="mt-1 text-2xl font-semibold text-theme-text">{{ checks.length }}</p>
            <p class="mt-1 text-sm text-theme-textLight">{{ checkTypesSummary }}</p>
          </a>
          <a
            href="#protection"
            class="rounded-xl border border-theme-border bg-white p-4 shadow-sm transition-colors hover:border-primary-300"
          >
            <p class="text-xs font-medium uppercase tracking-wider text-theme-textLight">
              {{ $t('agents_page.protection_title') }}
            </p>
            <p
              class="mt-1 text-2xl font-semibold"
              :class="protection?.isActive ? 'text-emerald-700' : 'text-theme-text'"
            >
              {{
                protection
                  ? protection.isActive
                    ? $t('agents_page.protection_active')
                    : $t('agents_page.protection_inactive')
                  : $t('agents_page.protection_not_set')
              }}
            </p>
            <p class="mt-1 text-sm text-theme-textLight">
              {{
                protection
                  ? $t(`agents_page.protection_type_${protection.type}`)
                  : $t('agents_page.protection_not_set_hint')
              }}
            </p>
          </a>
          <a
            href="#identities"
            class="rounded-xl border border-theme-border bg-white p-4 shadow-sm transition-colors hover:border-primary-300"
          >
            <p class="text-xs font-medium uppercase tracking-wider text-theme-textLight">
              {{ $t('agents_page.identities_title') }}
            </p>
            <p class="mt-1 text-2xl font-semibold text-theme-text">{{ identities.length }}</p>
            <p class="mt-1 text-sm text-theme-textLight">
              {{
                $t('agents_page.identities_active_count', {
                  count: identities.filter((identity) => identity.isActive).length,
                })
              }}
            </p>
          </a>
        </div>

        <div class="grid gap-6 lg:grid-cols-3">
          <div class="min-w-0 space-y-6 lg:col-span-2">
            <AgentTraffic :workspace-id="currentWorkspace.id" :agent-name="agent.name" />
            <AgentInstances
              ref="instancesSection"
              v-model:total="instancesTotal"
              :workspace-id="currentWorkspace.id"
              :agent-name="agent.name"
              :can-manage="canManage"
              @error="errorMessage = $event"
              @deregistered="refreshAgent"
            />
            <AgentChecks
              v-model:checks="checks"
              :workspace-id="currentWorkspace.id"
              :agent-name="agent.name"
              :can-manage="canManage"
              @error="errorMessage = $event"
              @changed="refreshHealth"
            />
            <AgentSkills :skills="card.skills" />
          </div>

          <div class="min-w-0 space-y-6">
            <AgentCardDetails :agent="agent" />
            <AgentProtection
              v-model:protection="protection"
              :workspace-id="currentWorkspace.id"
              :agent-name="agent.name"
              :can-manage="canManage"
              @error="errorMessage = $event"
            />
            <AgentIdentities
              v-model:identities="identities"
              :workspace-id="currentWorkspace.id"
              :agent-name="agent.name"
              :can-manage="canManage"
              @error="errorMessage = $event"
            />
          </div>
        </div>
      </template>
    </main>

    <AgentCardModal
      v-if="agent"
      :open="showEditModal"
      :workspace-id="currentWorkspace.id"
      :agent="agent"
      @close="showEditModal = false"
      @saved="onCardSaved"
    />

    <ConfirmModal
      :open="showDeleteModal"
      :title="$t('agents_page.delete_title')"
      :message="$t('agents_page.delete_agent_confirm', { name: agent?.name })"
      :error="deleteError"
      :busy="deleting"
      :disabled="deleteConfirmName.trim() !== agent?.name"
      :confirm-label="$t('agents_page.delete')"
      :busy-label="$t('agents_page.deleting')"
      @cancel="closeConfirm"
      @confirm="confirmDelete"
    >
      <div class="mt-4">
        <label for="delete-agent-confirm" class="block text-sm font-medium text-theme-text mb-1.5">
          {{ $t('agents_page.type_to_confirm', { name: agent?.name }) }}
        </label>
        <input
          id="delete-agent-confirm"
          v-model="deleteConfirmName"
          type="text"
          class="input-field font-mono"
          :placeholder="agent?.name"
          autocomplete="off"
          :disabled="deleting"
        />
      </div>
    </ConfirmModal>
  </div>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import AppNav from '@/comp/AppNav.vue'
import AgentCardDetails from '@/comp/AgentCardDetails.vue'
import AgentCardModal from '@/comp/AgentCardModal.vue'
import AgentChecks from '@/comp/AgentChecks.vue'
import AgentIdentities from '@/comp/AgentIdentities.vue'
import AgentInstances from '@/comp/AgentInstances.vue'
import AgentProtection from '@/comp/AgentProtection.vue'
import AgentSkills from '@/comp/AgentSkills.vue'
import AgentTraffic from '@/comp/AgentTraffic.vue'
import ConfirmModal from '@/comp/ConfirmModal.vue'
import StatusBreakdown from '@/comp/StatusBreakdown.vue'
import { agent_api } from '@/api'
import { showFlash } from '@/lib/flash'
import { usePolling } from '@/lib/polling'
import { useWorkspaceContext } from '@/lib/permission'
import { healthClass, healthDotClass, parseCard } from '@/lib/agent'
import { errorText } from '@/lib/error'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const { currentWorkspace, canManage } = useWorkspaceContext()

const agent = ref(null)
const loading = ref(false)
const notFound = ref(false)
const errorMessage = ref(null)
const instancesSection = ref(null)

// Filled by the sections through v-model so the summary cards can count them
const instancesTotal = ref(0)
const checks = ref([])
const protection = ref(null)
const identities = ref([])

const showEditModal = ref(false)
const showDeleteModal = ref(false)
const deleteConfirmName = ref('')
const deleting = ref(false)
const deleteError = ref(null)

const agentName = computed(() => String(route.params.name || ''))
const card = computed(() => parseCard(agent.value?.card))
const checkTypesSummary = computed(() => {
  const types = [...new Set(checks.value.map((check) => t(`agents_page.check_type_${check.type}`)))]
  return types.length ? types.join(', ') : t('agents_page.no_checks_short')
})

async function loadAgent() {
  loading.value = true
  notFound.value = false
  errorMessage.value = null
  agent.value = null
  instancesTotal.value = 0
  checks.value = []
  protection.value = null
  identities.value = []

  try {
    const res = await agent_api.get(currentWorkspace.id, agentName.value)
    agent.value = res.data
  } catch (err) {
    if (err.response?.status === 404) {
      notFound.value = true
    } else {
      errorMessage.value = errorText(err, t('agents_page.failed_load_agent'))
    }
  } finally {
    loading.value = false
  }
}

// Health is derived from instances and checks, so reload the agent without the page spinner
async function refreshAgent() {
  try {
    const res = await agent_api.get(currentWorkspace.id, agentName.value)
    agent.value = res.data
  } catch (err) {
    errorMessage.value = errorText(err, t('agents_page.failed_load_agent'))
  }
}

async function refreshHealth() {
  await Promise.all([refreshAgent(), instancesSection.value?.load()])
}

function onCardSaved(updated) {
  showEditModal.value = false
  agent.value = updated
  showFlash(t('common.saved'))
}

function closeConfirm() {
  showDeleteModal.value = false
  deleteConfirmName.value = ''
  deleteError.value = null
}

async function confirmDelete() {
  deleting.value = true
  deleteError.value = null

  try {
    await agent_api.delete(currentWorkspace.id, agent.value.name)
    closeConfirm()
    showFlash(t('common.deleted'))
    router.push('/agents')
  } catch (err) {
    deleteError.value = errorText(err, t('agents_page.failed_delete'))
  } finally {
    deleting.value = false
  }
}

// Instances report health on their own, so keep the header and the instance list live
usePolling(() => {
  if (agent.value) refreshHealth()
})

watch(agentName, loadAgent, { immediate: true })
</script>
