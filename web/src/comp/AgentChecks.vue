<template>
  <section
    id="checks"
    class="scroll-mt-6 rounded-xl border border-theme-border bg-white shadow-sm overflow-hidden"
  >
    <div
      class="flex flex-wrap items-start justify-between gap-4 border-b border-theme-border px-6 py-4"
    >
      <div>
        <h2 class="text-lg font-semibold text-theme-text">{{ $t('agents_page.checks_title') }}</h2>
        <p class="mt-1 text-sm text-theme-textLight">{{ $t('agents_page.checks_desc') }}</p>
      </div>
      <button v-if="canManage" type="button" class="btn-secondary text-sm" @click="openModal(null)">
        {{ $t('agents_page.check_add') }}
      </button>
    </div>

    <div v-if="loading && checks.length === 0" class="p-8 text-center text-sm text-theme-textLight">
      {{ $t('agents_page.loading_checks') }}
    </div>
    <div v-else-if="checks.length === 0" class="p-8 text-center">
      <p class="text-sm text-theme-textLight">{{ $t('agents_page.no_checks') }}</p>
      <p class="mt-1 text-sm text-theme-textLight">{{ $t('agents_page.no_checks_hint') }}</p>
    </div>
    <div v-else class="overflow-x-auto">
      <table class="min-w-full divide-y divide-theme-border">
        <thead class="bg-theme-hover">
          <tr>
            <th
              class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider"
            >
              {{ $t('agents_page.check') }}
            </th>
            <th
              class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider"
            >
              {{ $t('agents_page.check_type') }}
            </th>
            <th
              class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider"
            >
              {{ $t('agents_page.check_settings') }}
            </th>
            <th
              v-if="canManage"
              class="px-6 py-3 text-right text-xs font-medium text-theme-textLight uppercase tracking-wider"
            >
              {{ $t('agents_page.actions') }}
            </th>
          </tr>
        </thead>
        <tbody class="bg-white divide-y divide-theme-border">
          <tr v-for="check in checks" :key="check.checkId" class="hover:bg-theme-hover">
            <td class="px-6 py-4 whitespace-nowrap">
              <p class="text-sm font-medium text-theme-text">{{ check.name }}</p>
              <p class="text-xs font-mono text-theme-textLight">{{ check.checkId }}</p>
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-theme-text">
              {{ $t(`agents_page.check_type_${check.type}`) }}
            </td>
            <td class="px-6 py-4 text-sm text-theme-text font-mono">{{ describeCheck(check) }}</td>
            <td v-if="canManage" class="px-6 py-4 whitespace-nowrap text-right text-sm">
              <span class="inline-flex items-center gap-3">
                <button
                  type="button"
                  class="text-primary-600 hover:text-primary-700 hover:underline"
                  @click="openModal(check)"
                >
                  {{ $t('common.edit') }}
                </button>
                <button
                  type="button"
                  class="text-red-600 hover:text-red-700 hover:underline"
                  @click="deleteTarget = check"
                >
                  {{ $t('agents_page.delete') }}
                </button>
              </span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <AgentCheckModal
      :open="modalOpen"
      :workspace-id="workspaceId"
      :agent-name="agentName"
      :check="editing"
      :existing-ids="checks.map((check) => check.checkId)"
      @close="modalOpen = false"
      @saved="onSaved"
    />

    <ConfirmModal
      :open="!!deleteTarget"
      :title="$t('agents_page.check_delete_title')"
      :message="$t('agents_page.check_delete_confirm', { name: deleteTarget?.name })"
      :error="deleteError"
      :busy="deleting"
      :confirm-label="$t('agents_page.delete')"
      :busy-label="$t('agents_page.deleting')"
      @cancel="closeConfirm"
      @confirm="confirmDelete"
    />
  </section>
</template>

<script setup>
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AgentCheckModal from '@/comp/AgentCheckModal.vue'
import ConfirmModal from '@/comp/ConfirmModal.vue'
import { agent_check_api } from '@/api'
import { showFlash } from '@/lib/flash'
import { errorText } from '@/lib/error'

const props = defineProps({
  workspaceId: { type: String, required: true },
  agentName: { type: String, required: true },
  canManage: { type: Boolean, default: false },
})

const emit = defineEmits(['error', 'changed'])

const checks = defineModel('checks', { type: Array, default: () => [] })

const { t } = useI18n()
const loading = ref(false)
const modalOpen = ref(false)
const editing = ref(null)
const deleteTarget = ref(null)
const deleting = ref(false)
const deleteError = ref(null)

async function load() {
  loading.value = true

  try {
    const res = await agent_check_api.list(props.workspaceId, props.agentName)
    checks.value = res.data.checks || []
  } catch (err) {
    emit('error', errorText(err, t('agents_page.failed_load_checks')))
  } finally {
    loading.value = false
  }
}

function openModal(check) {
  editing.value = check
  modalOpen.value = true
}

// Template changes alter instance statuses, so the page refreshes the agent and instances too
async function onSaved() {
  modalOpen.value = false
  showFlash(t('common.saved'))

  emit('changed')
  await load()
}

function closeConfirm() {
  deleteTarget.value = null
  deleteError.value = null
}

async function confirmDelete() {
  deleting.value = true
  deleteError.value = null

  try {
    await agent_check_api.delete(props.workspaceId, props.agentName, deleteTarget.value.checkId)
  } catch (err) {
    deleteError.value = errorText(err, t('agents_page.check_failed_delete'))
    return
  } finally {
    deleting.value = false
  }

  closeConfirm()
  showFlash(t('common.deleted'))

  emit('changed')
  await load()
}

function describeCheck(check) {
  if (check.type === 'ttl') {
    return t('agents_page.check_describe_ttl', { ttl: check.ttl })
  }
  const port = check.port ? `:${check.port}` : ''
  const target = check.type === 'http' ? `${check.scheme}://…${port}${check.path}` : `tcp …${port}`
  return t('agents_page.check_describe_probe', {
    target,
    interval: check.interval,
    timeout: check.timeout,
  })
}

load()
</script>
