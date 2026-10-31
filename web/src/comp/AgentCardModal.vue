<template>
  <BaseModal
    :open="open"
    :title="isEdit ? $t('agents_page.edit_card_title') : $t('agents_page.register_title')"
    :description="isEdit ? $t('agents_page.edit_card_desc') : $t('agents_page.register_desc')"
    :error="error"
    :busy="saving"
    @close="close"
  >
    <form class="mt-4 space-y-4" @submit.prevent="submit">
      <div class="space-y-4">
        <div>
          <label for="agent-card-name" class="form-label">{{ $t('agents_page.name') }}</label>
          <input
            id="agent-card-name"
            v-model="form.name"
            type="text"
            required
            class="input-field font-mono"
            placeholder="support-assistant"
            autocomplete="off"
            :disabled="saving || isEdit"
          />
          <p class="mt-1 text-xs text-theme-textLight">{{ $t('agents_page.name_hint') }}</p>
        </div>
        <div>
          <label for="agent-card-version" class="form-label">{{ $t('agents_page.version') }}</label>
          <input
            id="agent-card-version"
            v-model="form.version"
            type="text"
            required
            maxlength="40"
            class="input-field font-mono"
            placeholder="1.0.0"
            autocomplete="off"
            :disabled="saving"
          />
          <p class="mt-1 text-xs text-theme-textLight">{{ $t('agents_page.version_hint') }}</p>
        </div>
        <div>
          <label for="agent-card-path" class="form-label">{{ $t('agents_page.card_path') }}</label>
          <input
            id="agent-card-path"
            v-model="form.cardPath"
            type="text"
            maxlength="255"
            pattern="/.*"
            :title="$t('agents_page.card_path_hint')"
            class="input-field font-mono"
            :placeholder="AGENT_CARD_PATH"
            autocomplete="off"
            :disabled="saving"
          />
          <p class="mt-1 text-xs text-theme-textLight">{{ $t('agents_page.card_path_hint') }}</p>
        </div>
        <div>
          <label for="agent-lb-strategy" class="form-label">{{
            $t('agents_page.lb_strategy')
          }}</label>
          <select
            id="agent-lb-strategy"
            v-model="form.lbStrategy"
            required
            class="input-field"
            :disabled="saving"
          >
            <option value="round_robin">{{ $t('agents_page.lb_strategy_round_robin') }}</option>
            <option value="random">{{ $t('agents_page.lb_strategy_random') }}</option>
            <option value="least_outstanding_tasks">
              {{ $t('agents_page.lb_strategy_least_outstanding_tasks') }}
            </option>
            <option value="sticky">{{ $t('agents_page.lb_strategy_sticky') }}</option>
          </select>
          <p class="mt-1 text-xs text-theme-textLight">{{ $t('agents_page.lb_strategy_hint') }}</p>
        </div>
      </div>

      <div class="flex gap-3 justify-end">
        <button type="button" class="btn-secondary" :disabled="saving" @click="close">
          {{ $t('common.cancel') }}
        </button>
        <button type="submit" class="btn-primary" :disabled="saving">
          {{
            saving
              ? $t('agents_page.saving')
              : isEdit
                ? $t('common.save')
                : $t('agents_page.register')
          }}
        </button>
      </div>
    </form>
  </BaseModal>
</template>

<script setup>
import { computed, reactive, ref, watch } from 'vue'
import BaseModal from '@/comp/BaseModal.vue'
import { useI18n } from 'vue-i18n'
import { agent_api } from '@/api'
import { AGENT_CARD_PATH, AGENT_NAME_PATTERN } from '@/lib/agent'
import { errorText } from '@/lib/error'

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
  workspaceId: {
    type: String,
    required: true,
  },
  // Existing agent to edit; omit to register a new one.
  agent: {
    type: Object,
    default: null,
  },
})

const emit = defineEmits(['close', 'saved'])

const { t } = useI18n()
const isEdit = computed(() => !!props.agent)
const saving = ref(false)
const error = ref(null)
const form = reactive({
  name: '',
  version: '',
  cardPath: '',
  lbStrategy: 'round_robin',
})

watch(
  () => props.open,
  (open) => {
    if (!open) return
    error.value = null
    form.name = props.agent?.name || ''
    form.version = props.agent?.version || ''
    form.cardPath = props.agent?.cardPath || ''
    form.lbStrategy = props.agent?.lbStrategy || 'round_robin'
  }
)

function close() {
  if (saving.value) return
  emit('close')
}

async function agentExists(name) {
  try {
    await agent_api.get(props.workspaceId, name)
    return true
  } catch (err) {
    if (err.response?.status === 404) return false
    throw err
  }
}

async function submit() {
  const name = form.name.trim()
  if (!AGENT_NAME_PATTERN.test(name)) {
    error.value = t('agents_page.invalid_name')
    return
  }

  const version = form.version.trim()
  if (!version) {
    error.value = t('agents_page.version_required')
    return
  }

  const cardPath = form.cardPath.trim()
  if (cardPath && !cardPath.startsWith('/')) {
    error.value = t('agents_page.invalid_card_path')
    return
  }

  saving.value = true
  error.value = null

  try {
    // PUT is an upsert, so guard against replacing an existing agent from the register form
    if (!isEdit.value && (await agentExists(name))) {
      error.value = t('agents_page.already_exists', { name })
      return
    }

    const payload = { version, lbStrategy: form.lbStrategy }
    if (cardPath) payload.cardPath = cardPath
    const res = await agent_api.upsert(props.workspaceId, name, payload)
    emit('saved', res.data)
  } catch (err) {
    error.value = errorText(err, t('agents_page.failed_save'))
  } finally {
    saving.value = false
  }
}
</script>
