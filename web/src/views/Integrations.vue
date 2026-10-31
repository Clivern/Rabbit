<template>
  <div class="min-h-screen bg-theme-bg">
    <AppNav />

    <main class="w-full py-8 px-4 sm:px-6 lg:px-8">
      <div class="mb-8">
        <h1 class="text-3xl font-semibold text-theme-text">{{ $t('integrations_page.title') }}</h1>
        <p class="text-theme-textLight mt-2">{{ $t('integrations_page.subtitle') }}</p>
      </div>

      <div
        v-if="newSecret"
        class="mb-6 rounded-lg border border-amber-200 bg-amber-50 p-4 space-y-2"
      >
        <p class="text-sm font-medium text-amber-900">{{ $t('integrations_page.secret_desc') }}</p>
        <div class="flex flex-wrap items-center gap-2">
          <code
            class="flex-1 min-w-0 px-3 py-2 bg-white border border-amber-200 rounded text-sm text-theme-text break-all font-mono"
            >{{ newSecret }}</code
          >
          <button type="button" class="btn-secondary shrink-0" @click="copySecret">
            {{ copied ? $t('integrations_page.copied') : $t('integrations_page.copy') }}
          </button>
        </div>
        <button
          type="button"
          class="text-sm text-amber-800 hover:text-amber-900 font-medium"
          @click="dismissNewSecret"
        >
          {{ $t('integrations_page.dismiss') }}
        </button>
      </div>

      <section class="rounded-xl border border-theme-border bg-white shadow-sm overflow-hidden">
        <div
          class="flex flex-wrap items-start justify-between gap-4 border-b border-theme-border px-6 py-4"
        >
          <div>
            <h2 class="text-lg font-semibold text-theme-text">
              {{ $t('integrations_page.list_title') }}
            </h2>
            <p class="mt-1 text-sm text-theme-textLight">{{ $t('integrations_page.list_desc') }}</p>
          </div>
          <button type="button" class="btn-primary" @click="openForm()">
            {{ $t('integrations_page.add') }}
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
                  {{ $t('integrations_page.name') }}
                </th>
                <th
                  class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider"
                >
                  {{ $t('integrations_page.type') }}
                </th>
                <th
                  class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider"
                >
                  {{ $t('integrations_page.destination') }}
                </th>
                <th
                  class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider"
                >
                  {{ $t('integrations_page.events') }}
                </th>
                <th
                  class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider"
                >
                  {{ $t('integrations_page.status') }}
                </th>
                <th
                  class="px-6 py-3 text-right text-xs font-medium text-theme-textLight uppercase tracking-wider"
                >
                  {{ $t('integrations_page.actions') }}
                </th>
              </tr>
            </thead>
            <tbody class="bg-white divide-y divide-theme-border">
              <tr v-if="loading">
                <td colspan="6" class="px-6 py-4 text-center text-sm text-theme-textLight">
                  {{ $t('common.loading') }}
                </td>
              </tr>
              <tr v-else-if="!integrations.length">
                <td colspan="6" class="px-6 py-4 text-center text-sm text-theme-textLight">
                  {{ $t('integrations_page.empty') }}
                </td>
              </tr>
              <template v-else>
                <tr v-for="item in integrations" :key="item.id" class="hover:bg-theme-hover">
                  <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-theme-text">
                    {{ item.name }}
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap text-sm text-theme-text">
                    {{ $t(`integrations_page.type_${item.type}`) }}
                  </td>
                  <td class="px-6 py-4 text-sm text-theme-text font-mono break-all">
                    {{ item.url || item.site }}
                  </td>
                  <td class="px-6 py-4 text-sm">
                    <div class="flex flex-wrap gap-1">
                      <span
                        v-for="event in item.events"
                        :key="event"
                        class="inline-flex rounded-full bg-theme-hover px-2 py-0.5 text-xs font-medium text-theme-text"
                      >
                        {{ $t(`integrations_page.event_${event.replaceAll('.', '_')}`) }}
                      </span>
                    </div>
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap text-sm">
                    <div class="flex items-center gap-2">
                      <button
                        type="button"
                        role="switch"
                        :aria-checked="item.isActive"
                        :aria-label="$t('integrations_page.toggle', { name: item.name })"
                        class="relative inline-flex h-5 w-9 shrink-0 items-center rounded-full transition-colors disabled:opacity-50"
                        :class="item.isActive ? 'bg-emerald-600' : 'bg-gray-300'"
                        :disabled="togglingId === item.id"
                        @click="toggle(item)"
                      >
                        <span
                          class="inline-block h-4 w-4 rounded-full bg-white shadow transition-transform"
                          :class="item.isActive ? 'translate-x-4' : 'translate-x-0.5'"
                        />
                      </button>
                      <span :class="item.isActive ? 'text-emerald-800' : 'text-theme-textLight'">
                        {{
                          item.isActive
                            ? $t('integrations_page.enabled')
                            : $t('integrations_page.disabled')
                        }}
                      </span>
                    </div>
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap text-right text-sm">
                    <button
                      type="button"
                      class="font-medium text-primary-700 hover:underline"
                      @click="openForm(item)"
                    >
                      {{ $t('integrations_page.edit') }}
                    </button>
                    <button
                      type="button"
                      class="ml-4 font-medium text-red-600 hover:underline"
                      @click="deleteTarget = item"
                    >
                      {{ $t('integrations_page.delete') }}
                    </button>
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
              $t('integrations_page.showing', {
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
              {{ $t('integrations_page.previous') }}
            </button>
            <button
              type="button"
              class="btn-secondary text-sm disabled:opacity-50 disabled:cursor-not-allowed"
              :disabled="offset + limit >= total"
              @click="goToPage(offset + limit)"
            >
              {{ $t('integrations_page.next') }}
            </button>
          </div>
        </div>
      </section>
    </main>

    <BaseModal
      :open="showForm"
      :title="isEditing ? $t('integrations_page.edit_title') : $t('integrations_page.add_title')"
      :description="$t(`integrations_page.type_${form.type}_desc`)"
      :error="formError"
      @close="closeForm"
    >
      <form class="mt-4 space-y-4" @submit.prevent="submit">
        <div class="grid gap-4 sm:grid-cols-2">
          <div>
            <label for="integration-type" class="form-label">{{
              $t('integrations_page.type')
            }}</label>
            <select
              id="integration-type"
              v-model="form.type"
              class="input-field"
              :disabled="saving || isEditing"
            >
              <option v-for="type in TYPES" :key="type" :value="type">
                {{ $t(`integrations_page.type_${type}`) }}
              </option>
            </select>
          </div>
          <div>
            <label for="integration-name" class="form-label">{{
              $t('integrations_page.name')
            }}</label>
            <input
              id="integration-name"
              v-model="form.name"
              type="text"
              required
              maxlength="60"
              class="input-field"
              :placeholder="$t('integrations_page.name_placeholder')"
              autocomplete="off"
              :disabled="saving"
            />
          </div>
        </div>

        <div v-if="URL_PLACEHOLDERS[form.type]">
          <label for="integration-url" class="form-label">{{
            $t(`integrations_page.url_${form.type}`)
          }}</label>
          <input
            id="integration-url"
            v-model="form.url"
            type="url"
            required
            maxlength="2048"
            class="input-field font-mono"
            :placeholder="URL_PLACEHOLDERS[form.type]"
            autocomplete="off"
            :disabled="saving"
          />
        </div>

        <div v-if="form.type === 'datadog'">
          <label for="integration-site" class="form-label">{{
            $t('integrations_page.site')
          }}</label>
          <select
            id="integration-site"
            v-model="form.site"
            required
            class="input-field font-mono"
            :disabled="saving"
          >
            <option v-for="site in DATADOG_SITES" :key="site" :value="site">{{ site }}</option>
          </select>
        </div>

        <div v-if="form.type === 'pagerduty'">
          <label for="integration-region" class="form-label">{{
            $t('integrations_page.region')
          }}</label>
          <select
            id="integration-region"
            v-model="form.region"
            required
            class="input-field"
            :disabled="saving"
          >
            <option v-for="region in PAGERDUTY_REGIONS" :key="region" :value="region">
              {{ $t(`integrations_page.region_${region === PAGERDUTY_REGIONS[0] ? 'us' : 'eu'}`) }}
            </option>
          </select>
        </div>

        <div v-if="form.type === 'splunk'">
          <label for="integration-index" class="form-label">{{
            $t('integrations_page.index')
          }}</label>
          <input
            id="integration-index"
            v-model="form.index"
            type="text"
            maxlength="100"
            class="input-field font-mono"
            placeholder="main"
            autocomplete="off"
            :disabled="saving"
          />
          <p class="mt-1 text-xs text-theme-textLight">{{ $t('integrations_page.index_hint') }}</p>
        </div>

        <div v-if="form.type !== 'slack'">
          <label for="integration-secret" class="form-label">{{
            $t(`integrations_page.secret_${form.type}`)
          }}</label>
          <input
            id="integration-secret"
            v-model="form.secret"
            type="password"
            :required="secretRequired"
            minlength="8"
            maxlength="512"
            class="input-field font-mono"
            autocomplete="new-password"
            :disabled="saving"
          />
          <p class="mt-1 text-xs text-theme-textLight">{{ secretHint }}</p>
        </div>

        <fieldset>
          <legend class="form-label">{{ $t('integrations_page.events') }}</legend>
          <p class="text-xs text-theme-textLight">{{ $t('integrations_page.events_hint') }}</p>
          <label
            v-for="event in EVENTS"
            :key="event"
            class="mt-2 flex items-start gap-2 text-sm text-theme-text"
          >
            <input
              v-model="form.events"
              type="checkbox"
              :value="event"
              class="mt-0.5 rounded border-theme-border"
              :disabled="saving"
            />
            <span>
              {{ $t(`integrations_page.event_${event.replaceAll('.', '_')}`) }}
              <span class="block text-xs text-theme-textLight">{{
                $t(`integrations_page.event_${event.replaceAll('.', '_')}_desc`)
              }}</span>
            </span>
          </label>
        </fieldset>

        <div
          class="flex items-start justify-between gap-4 rounded-lg border border-theme-border px-4 py-3"
        >
          <div>
            <p id="integration-active-label" class="text-sm font-medium text-theme-text">
              {{
                form.isActive ? $t('integrations_page.enabled') : $t('integrations_page.disabled')
              }}
            </p>
            <p class="text-xs text-theme-textLight">{{ $t('integrations_page.active_hint') }}</p>
          </div>
          <button
            type="button"
            role="switch"
            :aria-checked="form.isActive"
            aria-labelledby="integration-active-label"
            class="relative inline-flex h-6 w-11 shrink-0 items-center rounded-full transition-colors disabled:opacity-50"
            :class="form.isActive ? 'bg-emerald-600' : 'bg-gray-300'"
            :disabled="saving"
            @click="form.isActive = !form.isActive"
          >
            <span
              class="inline-block h-5 w-5 rounded-full bg-white shadow transition-transform"
              :class="form.isActive ? 'translate-x-5' : 'translate-x-0.5'"
            />
          </button>
        </div>

        <div class="flex gap-3 justify-end">
          <button type="button" class="btn-secondary" :disabled="saving" @click="closeForm">
            {{ $t('common.cancel') }}
          </button>
          <button type="submit" class="btn-primary" :disabled="saving">
            {{ saving ? $t('integrations_page.saving') : $t('common.save') }}
          </button>
        </div>
      </form>
    </BaseModal>

    <ConfirmModal
      :open="!!deleteTarget"
      :title="$t('integrations_page.delete_title')"
      :message="$t('integrations_page.delete_confirm', { name: deleteTarget?.name })"
      :error="deleteError"
      :busy="deleting"
      :confirm-label="$t('integrations_page.delete')"
      :busy-label="$t('integrations_page.deleting')"
      @cancel="closeDelete"
      @confirm="confirmDelete"
    />
  </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import AppNav from '@/comp/AppNav.vue'
import ConfirmModal from '@/comp/ConfirmModal.vue'
import { integration_api } from '@/api'
import { useUrlPagination } from '@/lib/pagination'
import { useWorkspaceContext } from '@/lib/permission'
import { errorText } from '@/lib/error'
import BaseModal from '@/comp/BaseModal.vue'

const TYPES = ['webhook', 'slack', 'datadog', 'splunk', 'pagerduty']
const PAGERDUTY_REGIONS = ['events.pagerduty.com', 'events.eu.pagerduty.com']
const DATADOG_SITES = [
  'datadoghq.com',
  'us3.datadoghq.com',
  'us5.datadoghq.com',
  'datadoghq.eu',
  'ap1.datadoghq.com',
  'ddog-gov.com',
]
const EVENTS = [
  'audit.log',
  'agent.health.changed',
  'agent.instance.changed',
  'catalog.changed',
  'gateway.call.denied',
  'system.health.degraded',
]
const URL_PLACEHOLDERS = {
  webhook: 'https://example.com/webhooks/sloth',
  slack: 'https://hooks.slack.com/services/T000/B000/XXXX',
  splunk: 'https://splunk.example.com:8088',
}

const { t } = useI18n()
const route = useRoute()
const { currentWorkspace } = useWorkspaceContext()
const { offset, limit, goToOffset } = useUrlPagination({ limit: 50 })

const integrations = ref([])
const total = ref(0)
const loading = ref(false)
const errorMessage = ref(null)

const showForm = ref(false)
const editing = ref(null)
const isEditing = computed(() => editing.value !== null)
const saving = ref(false)
const formError = ref(null)
const form = ref(emptyForm())

const newSecret = ref(null)
const copied = ref(false)

const togglingId = ref(null)

const deleteTarget = ref(null)
const deleting = ref(false)
const deleteError = ref(null)

// Datadog and Splunk need a credential unless one is already stored for this integration
const secretRequired = computed(
  () => ['datadog', 'splunk', 'pagerduty'].includes(form.value.type) && !editing.value?.hasSecret
)

const secretHint = computed(() => {
  if (editing.value?.hasSecret) return t('integrations_page.secret_keep_hint')
  return form.value.type === 'webhook' ? t('integrations_page.secret_webhook_hint') : ''
})

function emptyForm() {
  return {
    type: 'webhook',
    name: '',
    url: '',
    site: DATADOG_SITES[0],
    region: PAGERDUTY_REGIONS[0],
    index: '',
    secret: '',
    events: ['audit.log'],
    isActive: true,
  }
}

// Maps an integration from the API to the form fields
function toForm(item) {
  return {
    type: item.type,
    name: item.name,
    url: item.url || '',
    site: item.type === 'datadog' ? item.site : DATADOG_SITES[0],
    region: item.type === 'pagerduty' ? item.site : PAGERDUTY_REGIONS[0],
    index: item.index || '',
    secret: '',
    events: [...item.events],
    isActive: item.isActive,
  }
}

async function loadIntegrations() {
  loading.value = true
  errorMessage.value = null

  try {
    const res = await integration_api.list(currentWorkspace.id, {
      limit: limit,
      offset: offset.value,
    })
    integrations.value = res.data.integrations || []
    total.value = res.data._meta?.total || 0
  } catch (err) {
    errorMessage.value = errorText(err, t('integrations_page.failed_load'))
    integrations.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

function goToPage(newOffset) {
  if (newOffset < 0 || newOffset >= total.value) return
  goToOffset(newOffset)
}

function openForm(item = null) {
  editing.value = item
  form.value = item ? toForm(item) : emptyForm()
  formError.value = null
  showForm.value = true
}

function closeForm() {
  showForm.value = false
}

async function submit() {
  formError.value = null
  if (!form.value.events.length) {
    formError.value = t('integrations_page.events_required')
    return
  }

  saving.value = true

  const payload = {
    type: form.value.type,
    name: form.value.name.trim(),
    url: URL_PLACEHOLDERS[form.value.type] ? form.value.url.trim() : '',
    site: { datadog: form.value.site, pagerduty: form.value.region }[form.value.type] || '',
    index: form.value.type === 'splunk' ? form.value.index.trim() : '',
    secret: form.value.type === 'slack' ? '' : form.value.secret,
    events: form.value.events,
    isActive: form.value.isActive,
  }

  try {
    if (isEditing.value) {
      await integration_api.update(currentWorkspace.id, editing.value.id, payload)
    } else {
      const res = await integration_api.create(currentWorkspace.id, payload)
      newSecret.value = res.data.secret || null
      copied.value = false
    }
    showForm.value = false
    await loadIntegrations()
  } catch (err) {
    formError.value = errorText(err, t('integrations_page.failed_save'))
  } finally {
    saving.value = false
  }
}

// Flips one integration on or off; an empty secret keeps the stored one
async function toggle(item) {
  togglingId.value = item.id
  errorMessage.value = null

  try {
    const res = await integration_api.update(currentWorkspace.id, item.id, {
      type: item.type,
      name: item.name,
      url: item.url || '',
      site: item.site || '',
      index: item.index || '',
      secret: '',
      events: item.events,
      isActive: !item.isActive,
    })
    item.isActive = res.data.isActive
  } catch (err) {
    errorMessage.value = errorText(err, t('integrations_page.failed_toggle'))
  } finally {
    togglingId.value = null
  }
}

async function copySecret() {
  try {
    await navigator.clipboard.writeText(newSecret.value)
    copied.value = true
  } catch {
    errorMessage.value = t('integrations_page.copy_failed')
  }
}

function dismissNewSecret() {
  newSecret.value = null
  copied.value = false
}

function closeDelete() {
  deleteTarget.value = null
  deleteError.value = null
}

async function confirmDelete() {
  deleting.value = true
  deleteError.value = null

  try {
    await integration_api.delete(currentWorkspace.id, deleteTarget.value.id)
    deleteTarget.value = null
    await loadIntegrations()
  } catch (err) {
    deleteError.value = errorText(err, t('integrations_page.failed_delete'))
  } finally {
    deleting.value = false
  }
}

watch(() => route.query.page, loadIntegrations)
onMounted(loadIntegrations)
</script>
