<template>
  <div class="min-h-screen bg-theme-bg">
    <AppNav />

    <main class="w-full py-8 px-4 sm:px-6 lg:px-8">
      <div class="mb-8">
        <h1 class="text-3xl font-semibold text-theme-text">
          {{ $t('workspace_settings_page.title') }}
        </h1>
        <p class="text-theme-textLight mt-2">{{ $t('workspace_settings_page.subtitle') }}</p>
      </div>

      <div
        v-if="errorMessage"
        class="mb-6 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800"
      >
        {{ errorMessage }}
      </div>

      <div class="space-y-8">
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-8">
          <section class="rounded-xl border border-theme-border bg-white shadow-sm overflow-hidden">
            <div class="border-b border-theme-border px-6 py-4">
              <h2 class="text-lg font-semibold text-theme-text">
                {{ $t('workspace_settings_page.general') }}
              </h2>
              <p class="mt-1 text-sm text-theme-textLight">
                {{ $t('workspace_settings_page.general_desc') }}
              </p>
            </div>

            <div v-if="workspaceLoading" class="p-8 text-center text-sm text-theme-textLight">
              {{ $t('workspace_settings_page.loading_workspace') }}
            </div>

            <form v-else class="p-6 space-y-5" @submit.prevent="handleSaveWorkspace">
              <div>
                <label for="workspace-name" class="form-label">{{
                  $t('workspace_settings_page.name')
                }}</label>
                <input
                  id="workspace-name"
                  v-model="workspaceName"
                  type="text"
                  required
                  minlength="3"
                  maxlength="60"
                  class="input-field"
                  :disabled="savingWorkspace"
                />
              </div>

              <div class="pt-2">
                <button
                  type="submit"
                  class="btn-primary disabled:opacity-50"
                  :disabled="savingWorkspace || !canSaveWorkspace"
                >
                  {{
                    savingWorkspace
                      ? $t('workspace_settings_page.saving')
                      : $t('workspace_settings_page.save')
                  }}
                </button>
              </div>
            </form>
          </section>

          <section class="rounded-xl border border-theme-border bg-white shadow-sm overflow-hidden">
            <div class="border-b border-theme-border px-6 py-4">
              <h2 class="text-lg font-semibold text-theme-text">
                {{ $t('workspace_settings_page.identifiers') }}
              </h2>
              <p class="mt-1 text-sm text-theme-textLight">
                {{ $t('workspace_settings_page.identifiers_desc') }}
              </p>
            </div>

            <div v-if="workspaceLoading" class="p-8 text-center text-sm text-theme-textLight">
              {{ $t('workspace_settings_page.loading_workspace') }}
            </div>

            <div v-else class="p-6 space-y-5">
              <div>
                <label class="form-label">{{ $t('workspace_settings_page.handle') }}</label>
                <div class="flex items-center gap-2">
                  <input
                    :value="workspace.handle"
                    type="text"
                    readonly
                    class="input-field bg-theme-bg text-theme-textLight"
                  />
                  <button
                    type="button"
                    class="btn-secondary shrink-0"
                    @click="copyText(workspace.handle, 'handle')"
                  >
                    {{
                      copiedField === 'handle'
                        ? $t('workspace_settings_page.copied')
                        : $t('workspace_settings_page.copy')
                    }}
                  </button>
                </div>
              </div>

              <div>
                <label class="form-label">{{ $t('workspace_settings_page.id') }}</label>
                <div class="flex items-center gap-2">
                  <input
                    :value="workspace.id"
                    type="text"
                    readonly
                    class="input-field bg-theme-bg text-theme-textLight font-mono text-sm"
                  />
                  <button
                    type="button"
                    class="btn-secondary shrink-0"
                    @click="copyText(workspace.id, 'id')"
                  >
                    {{
                      copiedField === 'id'
                        ? $t('workspace_settings_page.copied')
                        : $t('workspace_settings_page.copy')
                    }}
                  </button>
                </div>
              </div>
            </div>
          </section>
        </div>

        <section class="rounded-xl border border-theme-border bg-white shadow-sm overflow-hidden">
          <div class="border-b border-theme-border px-6 py-4">
            <h2 class="text-lg font-semibold text-theme-text">
              {{ $t('workspace_settings_page.access_keys') }}
            </h2>
            <p class="mt-1 text-sm text-theme-textLight">
              {{ $t('workspace_settings_page.access_keys_desc') }}
            </p>
          </div>

          <div class="p-6 space-y-6">
            <form
              class="rounded-lg border border-theme-border bg-theme-bg/40 p-5 space-y-5"
              @submit.prevent="handleCreateKey"
            >
              <div class="grid grid-cols-1 sm:grid-cols-[minmax(0,1fr)_14rem] gap-4 items-end">
                <div class="min-w-0">
                  <label for="new-access-key-name" class="form-label">{{
                    $t('workspace_settings_page.key_name')
                  }}</label>
                  <input
                    id="new-access-key-name"
                    v-model="newKeyName"
                    type="text"
                    required
                    class="input-field"
                    :placeholder="$t('workspace_settings_page.key_name_placeholder')"
                    :disabled="accessKeysLoading || creatingKey"
                  />
                </div>
                <div class="min-w-0">
                  <label for="new-access-key-expires" class="form-label">{{
                    $t('workspace_settings_page.expires_at_optional')
                  }}</label>
                  <input
                    id="new-access-key-expires"
                    v-model="newKeyExpiresAt"
                    type="datetime-local"
                    class="input-field w-full"
                    :disabled="accessKeysLoading || creatingKey"
                    :min="minExpiresAt"
                  />
                </div>
              </div>

              <div v-if="WORKSPACE_KEY_PERMISSIONS.length">
                <p class="form-label mb-2">{{ $t('workspace_settings_page.permissions') }}</p>
                <div class="grid grid-cols-1 md:grid-cols-2 items-start gap-3">
                  <details
                    v-for="group in WORKSPACE_KEY_PERMISSION_GROUPS"
                    :key="group.labelKey"
                    class="rounded-md border border-theme-border bg-white"
                  >
                    <summary
                      class="flex items-center justify-between cursor-pointer select-none px-2.5 py-2 text-xs font-medium text-theme-text"
                    >
                      <span>{{ $t(group.labelKey) }}</span>
                      <span class="text-theme-textLight"
                        >{{ countSelected(group) }}/{{ group.permissions.length }}</span
                      >
                    </summary>
                    <div class="grid grid-cols-1 sm:grid-cols-2 gap-1.5 px-2.5 pb-2.5">
                      <label
                        v-for="perm in group.permissions"
                        :key="perm.value"
                        class="flex items-center gap-1.5 rounded-md border border-theme-border bg-white px-2 py-1.5 text-xs text-theme-text cursor-pointer hover:bg-theme-hover/40 transition-colors has-[:checked]:border-primary-700 has-[:checked]:bg-primary-50/40"
                      >
                        <input
                          v-model="selectedPermissions"
                          type="checkbox"
                          :value="perm.value"
                          class="checkbox shrink-0 !size-3.5"
                          :disabled="accessKeysLoading || creatingKey"
                        />
                        <span>{{ $t(perm.labelKey) }}</span>
                      </label>
                    </div>
                  </details>
                </div>
              </div>

              <div class="pt-1">
                <button
                  type="submit"
                  class="btn-primary"
                  :disabled="accessKeysLoading || creatingKey"
                >
                  <span v-if="!creatingKey">{{ $t('workspace_settings_page.create_key') }}</span>
                  <span v-else class="inline-flex items-center gap-2">
                    <svg class="spinner" fill="none" viewBox="0 0 24 24" aria-hidden="true">
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
                    {{ $t('workspace_settings_page.creating') }}
                  </span>
                </button>
              </div>
            </form>

            <div
              v-if="newlyCreatedKey"
              class="rounded-lg border border-amber-200 bg-amber-50 p-4 space-y-2"
            >
              <p class="text-sm font-medium text-amber-900">
                {{ $t('workspace_settings_page.copy_key_now') }}
              </p>
              <div class="flex flex-wrap items-center gap-2">
                <code
                  class="flex-1 min-w-0 px-3 py-2 bg-white border border-amber-200 rounded text-sm text-theme-text break-all font-mono"
                  >{{ newlyCreatedKey }}</code
                >
                <button
                  type="button"
                  class="btn-secondary shrink-0"
                  @click="copyText(newlyCreatedKey, 'key')"
                >
                  {{
                    copiedField === 'key'
                      ? $t('workspace_settings_page.copied')
                      : $t('workspace_settings_page.copy')
                  }}
                </button>
              </div>
              <button
                type="button"
                class="text-sm text-amber-800 hover:text-amber-900 font-medium"
                @click="dismissNewKey"
              >
                {{ $t('workspace_settings_page.dismiss') }}
              </button>
            </div>

            <div
              v-if="accessKeysLoading && accessKeys.length === 0"
              class="rounded-lg border border-dashed border-theme-border bg-theme-bg/30 px-6 py-10 text-center text-sm text-theme-textLight"
            >
              {{ $t('workspace_settings_page.loading_keys') }}
            </div>
            <div
              v-else-if="accessKeys.length === 0"
              class="rounded-lg border border-dashed border-theme-border bg-theme-bg/30 px-6 py-10 text-center text-sm text-theme-textLight"
            >
              {{ $t('workspace_settings_page.no_keys_yet') }}
            </div>
            <div v-else class="border border-theme-border rounded-lg overflow-x-auto">
              <table class="min-w-full divide-y divide-theme-border">
                <thead class="bg-theme-bg">
                  <tr>
                    <th
                      class="px-4 py-2.5 text-left text-xs font-medium text-theme-textLight uppercase tracking-wide"
                    >
                      {{ $t('workspace_settings_page.key_name') }}
                    </th>
                    <th
                      class="px-4 py-2.5 text-left text-xs font-medium text-theme-textLight uppercase tracking-wide"
                    >
                      {{ $t('workspace_settings_page.permissions') }}
                    </th>
                    <th
                      class="px-4 py-2.5 text-left text-xs font-medium text-theme-textLight uppercase tracking-wide"
                    >
                      {{ $t('workspace_settings_page.created') }}
                    </th>
                    <th
                      class="px-4 py-2.5 text-left text-xs font-medium text-theme-textLight uppercase tracking-wide"
                    >
                      {{ $t('workspace_settings_page.expires') }}
                    </th>
                    <th
                      class="px-4 py-2.5 text-right text-xs font-medium text-theme-textLight uppercase tracking-wide"
                    >
                      {{ $t('workspace_settings_page.actions') }}
                    </th>
                  </tr>
                </thead>
                <tbody class="bg-white divide-y divide-theme-border">
                  <tr v-for="key in accessKeys" :key="key.id" class="hover:bg-theme-hover/50">
                    <td class="px-4 py-3 text-sm text-theme-text">{{ key.name }}</td>
                    <td class="px-4 py-3 text-sm text-theme-textLight">
                      <button
                        v-if="key.permissions?.length"
                        type="button"
                        class="inline-flex items-center gap-1.5 rounded-md px-2 py-1 -mx-2 text-primary-700 hover:text-primary-800 hover:bg-primary-50/60 font-medium transition-colors"
                        @click="openPermissionsModal(key)"
                      >
                        <span>{{
                          $t('workspace_settings_page.view_permissions', {
                            count: key.permissions.length,
                          })
                        }}</span>
                        <svg
                          class="w-4 h-4 shrink-0"
                          fill="none"
                          stroke="currentColor"
                          viewBox="0 0 24 24"
                          aria-hidden="true"
                        >
                          <path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            stroke-width="2"
                            d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
                          />
                          <path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            stroke-width="2"
                            d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"
                          />
                        </svg>
                      </button>
                      <span v-else>—</span>
                    </td>
                    <td class="px-4 py-3 text-sm text-theme-textLight">
                      {{ formatDate(key.createdAt) }}
                    </td>
                    <td class="px-4 py-3 text-sm text-theme-textLight">
                      {{
                        key.expiresAt
                          ? formatDateTime(key.expiresAt)
                          : $t('workspace_settings_page.never')
                      }}
                    </td>
                    <td class="px-4 py-3 text-right">
                      <button
                        type="button"
                        class="text-sm text-red-600 hover:text-red-700 font-medium disabled:opacity-50"
                        :disabled="deletingId === key.id"
                        @click="revokeModalKey = key"
                      >
                        {{
                          deletingId === key.id
                            ? $t('workspace_settings_page.revoking')
                            : $t('workspace_settings_page.revoke')
                        }}
                      </button>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </section>
      </div>
    </main>

    <ConfirmModal
      :open="!!revokeModalKey"
      :title="$t('workspace_settings_page.revoke_title')"
      :message="$t('workspace_settings_page.revoke_confirm', { name: revokeModalKey?.name })"
      :busy="!!deletingId"
      :confirm-label="$t('workspace_settings_page.revoke')"
      :busy-label="$t('workspace_settings_page.revoking')"
      @cancel="revokeModalKey = null"
      @confirm="confirmRevokeKey"
    />

    <Teleport to="body">
      <Transition
        enter-active-class="transition ease-out duration-200"
        enter-from-class="opacity-0"
        enter-to-class="opacity-100"
        leave-active-class="transition ease-in duration-150"
        leave-from-class="opacity-100"
        leave-to-class="opacity-0"
      >
        <div
          v-if="permissionsModalKey"
          class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50"
          role="dialog"
          aria-modal="true"
          aria-labelledby="access-key-permissions-title"
          @click.self="closePermissionsModal"
        >
          <div
            class="w-full max-w-xl max-h-[85vh] flex flex-col rounded-xl bg-white border border-theme-border shadow-xl"
          >
            <div class="border-b border-theme-border px-4 py-3 shrink-0">
              <h2 id="access-key-permissions-title" class="text-base font-semibold text-theme-text">
                {{ $t('workspace_settings_page.permissions_title') }}
              </h2>
              <p class="mt-0.5 text-xs text-theme-textLight">
                {{
                  $t('workspace_settings_page.permissions_for', { name: permissionsModalKey.name })
                }}
              </p>
            </div>
            <div class="overflow-y-auto px-4 py-3">
              <ul class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-1.5">
                <li
                  v-for="label in getPermissionLabels(permissionsModalKey.permissions)"
                  :key="label"
                  class="rounded-md border border-theme-border bg-theme-bg/40 px-2 py-1.5 text-xs text-theme-text"
                >
                  {{ label }}
                </li>
              </ul>
            </div>
            <div class="border-t border-theme-border px-4 py-3 shrink-0 flex justify-end">
              <button type="button" class="btn-secondary" @click="closePermissionsModal">
                {{ $t('workspace_settings_page.dismiss') }}
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import AppNav from '@/comp/AppNav.vue'
import ConfirmModal from '@/comp/ConfirmModal.vue'
import { workspace_api, workspace_access_key_api } from '@/api'
import { showFlash } from '@/lib/flash'
import {
  WORKSPACE_KEY_PERMISSION_GROUPS,
  WORKSPACE_KEY_PERMISSIONS,
  useWorkspaceContext,
} from '@/lib/permission'
import { saveWorkspaceToStorage } from '@/utils/storage'
import { errorText } from '@/lib/error'
import { formatDate, formatDateTime } from '@/utils/date'

const { t } = useI18n()

const { currentWorkspace } = useWorkspaceContext()

const errorMessage = ref(null)
const workspace = ref(null)
const workspaceName = ref('')
const workspaceLoading = ref(true)
const savingWorkspace = ref(false)
const copiedField = ref(null)

const accessKeys = ref([])
const accessKeysLoading = ref(false)
const creatingKey = ref(false)
const newKeyName = ref('')
const newKeyExpiresAt = ref('')
const selectedPermissions = ref([])
const newlyCreatedKey = ref(null)
const deletingId = ref(null)
const revokeModalKey = ref(null)
const permissionsModalKey = ref(null)

const minExpiresAt = computed(() => {
  const d = new Date()
  d.setMinutes(d.getMinutes() - d.getTimezoneOffset())
  return d.toISOString().slice(0, 16)
})

const canSaveWorkspace = computed(() => {
  return workspaceName.value.trim() !== workspace.value?.name
})

function countSelected(group) {
  return group.permissions.filter((perm) => selectedPermissions.value.includes(perm.value)).length
}

function getPermissionLabels(permissions) {
  if (!permissions?.length) return []
  return permissions.map((value) => {
    const perm = WORKSPACE_KEY_PERMISSIONS.find((item) => item.value === value)
    return perm ? t(perm.labelKey) : value
  })
}

function openPermissionsModal(key) {
  permissionsModalKey.value = key
}

function closePermissionsModal() {
  permissionsModalKey.value = null
}

async function confirmRevokeKey() {
  const key = revokeModalKey.value
  if (!key) return

  deletingId.value = key.id
  errorMessage.value = null

  try {
    await workspace_access_key_api.delete(currentWorkspace.id, key.id)
    await loadAccessKeys()
    newlyCreatedKey.value = null
    revokeModalKey.value = null
  } catch (err) {
    errorMessage.value = errorText(err, t('workspace_settings_page.failed_revoke_key'))
  } finally {
    deletingId.value = null
  }
}

async function copyText(text, field) {
  try {
    await navigator.clipboard.writeText(text)
    copiedField.value = field
    setTimeout(() => {
      if (copiedField.value === field) copiedField.value = null
    }, 2000)
  } catch {
    errorMessage.value = t('workspace_settings_page.copy_failed')
  }
}

function dismissNewKey() {
  newlyCreatedKey.value = null
  copiedField.value = null
}

async function loadWorkspace() {
  workspaceLoading.value = true
  errorMessage.value = null

  try {
    const res = await workspace_api.get(currentWorkspace.id)
    workspace.value = res.data
    workspaceName.value = res.data.name ?? ''
    saveWorkspaceToStorage({
      ...currentWorkspace,
      ...res.data,
      role: res.data.role ?? currentWorkspace.role,
    })
  } catch (err) {
    errorMessage.value = errorText(err, t('workspace_settings_page.failed_load_workspace'))
  } finally {
    workspaceLoading.value = false
  }
}

async function handleSaveWorkspace() {
  if (!canSaveWorkspace.value) return

  savingWorkspace.value = true
  errorMessage.value = null

  try {
    const res = await workspace_api.update(currentWorkspace.id, {
      name: workspaceName.value.trim(),
    })
    workspace.value = res.data
    workspaceName.value = res.data.name
    saveWorkspaceToStorage({
      ...currentWorkspace,
      ...res.data,
      role: res.data.role ?? currentWorkspace.role,
    })
    showFlash(t('workspace_settings_page.workspace_saved'))
  } catch (err) {
    errorMessage.value = errorText(err, t('workspace_settings_page.failed_save_workspace'))
  } finally {
    savingWorkspace.value = false
  }
}

async function loadAccessKeys() {
  accessKeysLoading.value = true
  errorMessage.value = null

  try {
    const res = await workspace_access_key_api.list(currentWorkspace.id, { limit: 100, offset: 0 })
    accessKeys.value = res.data.keys ?? []
  } catch (err) {
    errorMessage.value = errorText(err, t('workspace_settings_page.failed_load_keys'))
  } finally {
    accessKeysLoading.value = false
  }
}

async function handleCreateKey() {
  const name = newKeyName.value.trim()
  if (!name) {
    errorMessage.value = t('workspace_settings_page.key_name_required')
    return
  }
  if (WORKSPACE_KEY_PERMISSIONS.length > 0 && selectedPermissions.value.length === 0) {
    errorMessage.value = t('workspace_settings_page.permissions_required')
    return
  }

  creatingKey.value = true
  errorMessage.value = null
  newlyCreatedKey.value = null

  try {
    const res = await workspace_access_key_api.create(currentWorkspace.id, {
      name,
      expiresAt: newKeyExpiresAt.value ? new Date(newKeyExpiresAt.value).toISOString() : '',
      permissions: selectedPermissions.value,
    })
    newlyCreatedKey.value = res.data.key
    newKeyName.value = ''
    newKeyExpiresAt.value = ''
    await loadAccessKeys()
  } catch (err) {
    errorMessage.value = errorText(err, t('workspace_settings_page.failed_create_key'))
  } finally {
    creatingKey.value = false
  }
}

onMounted(async () => {
  await Promise.all([loadWorkspace(), loadAccessKeys()])
})
</script>
