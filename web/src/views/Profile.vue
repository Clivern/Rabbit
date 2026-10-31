<template>
  <div class="min-h-screen bg-theme-bg">
    <AppNav />

    <div class="w-full px-4 sm:px-6 lg:px-8 py-8">
      <main>
        <div class="mb-8">
          <h1 class="text-2xl font-semibold text-theme-text">{{ $t('profile.title') }}</h1>
          <p class="text-sm text-theme-textLight mt-1">{{ $t('profile.subtitle') }}</p>
        </div>

        <div
          v-if="errorMessage"
          class="mb-6 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800"
        >
          {{ errorMessage }}
        </div>

        <section class="rounded-xl bg-white border border-theme-border shadow-sm overflow-hidden">
          <div class="px-6 py-4 border-b border-theme-border bg-primary-50/50">
            <h2 class="text-base font-semibold text-theme-text">{{ $t('profile.account') }}</h2>
            <p class="text-xs text-theme-textLight mt-0.5">{{ $t('profile.account_subtitle') }}</p>
          </div>
          <div class="p-6 space-y-6">
            <div class="flex flex-col sm:flex-row items-start sm:items-center gap-6">
              <div class="shrink-0">
                <img
                  :src="user?.avatar ?? ''"
                  :alt="user?.email ?? 'Avatar'"
                  class="h-20 w-20 rounded-full border-2 border-theme-border object-cover"
                />
              </div>
              <div class="min-w-0 flex-1 space-y-4">
                <div>
                  <label
                    for="profile-name"
                    class="block text-sm font-medium text-theme-text mb-1.5"
                    >{{ $t('profile.display_name') }}</label
                  >
                  <input
                    id="profile-name"
                    v-model="form.name"
                    type="text"
                    class="input-field max-w-md"
                    :placeholder="$t('profile.display_name_placeholder')"
                    :disabled="loading"
                  />
                </div>
                <div>
                  <p class="block text-sm font-medium text-theme-text mb-1.5">
                    {{ $t('profile.email') }}
                  </p>
                  <p class="text-sm text-theme-text">{{ user?.email }}</p>
                </div>
              </div>
            </div>

            <div class="pt-4 border-t border-theme-border space-y-4">
              <h3 class="text-sm font-semibold text-theme-text">{{ $t('profile.preferences') }}</h3>
              <p class="text-xs text-theme-textLight">{{ $t('profile.preferences_subtitle') }}</p>
              <div class="grid grid-cols-1 sm:grid-cols-2 gap-5 max-w-2xl">
                <div>
                  <label
                    for="profile-theme"
                    class="block text-xs font-medium text-theme-textLight uppercase tracking-wider mb-2"
                    >{{ $t('nav.theme') }}</label
                  >
                  <select
                    id="profile-theme"
                    v-model="form.theme"
                    class="input-field max-w-xs"
                    :disabled="loading"
                  >
                    <option value="default">{{ $t('nav.theme_default') }}</option>
                    <option value="blue">{{ $t('nav.theme_blue') }}</option>
                    <option value="slate">{{ $t('nav.theme_slate') }}</option>
                    <option value="emerald">{{ $t('nav.theme_emerald') }}</option>
                    <option value="dark">{{ $t('nav.theme_dark') }}</option>
                  </select>
                </div>
                <div>
                  <label
                    for="profile-language"
                    class="block text-xs font-medium text-theme-textLight uppercase tracking-wider mb-2"
                    >{{ $t('nav.language') }}</label
                  >
                  <select
                    id="profile-language"
                    v-model="form.language"
                    class="input-field max-w-xs"
                    :disabled="loading"
                  >
                    <option value="en">{{ $t('nav.english') }}</option>
                    <option value="fr">{{ $t('nav.french') }}</option>
                  </select>
                </div>
              </div>
            </div>

            <div class="grid grid-cols-1 sm:grid-cols-2 gap-5 pt-4 border-t border-theme-border">
              <div>
                <p class="text-xs font-medium text-theme-textLight uppercase tracking-wide mb-1">
                  {{ $t('profile.role') }}
                </p>
                <p class="text-sm text-theme-text capitalize">{{ user?.role }}</p>
              </div>
              <div>
                <p class="text-xs font-medium text-theme-textLight uppercase tracking-wide mb-1">
                  {{ $t('profile.last_login') }}
                </p>
                <p class="text-sm text-theme-text">{{ formatDate(user.lastLoginAt) }}</p>
              </div>
              <div>
                <p class="text-xs font-medium text-theme-textLight uppercase tracking-wide mb-1">
                  {{ $t('profile.member_since') }}
                </p>
                <p class="text-sm text-theme-text">{{ formatDate(user.createdAt) }}</p>
              </div>
            </div>

            <div class="flex items-center justify-end pt-2">
              <button type="button" class="btn-primary" :disabled="loading" @click="handleSave">
                <span v-if="!loading">{{ $t('profile.save_changes') }}</span>
                <span v-else class="inline-flex items-center gap-2">
                  <svg
                    class="animate-spin h-4 w-4"
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
                  {{ $t('profile.saving') }}
                </span>
              </button>
            </div>
          </div>
        </section>

        <section
          class="rounded-xl bg-white border border-theme-border shadow-sm overflow-hidden mt-8"
        >
          <div class="px-6 py-4 border-b border-theme-border bg-primary-50/50">
            <h2 class="text-base font-semibold text-theme-text">{{ $t('profile.api_keys') }}</h2>
            <p class="text-xs text-theme-textLight mt-0.5">{{ $t('profile.api_keys_subtitle') }}</p>
          </div>
          <div class="p-6 space-y-6">
            <div class="space-y-4">
              <div
                class="grid grid-cols-1 sm:grid-cols-[minmax(0,1fr)_minmax(0,1fr)_auto] items-end gap-3"
              >
                <div class="min-w-0">
                  <label
                    for="new-key-name"
                    class="block text-sm font-medium text-theme-text mb-1.5"
                    >{{ $t('profile.new_key_name') }}</label
                  >
                  <input
                    id="new-key-name"
                    v-model="newKeyName"
                    type="text"
                    class="input-field"
                    :placeholder="$t('profile.new_key_name_placeholder')"
                    :disabled="apiKeysLoading || creatingKey"
                    @keydown.enter.prevent="handleCreateKey"
                  />
                </div>
                <div class="min-w-0">
                  <label
                    for="new-key-expires"
                    class="block text-sm font-medium text-theme-text mb-1.5"
                    >{{ $t('profile.expires_at_optional') }}</label
                  >
                  <input
                    id="new-key-expires"
                    v-model="newKeyExpiresAt"
                    type="datetime-local"
                    class="input-field"
                    :disabled="apiKeysLoading || creatingKey"
                    :min="minExpiresAt"
                    @keydown.enter.prevent="handleCreateKey"
                  />
                </div>
                <button
                  type="button"
                  class="btn-primary"
                  :disabled="apiKeysLoading || creatingKey"
                  @click="handleCreateKey"
                >
                  <span v-if="!creatingKey">{{ $t('profile.create_key') }}</span>
                  <span v-else class="inline-flex items-center gap-2">
                    <svg
                      class="animate-spin h-4 w-4"
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
                    {{ $t('profile.creating') }}
                  </span>
                </button>
              </div>
            </div>

            <div
              v-if="newlyCreatedKey"
              class="rounded-lg border border-amber-200 bg-amber-50 p-4 space-y-2"
            >
              <p class="text-sm font-medium text-amber-900">{{ $t('profile.copy_key_now') }}</p>
              <div class="flex flex-wrap items-center gap-2">
                <code
                  class="flex-1 min-w-0 px-3 py-2 bg-white border border-amber-200 rounded text-sm text-theme-text break-all font-mono"
                  >{{ newlyCreatedKey }}</code
                >
                <button type="button" class="btn-secondary shrink-0" @click="copyKeyToClipboard">
                  {{ keyCopied ? $t('profile.copied') : $t('profile.copy') }}
                </button>
              </div>
              <button
                type="button"
                class="text-sm text-amber-800 hover:text-amber-900 font-medium"
                @click="dismissNewKey"
              >
                {{ $t('profile.dismiss') }}
              </button>
            </div>

            <div
              v-if="apiKeysLoading && apiKeys.length === 0"
              class="text-sm text-theme-textLight py-4"
            >
              {{ $t('profile.loading_api_keys') }}
            </div>
            <div v-else-if="apiKeys.length === 0" class="text-sm text-theme-textLight py-4">
              {{ $t('profile.no_api_keys_yet') }}
            </div>
            <div v-else class="border border-theme-border rounded-lg overflow-x-auto">
              <table class="min-w-full divide-y divide-theme-border">
                <thead class="bg-theme-bg">
                  <tr>
                    <th
                      scope="col"
                      class="px-4 py-2.5 text-left text-xs font-medium text-theme-textLight uppercase tracking-wide"
                    >
                      {{ $t('profile.name') }}
                    </th>
                    <th
                      scope="col"
                      class="px-4 py-2.5 text-left text-xs font-medium text-theme-textLight uppercase tracking-wide"
                    >
                      {{ $t('profile.created') }}
                    </th>
                    <th
                      scope="col"
                      class="px-4 py-2.5 text-left text-xs font-medium text-theme-textLight uppercase tracking-wide"
                    >
                      {{ $t('profile.expires') }}
                    </th>
                    <th
                      scope="col"
                      class="px-4 py-2.5 text-right text-xs font-medium text-theme-textLight uppercase tracking-wide"
                    >
                      {{ $t('profile.actions') }}
                    </th>
                  </tr>
                </thead>
                <tbody class="bg-white divide-y divide-theme-border">
                  <tr v-for="key in apiKeys" :key="key.id" class="hover:bg-theme-hover/50">
                    <td class="px-4 py-3 text-sm text-theme-text">
                      {{ key.name }}
                    </td>
                    <td class="px-4 py-3 text-sm text-theme-textLight">
                      {{ formatDate(key.createdAt) }}
                    </td>
                    <td class="px-4 py-3 text-sm text-theme-textLight">
                      {{ key.expiresAt ? formatDateTime(key.expiresAt) : $t('profile.never') }}
                    </td>
                    <td class="px-4 py-3 text-right">
                      <button
                        type="button"
                        class="text-sm text-red-600 hover:text-red-700 font-medium disabled:opacity-50"
                        :disabled="deletingId === key.id"
                        @click="revokeModalKey = key"
                      >
                        {{ deletingId === key.id ? $t('profile.revoking') : $t('profile.revoke') }}
                      </button>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </section>
      </main>
    </div>

    <ConfirmModal
      :open="!!revokeModalKey"
      :title="$t('profile.revoke_title')"
      :message="$t('profile.revoke_confirm', { name: revokeModalKey?.name })"
      :busy="!!deletingId"
      :confirm-label="$t('profile.revoke')"
      :busy-label="$t('profile.revoking')"
      @cancel="revokeModalKey = null"
      @confirm="confirmRevokeKey"
    />
  </div>
</template>

<script setup>
import { ref, reactive, computed, watch, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import { api_keys_api, auth_api } from '@/api'
import AppNav from '@/comp/AppNav.vue'
import ConfirmModal from '@/comp/ConfirmModal.vue'
import { showFlash } from '@/lib/flash'
import { user, saveUser } from '@/lib/auth'
import {
  applyTheme,
  applyUserPreferences,
  USER_LANGUAGE_EN,
  USER_THEME_DEFAULT,
} from '@/lib/preferences'
import { errorText } from '@/lib/error'
import { formatDate, formatDateTime } from '@/utils/date'

const { t } = useI18n()
const loading = ref(false)
const errorMessage = ref(null)

const form = reactive({
  name: user.value?.name,
  language: user.value?.language ?? USER_LANGUAGE_EN,
  theme: user.value?.theme ?? USER_THEME_DEFAULT,
})

// The theme previews at once without saving; leaving without saving restores the saved one.
watch(
  () => form.theme,
  (theme) => {
    if (theme) applyTheme(theme)
  }
)

onBeforeUnmount(() => {
  if (user.value?.theme) applyTheme(user.value.theme)
})

const apiKeys = ref([])
const apiKeysLoading = ref(false)
const creatingKey = ref(false)
const newKeyName = ref('')
const newKeyExpiresAt = ref('')
const newlyCreatedKey = ref(null)
const keyCopied = ref(false)
const deletingId = ref(null)
const revokeModalKey = ref(null)

const minExpiresAt = computed(() => {
  const d = new Date()
  d.setMinutes(d.getMinutes() - d.getTimezoneOffset())
  return d.toISOString().slice(0, 16)
})

watch(
  user,
  (userData) => {
    if (userData) {
      form.name = userData.name
      form.language = userData.language ?? USER_LANGUAGE_EN
      form.theme = userData.theme ?? USER_THEME_DEFAULT
    }
  },
  { immediate: true, deep: true }
)

async function loadApiKeys() {
  apiKeysLoading.value = true
  errorMessage.value = null

  try {
    const res = await api_keys_api.list({ limit: 100, offset: 0 })
    apiKeys.value = res.data.keys ?? []
  } catch (err) {
    errorMessage.value = errorText(err, t('profile.failed_load_api_keys'))
  } finally {
    apiKeysLoading.value = false
  }
}

async function handleCreateKey() {
  const name = newKeyName.value.trim()
  errorMessage.value = null
  newlyCreatedKey.value = null

  if (!name) {
    errorMessage.value = t('profile.key_name_required')
    return
  }
  creatingKey.value = true

  try {
    const res = await api_keys_api.create({
      name: name,
      expiresAt: newKeyExpiresAt.value ? new Date(newKeyExpiresAt.value).toISOString() : '',
    })

    newlyCreatedKey.value = res.data.key
    newKeyName.value = ''
    newKeyExpiresAt.value = ''
    await loadApiKeys()
  } catch (err) {
    errorMessage.value = errorText(err, t('profile.failed_create_key'))
  } finally {
    creatingKey.value = false
  }
}

function copyKeyToClipboard() {
  if (!newlyCreatedKey.value) return
  navigator.clipboard.writeText(newlyCreatedKey.value).then(() => {
    keyCopied.value = true
    setTimeout(() => {
      keyCopied.value = false
    }, 2000)
  })
}

function dismissNewKey() {
  newlyCreatedKey.value = null
  keyCopied.value = false
}

async function confirmRevokeKey() {
  const key = revokeModalKey.value

  if (!key) return
  deletingId.value = key.id
  errorMessage.value = null

  try {
    await api_keys_api.delete(key.id)
    await loadApiKeys()
    newlyCreatedKey.value = null
    revokeModalKey.value = null
  } catch (err) {
    errorMessage.value = errorText(err, t('profile.failed_revoke_key'))
  } finally {
    deletingId.value = null
  }
}

async function handleSave() {
  loading.value = true
  errorMessage.value = null

  const payload = {
    name: form.name.trim(),
    language: form.language,
    theme: form.theme,
  }

  try {
    const res = await auth_api.updateProfile(payload)
    saveUser(res.data.user)
    applyUserPreferences(res.data.user)
    showFlash(t('common.saved'))
  } catch (err) {
    errorMessage.value = errorText(err, t('profile.update_failed'))
  } finally {
    loading.value = false
  }
}

onMounted(loadApiKeys)
</script>
