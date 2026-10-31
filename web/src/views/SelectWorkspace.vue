<template>
  <div class="min-h-screen bg-theme-bg flex flex-col items-center justify-center px-4 py-12">
    <div class="w-full max-w-md">
      <div class="flex justify-center mb-6">
        <AppLogo
          variant="hanging"
          :alt="$t('common.logo_alt')"
          class="h-36 w-auto text-theme-text"
        />
      </div>

      <h1 class="text-xl font-semibold text-theme-text text-center">
        {{ pageTitle }}
      </h1>
      <p class="mt-2 text-sm text-theme-textLight text-center">
        {{ pageDescription }}
      </p>

      <div v-if="workspaces.length" class="mt-6 space-y-3">
        <div
          v-for="workspace in workspaces"
          :key="workspace.id"
          class="rounded-lg border border-theme-border bg-white shadow-sm overflow-hidden"
        >
          <button
            type="button"
            @click="toggleWorkspace(workspace)"
            :aria-expanded="expandedId === workspace.id"
            class="w-full flex items-center justify-between px-4 py-3 text-left hover:bg-theme-hover transition-colors focus:outline-none focus:ring-2 focus:ring-primary-800 focus:ring-inset min-w-0"
          >
            <div class="flex min-w-0 flex-1 items-center gap-3">
              <div class="min-w-0 flex-1">
                <p class="text-sm font-semibold text-theme-text truncate">{{ workspace.name }}</p>
                <p class="text-xs text-theme-textLight mt-0.5">
                  {{ formatDate(workspace.createdAt) }}
                </p>
              </div>
              <span
                v-if="workspace.status === 'active' || storedWorkspace?.id === workspace.id"
                class="inline-flex shrink-0 rounded-full bg-primary-200 px-2.5 py-0.5 text-xs font-medium text-theme-text"
              >
                {{ $t('select_workspace.active') }}
              </span>
            </div>
            <svg
              class="ml-3 h-5 w-5 shrink-0 text-theme-textLight transition-transform"
              :class="{ 'rotate-90': expandedId === workspace.id }"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              aria-hidden="true"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M9 5l7 7-7 7"
              />
            </svg>
          </button>
          <div
            v-if="expandedId === workspace.id"
            class="grid grid-cols-1 gap-2 border-t border-theme-border px-4 py-3 sm:grid-cols-3"
            role="group"
            :aria-label="$t('select_workspace.gateway')"
          >
            <button
              v-for="gateway in gateways"
              :key="gateway.id"
              type="button"
              :disabled="gateway.disabled"
              class="group rounded-md border border-theme-border px-3 py-2 text-sm font-medium text-theme-text transition-colors hover:bg-theme-hover focus:outline-none focus:ring-2 focus:ring-primary-800 focus:ring-offset-1 disabled:cursor-not-allowed disabled:opacity-50 disabled:hover:bg-transparent"
              @click="selectWorkspace(workspace)"
            >
              <span :class="{ 'group-hover:hidden': gateway.disabled }">{{ gateway.label }}</span>
              <span v-if="gateway.disabled" class="hidden group-hover:inline">{{
                $t('select_workspace.coming_soon')
              }}</span>
            </button>
          </div>
        </div>
      </div>

      <div class="mt-6">
        <button
          type="button"
          @click="openCreateModal"
          :class="[
            'w-full flex items-center justify-center gap-2 rounded-lg px-4 py-3 text-sm font-medium focus:outline-none focus:ring-2 focus:ring-primary-800 focus:ring-offset-1 transition-colors',
            workspaces.length
              ? 'border border-theme-border bg-white text-theme-text hover:bg-theme-hover'
              : 'border-2 border-dashed border-primary-400 bg-primary-50/50 text-primary-800 hover:bg-primary-100',
          ]"
        >
          <svg
            class="h-5 w-5"
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
          {{
            workspaces.length
              ? $t('select_workspace.create_new')
              : $t('select_workspace.create_first')
          }}
        </button>
      </div>

      <p v-if="error" class="mt-4 text-sm text-red-600 text-center">{{ error }}</p>
    </div>

    <BaseModal
      :open="showCreateModal"
      :title="$t('select_workspace.create_modal_title')"
      :description="$t('select_workspace.create_modal_description')"
      :error="createModalError"
      size="sm"
      @close="closeCreateModal"
    >
      <form @submit.prevent="handleCreate" class="mt-4 space-y-4">
        <div>
          <label for="workspace-name" class="form-label">{{
            $t('select_workspace.workspace_name')
          }}</label>
          <input
            id="workspace-name"
            v-model="newWorkspaceName"
            type="text"
            required
            class="input-field"
            :placeholder="$t('select_workspace.workspace_name_placeholder')"
            autofocus
          />
        </div>
        <div class="flex gap-3 justify-end">
          <button type="button" @click="closeCreateModal" class="btn-secondary">
            {{ $t('common.cancel') }}
          </button>
          <button type="submit" :disabled="createLoading" class="btn-primary">
            {{ createLoading ? $t('common.loading') : $t('common.create') }}
          </button>
        </div>
      </form>
    </BaseModal>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { workspace_api } from '@/api'
import { loadWorkspaceFromStorage, saveWorkspaceToStorage } from '@/utils/storage'
import AppLogo from '@/comp/AppLogo.vue'
import { errorText } from '@/lib/error'
import { formatDate } from '@/utils/date'
import BaseModal from '@/comp/BaseModal.vue'

const { t } = useI18n()
const router = useRouter()
const storedWorkspace = loadWorkspaceFromStorage()
const workspaces = ref([])
const error = ref(null)
const showCreateModal = ref(false)
const createModalError = ref(null)
const newWorkspaceName = ref('')
const createLoading = ref(false)
const expandedId = ref(null)

const gateways = computed(() => [
  { id: 'ai', label: t('select_workspace.ai_gateway') },
  { id: 'api', label: t('select_workspace.api_gateway'), disabled: true },
  { id: 'mcp', label: t('select_workspace.mcp_gateway'), disabled: true },
])

const pageTitle = computed(() => {
  if (workspaces.value.length) {
    return t('select_workspace.select_title')
  }
  return t('select_workspace.create_title')
})

const pageDescription = computed(() => {
  if (workspaces.value.length) {
    return t('select_workspace.select_description')
  }
  return t('select_workspace.no_workspaces')
})

function openCreateModal() {
  createModalError.value = null
  showCreateModal.value = true
}

function closeCreateModal() {
  showCreateModal.value = false
  createModalError.value = null
  newWorkspaceName.value = ''
}

function toggleWorkspace(workspace) {
  expandedId.value = expandedId.value === workspace.id ? null : workspace.id
}

function selectWorkspace(workspace) {
  saveWorkspaceToStorage(workspace)
  router.push('/dashboard')
}

async function loadWorkspaces() {
  error.value = null
  try {
    const res = await workspace_api.list({})
    workspaces.value = res.data?.workspaces ?? []
  } catch (err) {
    error.value = errorText(err, t('select_workspace.failed_load'))
    workspaces.value = []
  }
}

async function handleCreate() {
  createLoading.value = true
  createModalError.value = null

  try {
    const res = await workspace_api.create({ name: newWorkspaceName.value })
    if (res.data?.id) {
      const workspace = { ...res.data, role: 'owner', status: 'active' }
      saveWorkspaceToStorage(workspace)
      closeCreateModal()
      await loadWorkspaces()
      router.push('/dashboard')
    } else {
      createModalError.value = 'Invalid response'
    }
  } catch (err) {
    createModalError.value = errorText(err, 'Failed to create workspace')
  } finally {
    createLoading.value = false
  }
}

onMounted(async () => {
  await loadWorkspaces()
  if (workspaces.value.length === 0 && !error.value) openCreateModal()
})
</script>
