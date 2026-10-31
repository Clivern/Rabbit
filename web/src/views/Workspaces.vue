<template>
  <div class="min-h-screen bg-theme-bg">
    <AppNav />

    <main class="w-full py-8 px-4 sm:px-6 lg:px-8">
      <div class="mb-8">
        <h1 class="text-3xl font-semibold text-theme-text">{{ $t('workspaces.title') }}</h1>
        <p class="text-theme-textLight mt-2">{{ $t('workspaces_page.subtitle') }}</p>
      </div>

      <div
        v-if="errorMessage"
        class="mb-6 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800"
      >
        {{ errorMessage }}
      </div>

      <div class="rounded-xl border border-theme-border bg-white shadow-sm overflow-hidden">
        <div class="flex justify-end border-b border-theme-border bg-white px-6 py-4">
          <button
            type="button"
            @click="openCreateModal"
            class="btn-primary inline-flex items-center justify-center gap-2"
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
            {{ $t('workspaces_page.create') }}
          </button>
        </div>

        <div v-if="loading" class="p-12 text-center">
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
          <p class="text-theme-textLight mt-3">{{ $t('workspaces_page.loading') }}</p>
        </div>

        <div v-else-if="workspaces.length === 0" class="p-12 text-center">
          <div class="flex justify-center mb-4">
            <div class="flex h-12 w-12 items-center justify-center rounded-full bg-primary-100">
              <svg
                class="h-6 w-6 text-theme-text"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                aria-hidden="true"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="1.5"
                  d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4"
                />
              </svg>
            </div>
          </div>
          <p class="text-theme-textLight">{{ $t('workspaces_page.no_workspaces') }}</p>
          <p class="text-sm text-theme-textLight mt-1">
            {{ $t('workspaces_page.no_workspaces_hint') }}
          </p>
        </div>

        <div v-else class="overflow-x-auto">
          <table class="min-w-full divide-y divide-theme-border">
            <thead class="bg-theme-hover">
              <tr>
                <th
                  class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider"
                >
                  {{ $t('workspaces_page.name') }}
                </th>
                <th
                  class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider"
                >
                  {{ $t('workspaces_page.created') }}
                </th>
                <th
                  class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider"
                >
                  {{ $t('workspaces_page.members') }}
                </th>
                <th
                  class="px-6 py-3 text-right text-xs font-medium text-theme-textLight uppercase tracking-wider"
                >
                  {{ $t('workspaces_page.actions') }}
                </th>
              </tr>
            </thead>
            <tbody class="bg-white divide-y divide-theme-border">
              <tr v-for="workspace in workspaces" :key="workspace.id" class="hover:bg-theme-hover">
                <td class="px-6 py-4 whitespace-nowrap">
                  <div class="flex items-center gap-3">
                    <span
                      class="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-primary-100 text-theme-text"
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
                          d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4"
                        />
                      </svg>
                    </span>
                    <span class="text-sm font-medium text-theme-text">{{ workspace.name }}</span>
                    <span v-if="isCurrent(workspace)" class="text-xs text-theme-textLight">
                      {{ $t('workspaces_page.current') }}
                    </span>
                  </div>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-theme-textLight">
                  {{ formatDate(workspace.createdAt) }}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-theme-text">
                  {{ workspace.membersCount }}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-right text-sm">
                  <span class="inline-flex items-center gap-3">
                    <button
                      type="button"
                      @click="switchToWorkspace(workspace)"
                      class="text-emerald-600 hover:text-emerald-700 hover:underline disabled:opacity-50 text-sm"
                      :disabled="switchLoadingId === workspace.id"
                      title="Switch into this workspace"
                    >
                      {{
                        switchLoadingId === workspace.id
                          ? $t('workspaces_page.switching')
                          : $t('workspaces_page.switch_into')
                      }}
                    </button>
                    <button
                      v-if="canManage(workspace)"
                      type="button"
                      @click="openEditModal(workspace)"
                      class="text-primary-600 hover:text-primary-700 hover:underline disabled:opacity-50"
                      :disabled="updateLoadingId === workspace.id"
                      :title="$t('workspaces_page.edit_title')"
                    >
                      {{
                        updateLoadingId === workspace.id
                          ? $t('workspaces_page.saving')
                          : $t('workspaces_page.edit')
                      }}
                    </button>
                    <button
                      v-if="canManage(workspace)"
                      type="button"
                      @click="deleteModalWorkspace = workspace"
                      class="text-red-600 hover:text-red-700 hover:underline disabled:opacity-50"
                      :disabled="deletingId === workspace.id"
                      :title="$t('workspaces_page.delete_title')"
                    >
                      {{
                        deletingId === workspace.id
                          ? $t('workspaces_page.deleting')
                          : $t('workspaces_page.delete')
                      }}
                    </button>
                  </span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Pagination footer -->
        <div
          v-if="total > 0"
          class="bg-white px-6 py-4 border-t border-theme-border flex items-center justify-between"
        >
          <div class="text-sm text-theme-textLight">
            {{
              $t('workspaces_page.showing', {
                from: offset + 1,
                to: Math.min(offset + limit, total),
                total,
              })
            }}
          </div>
          <div class="flex items-center gap-2">
            <button
              type="button"
              @click="goToPage(offset - limit)"
              :disabled="offset === 0"
              class="btn-secondary text-sm disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {{ $t('workspaces_page.previous') }}
            </button>
            <button
              type="button"
              @click="goToPage(offset + limit)"
              :disabled="offset + limit >= total"
              class="btn-secondary text-sm disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {{ $t('workspaces_page.next') }}
            </button>
          </div>
        </div>
      </div>
    </main>

    <!-- Create workspace modal -->
    <BaseModal
      :open="showCreateModal"
      :title="$t('workspaces_page.create_title')"
      :description="$t('workspaces_page.create_desc')"
      :error="createModalError"
      size="sm"
      @close="closeCreateModal"
    >
      <form @submit.prevent="handleCreate" class="mt-4 space-y-4">
        <div>
          <label for="create-workspace-name" class="form-label">{{
            $t('workspaces_page.name')
          }}</label>
          <input
            id="create-workspace-name"
            v-model="newWorkspaceName"
            type="text"
            required
            class="input-field"
            placeholder="My Workspace"
            autofocus
          />
        </div>
        <div class="flex gap-3 justify-end">
          <button
            type="button"
            @click="closeCreateModal"
            class="btn-secondary"
            :disabled="createLoading"
          >
            {{ $t('common.cancel') }}
          </button>
          <button type="submit" :disabled="createLoading" class="btn-primary">
            {{ createLoading ? $t('workspaces_page.creating') : $t('common.create') }}
          </button>
        </div>
      </form>
    </BaseModal>

    <!-- Edit workspace modal -->
    <BaseModal
      :open="!!editModalWorkspace"
      :title="$t('workspaces_page.edit_title')"
      :description="$t('workspaces_page.edit_desc')"
      :error="editModalError"
      size="sm"
      @close="closeEditModal"
    >
      <form @submit.prevent="handleUpdate" class="mt-4 space-y-4">
        <div>
          <label for="edit-workspace-name" class="form-label">{{
            $t('workspaces_page.name')
          }}</label>
          <input
            id="edit-workspace-name"
            v-model="editWorkspaceName"
            type="text"
            required
            class="input-field"
            placeholder="My Workspace"
            autofocus
          />
        </div>
        <div class="flex gap-3 justify-end">
          <button
            type="button"
            @click="closeEditModal"
            class="btn-secondary"
            :disabled="updateLoadingId"
          >
            {{ $t('common.cancel') }}
          </button>
          <button
            type="submit"
            :disabled="updateLoadingId || editWorkspaceName.trim() === editModalWorkspace?.name"
            class="btn-primary"
          >
            {{ updateLoadingId ? $t('workspaces_page.saving') : $t('common.save') }}
          </button>
        </div>
      </form>
    </BaseModal>

    <ConfirmModal
      :open="!!deleteModalWorkspace"
      :title="$t('workspaces_page.delete_title')"
      :message="$t('workspaces_page.delete_confirm', { name: deleteModalWorkspace?.name })"
      :error="deleteModalError"
      :busy="deleting"
      :disabled="!canConfirmDelete"
      :confirm-label="$t('workspaces_page.delete')"
      :busy-label="$t('workspaces_page.deleting')"
      @cancel="closeDeleteModal"
      @confirm="confirmDelete"
    >
      <div class="mt-4">
        <label for="delete-confirm-name" class="block text-sm font-medium text-theme-text mb-1.5">
          {{ $t('workspaces_page.type_to_confirm', { name: deleteModalWorkspace?.name }) }}
        </label>
        <input
          id="delete-confirm-name"
          v-model="deleteConfirmName"
          type="text"
          class="input-field"
          :placeholder="deleteModalWorkspace?.name"
          autocomplete="off"
          :disabled="deleting"
        />
      </div>
    </ConfirmModal>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import AppNav from '@/comp/AppNav.vue'
import ConfirmModal from '@/comp/ConfirmModal.vue'
import { showFlash } from '@/lib/flash'
import { useUrlPagination } from '@/lib/pagination'
import { workspace_api } from '@/api'
import { saveWorkspaceToStorage, removeWorkspaceFromStorage } from '@/utils/storage'
import { canManageWorkspace, useWorkspaceContext } from '@/lib/permission'
import { errorText } from '@/lib/error'
import { formatDate } from '@/utils/date'
import BaseModal from '@/comp/BaseModal.vue'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const { currentUser, currentWorkspace } = useWorkspaceContext()
const { offset, limit, goToOffset, goToPage: setPage } = useUrlPagination({ limit: 50 })
const workspaces = ref([])
const loading = ref(false)
const total = ref(0)
const errorMessage = ref(null)

const editModalWorkspace = ref(null)
const editWorkspaceName = ref('')
const showCreateModal = ref(false)
const newWorkspaceName = ref('')
const createLoading = ref(false)
const updateLoadingId = ref(null)
const deleteModalWorkspace = ref(null)
const deleteConfirmName = ref('')
const deleting = ref(false)
const deletingId = ref(null)
const switchLoadingId = ref(null)

const editModalError = ref(null)
const createModalError = ref(null)
const deleteModalError = ref(null)

async function switchToWorkspace(workspace) {
  switchLoadingId.value = workspace.id
  errorMessage.value = null

  try {
    const res = await workspace_api.get(workspace.id)
    saveWorkspaceToStorage({ ...workspace, ...res.data, role: res.data.role })
    router.push('/dashboard')
  } catch (err) {
    errorMessage.value = errorText(err, t('workspaces_page.failed_switch'))
  } finally {
    switchLoadingId.value = null
  }
}

function openEditModal(workspace) {
  editModalWorkspace.value = workspace
  editWorkspaceName.value = workspace.name
  editModalError.value = null
}

function closeEditModal() {
  if (updateLoadingId.value) return
  editModalWorkspace.value = null
  editWorkspaceName.value = ''
  editModalError.value = null
}

function openCreateModal() {
  createModalError.value = null
  showCreateModal.value = true
}

function closeCreateModal() {
  if (createLoading.value) return
  showCreateModal.value = false
  createModalError.value = null
  newWorkspaceName.value = ''
}

async function handleCreate() {
  const name = newWorkspaceName.value.trim()
  if (!name) return

  createLoading.value = true
  errorMessage.value = null
  createModalError.value = null

  try {
    await workspace_api.create({ name })
    showCreateModal.value = false
    newWorkspaceName.value = ''
    setPage(1)
    await loadPage()
    showFlash(t('common.created'))
  } catch (err) {
    createModalError.value = errorText(err, t('workspaces_page.failed_create'))
  } finally {
    createLoading.value = false
  }
}

async function handleUpdate() {
  const workspace = editModalWorkspace.value
  const name = editWorkspaceName.value.trim()

  if (!workspace || !name) return
  updateLoadingId.value = workspace.id
  errorMessage.value = null
  editModalError.value = null

  try {
    const res = await workspace_api.update(workspace.id, { name })
    if (currentWorkspace.id === workspace.id) {
      saveWorkspaceToStorage({ ...currentWorkspace, ...res.data })
    }
    editModalWorkspace.value = null
    editWorkspaceName.value = ''
    editModalError.value = null
    await loadPage()
    showFlash(t('common.saved'))
  } catch (err) {
    editModalError.value = errorText(err, t('workspaces_page.failed_update'))
  } finally {
    updateLoadingId.value = null
  }
}

const canConfirmDelete = computed(() => {
  const name = deleteModalWorkspace.value?.name
  return name && deleteConfirmName.value.trim() === name
})

function isCurrent(workspace) {
  return currentWorkspace.id === workspace.id
}

function canManage(workspace) {
  return canManageWorkspace(currentUser, workspace)
}

async function loadPage() {
  loading.value = true
  errorMessage.value = null

  try {
    const res = await workspace_api.list({
      limit,
      offset: offset.value,
    })
    workspaces.value = res.data.workspaces
    total.value = res.data._meta.total
  } catch (err) {
    errorMessage.value = errorText(err, t('workspaces_page.failed_load'))
    workspaces.value = []
  } finally {
    loading.value = false
  }
}

function goToPage(newOffset) {
  if (newOffset < 0) return
  if (newOffset >= total.value) return
  goToOffset(newOffset)
}

watch(() => route.query.page, loadPage)

function closeDeleteModal() {
  deleteModalWorkspace.value = null
  deleteConfirmName.value = ''
  deleteModalError.value = null
}

async function confirmDelete() {
  const workspace = deleteModalWorkspace.value
  if (!workspace) return
  deleting.value = true
  deletingId.value = workspace.id
  errorMessage.value = null
  deleteModalError.value = null

  try {
    await workspace_api.delete(workspace.id)
    if (currentWorkspace.id === workspace.id) {
      removeWorkspaceFromStorage()
    }
    closeDeleteModal()
    await loadPage()
    showFlash(t('common.deleted'))
  } catch (err) {
    deleteModalError.value = errorText(err, t('workspaces_page.failed_delete'))
  } finally {
    deleting.value = false
    deletingId.value = null
  }
}

onMounted(loadPage)
</script>
