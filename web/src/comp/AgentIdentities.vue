<template>
  <section
    id="identities"
    class="scroll-mt-6 rounded-xl border border-theme-border bg-white shadow-sm overflow-hidden"
  >
    <div
      class="flex flex-wrap items-start justify-between gap-4 border-b border-theme-border px-5 py-4"
    >
      <div>
        <h2 class="text-lg font-semibold text-theme-text">
          {{ $t('agents_page.identities_title') }}
        </h2>
        <p class="mt-1 text-sm text-theme-textLight">{{ $t('agents_page.identities_desc') }}</p>
      </div>
      <button
        v-if="canManage && identities.length"
        type="button"
        class="btn-secondary text-sm"
        @click="modalOpen = true"
      >
        {{ $t('agents_page.identity_add') }}
      </button>
    </div>

    <div v-if="created" class="border-b border-amber-200 bg-amber-50 px-5 py-4 space-y-2">
      <p class="text-sm font-medium text-amber-900">
        {{ $t('agents_page.identity_secret_title', { name: created.name }) }}
      </p>
      <p v-if="created.username" class="text-sm text-amber-900">
        {{ $t('agents_page.identity_username') }}:
        <code class="font-mono">{{ created.username }}</code>
      </p>
      <p v-if="created.type === 'client_credentials'" class="text-sm text-amber-900">
        {{ $t('agents_page.identity_client_id') }}:
        <code class="font-mono break-all">{{ created.id }}</code>
      </p>
      <div v-if="created.secret" class="flex flex-wrap items-center gap-2">
        <code
          class="flex-1 min-w-0 px-3 py-2 bg-white border border-amber-200 rounded text-sm text-theme-text break-all font-mono"
          >{{ created.secret }}</code
        >
        <button
          type="button"
          class="btn-secondary shrink-0"
          @click="copy(created.secret, 'identity')"
        >
          {{ copiedField === 'identity' ? $t('agents_page.copied') : $t('agents_page.copy') }}
        </button>
      </div>
      <button
        type="button"
        class="text-sm text-amber-800 hover:text-amber-900 font-medium"
        @click="created = null"
      >
        {{ $t('common.close') }}
      </button>
    </div>

    <div
      v-if="loading && identities.length === 0"
      class="p-8 text-center text-sm text-theme-textLight"
    >
      {{ $t('agents_page.loading_identities') }}
    </div>
    <div v-else-if="identities.length === 0" class="p-5">
      <div class="flex items-start gap-3 rounded-lg border border-dashed border-theme-border p-4">
        <span
          class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-theme-hover text-theme-textLight"
        >
          <svg
            class="h-4 w-4"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            stroke-width="2"
            aria-hidden="true"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z"
            />
          </svg>
        </span>
        <div class="min-w-0">
          <p class="text-sm font-medium text-theme-text">{{ $t('agents_page.no_identities') }}</p>
          <p class="mt-1 text-sm text-theme-textLight">
            {{ $t('agents_page.no_identities_hint') }}
          </p>
          <button
            v-if="canManage"
            type="button"
            class="btn-secondary mt-3 text-sm"
            @click="modalOpen = true"
          >
            {{ $t('agents_page.identity_add') }}
          </button>
        </div>
      </div>
    </div>
    <ul v-else class="divide-y divide-theme-border">
      <li v-for="identity in identities" :key="identity.id" class="px-5 py-3">
        <div class="flex items-start justify-between gap-3">
          <div class="min-w-0">
            <p class="truncate text-sm font-medium text-theme-text font-mono">
              {{ identity.name }}
            </p>
            <p class="mt-0.5 text-xs text-theme-textLight">
              {{ $t(`agents_page.identity_type_${identity.type}`) }}
              <span v-if="identity.username" class="font-mono">· {{ identity.username }}</span>
              <span v-if="identity.tokenTtl"
                >·
                {{
                  $t('agents_page.identity_token_ttl_value', { seconds: identity.tokenTtl })
                }}</span
              >
              · {{ formatRelative(identity.createdAt) }}
            </p>
          </div>
          <span
            class="inline-flex shrink-0 items-center rounded-full px-2 py-0.5 text-xs font-medium"
            :class="
              identity.isActive
                ? 'bg-emerald-50 text-emerald-800'
                : 'bg-theme-hover text-theme-textLight'
            "
          >
            {{
              identity.isActive
                ? $t('agents_page.identity_active')
                : $t('agents_page.identity_inactive')
            }}
          </span>
        </div>
        <div v-if="canManage" class="mt-2 flex items-center gap-3 text-xs">
          <button
            type="button"
            class="text-primary-600 hover:text-primary-700 hover:underline disabled:opacity-50"
            :disabled="toggling === identity.id"
            @click="toggle(identity)"
          >
            {{
              identity.isActive
                ? $t('agents_page.identity_disable')
                : $t('agents_page.identity_enable')
            }}
          </button>
          <button
            type="button"
            class="text-red-600 hover:text-red-700 hover:underline"
            @click="deleteTarget = identity"
          >
            {{ $t('agents_page.identity_revoke') }}
          </button>
        </div>
      </li>
    </ul>

    <AgentIdentityModal
      :open="modalOpen"
      :workspace-id="workspaceId"
      :agent-name="agentName"
      @close="modalOpen = false"
      @saved="onSaved"
    />

    <ConfirmModal
      :open="!!deleteTarget"
      :title="$t('agents_page.identity_delete_title')"
      :message="$t('agents_page.identity_delete_confirm', { name: deleteTarget?.name })"
      :error="deleteError"
      :busy="deleting"
      :confirm-label="$t('agents_page.identity_revoke')"
      :busy-label="$t('agents_page.deleting')"
      @cancel="closeConfirm"
      @confirm="confirmDelete"
    />
  </section>
</template>

<script setup>
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AgentIdentityModal from '@/comp/AgentIdentityModal.vue'
import ConfirmModal from '@/comp/ConfirmModal.vue'
import { agent_identity_api } from '@/api'
import { useCopy } from '@/lib/clipboard'
import { showFlash } from '@/lib/flash'
import { formatRelative } from '@/utils/date'
import { errorText } from '@/lib/error'

const props = defineProps({
  workspaceId: { type: String, required: true },
  agentName: { type: String, required: true },
  canManage: { type: Boolean, default: false },
})

const emit = defineEmits(['error'])

const identities = defineModel('identities', { type: Array, default: () => [] })

const { t } = useI18n()
const { copiedField, copy } = useCopy()
const loading = ref(false)
const modalOpen = ref(false)
const created = ref(null)
const toggling = ref(null)
const deleteTarget = ref(null)
const deleting = ref(false)
const deleteError = ref(null)

async function load() {
  loading.value = true

  try {
    const res = await agent_identity_api.list(props.workspaceId, props.agentName)
    identities.value = res.data.identities || []
  } catch (err) {
    emit('error', errorText(err, t('agents_page.failed_load_identities')))
  } finally {
    loading.value = false
  }
}

// Keep the banner for client credentials even without a secret, since the list does not show the client id
async function onSaved(identity) {
  modalOpen.value = false
  created.value = identity.secret || identity.type === 'client_credentials' ? identity : null
  await load()
}

async function toggle(identity) {
  toggling.value = identity.id

  try {
    const res = await agent_identity_api.update(props.workspaceId, props.agentName, identity.id, {
      name: identity.name,
      isActive: !identity.isActive,
    })
    Object.assign(identity, res.data)
    showFlash(t('common.saved'))
  } catch (err) {
    emit('error', errorText(err, t('agents_page.identity_failed_save')))
  } finally {
    toggling.value = null
  }
}

function closeConfirm() {
  deleteTarget.value = null
  deleteError.value = null
}

async function confirmDelete() {
  deleting.value = true
  deleteError.value = null

  try {
    await agent_identity_api.delete(props.workspaceId, props.agentName, deleteTarget.value.id)
    if (created.value?.id === deleteTarget.value.id) created.value = null
    closeConfirm()
    showFlash(t('common.deleted'))
    await load()
  } catch (err) {
    deleteError.value = errorText(err, t('agents_page.identity_failed_delete'))
  } finally {
    deleting.value = false
  }
}

load()
</script>
