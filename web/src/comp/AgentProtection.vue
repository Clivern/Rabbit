<template>
  <section
    id="protection"
    class="scroll-mt-6 rounded-xl border border-theme-border bg-white shadow-sm overflow-hidden"
  >
    <div
      class="flex flex-wrap items-start justify-between gap-4 border-b border-theme-border px-5 py-4"
    >
      <div>
        <h2 class="text-lg font-semibold text-theme-text">
          {{ $t('agents_page.protection_title') }}
        </h2>
        <p class="mt-1 text-sm text-theme-textLight">{{ $t('agents_page.protection_desc') }}</p>
      </div>
      <button
        v-if="canManage && protection"
        type="button"
        class="btn-secondary text-sm"
        @click="modalOpen = true"
      >
        {{ $t('agents_page.protection_replace') }}
      </button>
    </div>

    <div v-if="newSecret" class="border-b border-amber-200 bg-amber-50 px-5 py-4 space-y-2">
      <p class="text-sm font-medium text-amber-900">
        {{ $t('agents_page.protection_secret_title') }}
      </p>
      <div class="flex flex-wrap items-center gap-2">
        <code
          class="flex-1 min-w-0 px-3 py-2 bg-white border border-amber-200 rounded text-sm text-theme-text break-all font-mono"
          >{{ newSecret }}</code
        >
        <button type="button" class="btn-secondary shrink-0" @click="copy(newSecret, 'protection')">
          {{ copiedField === 'protection' ? $t('agents_page.copied') : $t('agents_page.copy') }}
        </button>
      </div>
      <button
        type="button"
        class="text-sm text-amber-800 hover:text-amber-900 font-medium"
        @click="newSecret = null"
      >
        {{ $t('common.close') }}
      </button>
    </div>

    <div v-if="loading && !protection" class="p-8 text-center text-sm text-theme-textLight">
      {{ $t('agents_page.loading_protection') }}
    </div>
    <div v-else-if="!protection" class="p-5">
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
              d="M12 3l7 3v6c0 4.5-3 8-7 9-4-1-7-4.5-7-9V6l7-3z"
            />
          </svg>
        </span>
        <div class="min-w-0">
          <p class="text-sm font-medium text-theme-text">{{ $t('agents_page.no_protection') }}</p>
          <p class="mt-1 text-sm text-theme-textLight">
            {{ $t('agents_page.no_protection_hint') }}
          </p>
          <button
            v-if="canManage"
            type="button"
            class="btn-secondary mt-3 text-sm"
            @click="modalOpen = true"
          >
            {{ $t('agents_page.protection_add') }}
          </button>
        </div>
      </div>
    </div>
    <div v-else class="space-y-4 px-5 py-4">
      <dl class="grid grid-cols-2 gap-x-6 gap-y-3 text-sm">
        <div>
          <dt class="text-theme-textLight">{{ $t('agents_page.protection_type') }}</dt>
          <dd class="text-theme-text">
            {{ $t(`agents_page.protection_type_${protection.type}`) }}
          </dd>
        </div>
        <div>
          <dt class="text-theme-textLight">{{ $t('agents_page.protection_status') }}</dt>
          <dd>
            <span
              class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium"
              :class="
                protection.isActive
                  ? 'bg-green-50 text-green-700'
                  : 'bg-theme-hover text-theme-textLight'
              "
            >
              {{
                protection.isActive
                  ? $t('agents_page.protection_active')
                  : $t('agents_page.protection_inactive')
              }}
            </span>
          </dd>
        </div>
        <div v-if="protection.header">
          <dt class="text-theme-textLight">{{ $t('agents_page.protection_header') }}</dt>
          <dd class="text-theme-text font-mono">{{ protection.header }}</dd>
        </div>
        <div v-if="protection.username">
          <dt class="text-theme-textLight">{{ $t('agents_page.protection_username') }}</dt>
          <dd class="text-theme-text font-mono">{{ protection.username }}</dd>
        </div>
        <div>
          <dt class="text-theme-textLight">{{ $t('agents_page.protection_updated') }}</dt>
          <dd class="text-theme-text">{{ formatRelative(protection.updatedAt) }}</dd>
        </div>
      </dl>
      <div
        v-if="canManage"
        class="flex items-center gap-3 border-t border-theme-border pt-3 text-sm"
      >
        <button
          type="button"
          class="text-primary-600 hover:text-primary-700 hover:underline disabled:opacity-50"
          :disabled="toggling"
          @click="toggle"
        >
          {{
            protection.isActive
              ? $t('agents_page.protection_disable')
              : $t('agents_page.protection_enable')
          }}
        </button>
        <button
          type="button"
          class="text-red-600 hover:text-red-700 hover:underline"
          @click="deleteOpen = true"
        >
          {{ $t('agents_page.protection_remove') }}
        </button>
      </div>
    </div>

    <AgentProtectionModal
      :open="modalOpen"
      :workspace-id="workspaceId"
      :agent-name="agentName"
      :current="protection"
      @close="modalOpen = false"
      @saved="onSaved"
    />

    <ConfirmModal
      :open="deleteOpen"
      :title="$t('agents_page.protection_delete_title')"
      :message="$t('agents_page.protection_delete_confirm')"
      :error="deleteError"
      :busy="deleting"
      :confirm-label="$t('agents_page.protection_remove')"
      :busy-label="$t('agents_page.deleting')"
      @cancel="closeConfirm"
      @confirm="confirmDelete"
    />
  </section>
</template>

<script setup>
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AgentProtectionModal from '@/comp/AgentProtectionModal.vue'
import ConfirmModal from '@/comp/ConfirmModal.vue'
import { agent_protection_api } from '@/api'
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

const protection = defineModel('protection', { type: Object, default: null })

const { t } = useI18n()
const { copiedField, copy } = useCopy()
const loading = ref(false)
const modalOpen = ref(false)
const newSecret = ref(null)
const toggling = ref(false)
const deleteOpen = ref(false)
const deleting = ref(false)
const deleteError = ref(null)

async function load() {
  loading.value = true

  try {
    const res = await agent_protection_api.get(props.workspaceId, props.agentName)
    protection.value = res.data
  } catch (err) {
    if (err.response?.status === 404) {
      protection.value = null
      return
    }
    emit('error', errorText(err, t('agents_page.failed_load_protection')))
  } finally {
    loading.value = false
  }
}

function onSaved(saved) {
  modalOpen.value = false
  newSecret.value = saved.secret || null
  protection.value = { ...saved, secret: undefined }
  showFlash(t('common.saved'))
}

async function toggle() {
  toggling.value = true

  try {
    const res = await agent_protection_api.updateStatus(props.workspaceId, props.agentName, {
      isActive: !protection.value.isActive,
    })
    protection.value = res.data
    showFlash(t('common.saved'))
  } catch (err) {
    emit('error', errorText(err, t('agents_page.protection_failed_save')))
  } finally {
    toggling.value = false
  }
}

function closeConfirm() {
  deleteOpen.value = false
  deleteError.value = null
}

async function confirmDelete() {
  deleting.value = true
  deleteError.value = null

  try {
    await agent_protection_api.delete(props.workspaceId, props.agentName)
    protection.value = null
    newSecret.value = null
    closeConfirm()
    showFlash(t('common.deleted'))
  } catch (err) {
    deleteError.value = errorText(err, t('agents_page.protection_failed_delete'))
  } finally {
    deleting.value = false
  }
}

load()
</script>
