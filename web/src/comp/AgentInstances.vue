<template>
  <section
    id="instances"
    class="scroll-mt-6 rounded-xl border border-theme-border bg-white shadow-sm overflow-hidden"
  >
    <div
      class="flex flex-wrap items-start justify-between gap-4 border-b border-theme-border px-6 py-4"
    >
      <div>
        <h2 class="text-lg font-semibold text-theme-text">{{ $t('agents_page.instances') }}</h2>
        <p class="mt-1 text-sm text-theme-textLight">{{ $t('agents_page.instances_desc') }}</p>
      </div>
      <button
        type="button"
        class="btn-secondary text-sm disabled:opacity-50"
        :disabled="loading"
        @click="load"
      >
        {{ $t('agents_page.refresh') }}
      </button>
    </div>

    <div
      v-if="loading && instances.length === 0"
      class="p-12 text-center text-sm text-theme-textLight"
    >
      {{ $t('agents_page.loading_instances') }}
    </div>
    <div v-else-if="instances.length === 0" class="p-12 text-center">
      <p class="text-sm text-theme-textLight">{{ $t('agents_page.no_instances') }}</p>
      <p class="mt-1 text-sm text-theme-textLight">{{ $t('agents_page.no_instances_hint') }}</p>
    </div>
    <div v-else class="overflow-x-auto">
      <table class="min-w-full divide-y divide-theme-border">
        <thead class="bg-theme-hover">
          <tr>
            <th
              class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider"
            >
              {{ $t('agents_page.instance') }}
            </th>
            <th
              class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider"
            >
              {{ $t('agents_page.address') }}
            </th>
            <th
              class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider"
            >
              {{ $t('agents_page.health') }}
            </th>
            <th
              class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider"
            >
              {{ $t('agents_page.checks_title') }}
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
          <template v-for="instance in instances" :key="instance.id">
            <tr class="hover:bg-theme-hover">
              <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-theme-text font-mono">
                {{ instance.instanceId }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-theme-text font-mono">
                {{ instance.address }}:{{ instance.port }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm">
                <span
                  class="inline-flex items-center gap-1.5 rounded-full px-2 py-0.5 text-xs font-medium"
                  :class="healthClass(instance.status)"
                >
                  <span
                    class="h-1.5 w-1.5 rounded-full"
                    :class="healthDotClass(instance.status)"
                    aria-hidden="true"
                  />
                  {{ $t(`agents_page.health_${instance.status}`) }}
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm">
                <span v-if="!(instance.checks || []).length" class="text-theme-textLight">{{
                  $t('agents_page.instance_no_checks')
                }}</span>
                <button
                  v-else
                  type="button"
                  class="inline-flex items-center gap-1 text-primary-600 hover:text-primary-700 hover:underline"
                  :aria-expanded="!!expanded[instance.id]"
                  @click="expanded[instance.id] = !expanded[instance.id]"
                >
                  <StatusBreakdown :counts="countByStatus(instance.checks)" hide-zero />
                  <svg
                    class="h-3.5 w-3.5 transition-transform"
                    :class="expanded[instance.id] ? 'rotate-180' : ''"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                    aria-hidden="true"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M19 9l-7 7-7-7"
                    />
                  </svg>
                </button>
              </td>
              <td v-if="canManage" class="px-6 py-4 whitespace-nowrap text-right text-sm">
                <button
                  type="button"
                  class="text-red-600 hover:text-red-700 hover:underline"
                  @click="deregisterTarget = instance"
                >
                  {{ $t('agents_page.deregister') }}
                </button>
              </td>
            </tr>
            <tr v-if="expanded[instance.id]">
              <td :colspan="canManage ? 5 : 4" class="bg-theme-hover px-6 py-4">
                <ul class="space-y-3">
                  <li
                    v-for="check in instance.checks"
                    :key="check.checkId"
                    class="grid gap-1 sm:grid-cols-[220px_110px_1fr]"
                  >
                    <div>
                      <p class="text-sm font-medium text-theme-text">{{ check.name }}</p>
                      <p class="text-xs font-mono text-theme-textLight">
                        {{ check.checkId }} · {{ $t(`agents_page.check_type_${check.type}`) }}
                      </p>
                    </div>
                    <div>
                      <span
                        class="inline-flex items-center gap-1.5 rounded-full px-2 py-0.5 text-xs font-medium"
                        :class="healthClass(check.status)"
                      >
                        <span
                          class="h-1.5 w-1.5 rounded-full"
                          :class="healthDotClass(check.status)"
                          aria-hidden="true"
                        />
                        {{ $t(`agents_page.health_${check.status}`) }}
                      </span>
                    </div>
                    <div class="min-w-0">
                      <p class="text-sm text-theme-text break-words font-mono">
                        {{ check.output || '—' }}
                      </p>
                      <p class="text-xs text-theme-textLight">
                        {{
                          $t('agents_page.check_last_run', {
                            when: formatRelative(check.lastRunAt, $t('agents_page.check_never')),
                          })
                        }}
                      </p>
                    </div>
                  </li>
                </ul>
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

    <ConfirmModal
      :open="!!deregisterTarget"
      :title="$t('agents_page.deregister_title')"
      :message="$t('agents_page.deregister_confirm', { id: deregisterTarget?.instanceId })"
      :error="deleteError"
      :busy="deleting"
      :confirm-label="$t('agents_page.deregister')"
      :busy-label="$t('agents_page.deregistering')"
      @cancel="closeConfirm"
      @confirm="confirmDeregister"
    />
  </section>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import ConfirmModal from '@/comp/ConfirmModal.vue'
import StatusBreakdown from '@/comp/StatusBreakdown.vue'
import { agent_instance_api } from '@/api'
import { showFlash } from '@/lib/flash'
import { countByStatus, healthClass, healthDotClass } from '@/lib/agent'
import { formatRelative } from '@/utils/date'
import { errorText } from '@/lib/error'

const props = defineProps({
  workspaceId: { type: String, required: true },
  agentName: { type: String, required: true },
  canManage: { type: Boolean, default: false },
})

const emit = defineEmits(['error', 'deregistered'])

const total = defineModel('total', { type: Number, default: 0 })

const { t } = useI18n()
const limit = 20
const instances = ref([])
const offset = ref(0)
const loading = ref(false)
const expanded = reactive({})
const deregisterTarget = ref(null)
const deleting = ref(false)
const deleteError = ref(null)

async function load() {
  loading.value = true

  try {
    const res = await agent_instance_api.list(props.workspaceId, props.agentName, {
      limit,
      offset: offset.value,
    })
    instances.value = res.data.instances || []
    total.value = res.data._meta?.total || 0

    // The page emptied after a deregister, so step back to the last page with rows
    if (!instances.value.length && offset.value > 0) {
      offset.value = Math.max(0, offset.value - limit)
      await load()
    }
  } catch (err) {
    emit('error', errorText(err, t('agents_page.failed_load_instances')))
  } finally {
    loading.value = false
  }
}

function goToPage(newOffset) {
  if (newOffset < 0 || newOffset >= total.value) return
  offset.value = newOffset
  load()
}

function closeConfirm() {
  deregisterTarget.value = null
  deleteError.value = null
}

async function confirmDeregister() {
  deleting.value = true
  deleteError.value = null

  try {
    await agent_instance_api.delete(
      props.workspaceId,
      props.agentName,
      deregisterTarget.value.instanceId
    )
  } catch (err) {
    deleteError.value = errorText(err, t('agents_page.failed_deregister'))
    return
  } finally {
    deleting.value = false
  }

  closeConfirm()
  showFlash(t('common.deleted'))

  // The agent's health is derived from the remaining instances
  emit('deregistered')
  await load()
}

defineExpose({ load })

load()
</script>
