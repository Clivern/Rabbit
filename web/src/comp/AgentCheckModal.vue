<template>
  <BaseModal
    :open="open"
    :title="isEdit ? $t('agents_page.check_edit_title') : $t('agents_page.check_add_title')"
    :description="$t('agents_page.check_modal_desc')"
    :error="error"
    :busy="saving"
    @close="close"
  >
    <form class="mt-4 space-y-4" @submit.prevent="submit">
      <div>
        <label for="check-id" class="form-label">{{ $t('agents_page.check_id') }}</label>
        <input
          id="check-id"
          v-model="form.checkId"
          type="text"
          required
          class="input-field font-mono"
          placeholder="http"
          autocomplete="off"
          :disabled="saving || isEdit"
        />
        <p class="mt-1 text-xs text-theme-textLight">{{ $t('agents_page.check_id_hint') }}</p>
      </div>
      <div>
        <label for="check-name" class="form-label">{{ $t('agents_page.name') }}</label>
        <input
          id="check-name"
          v-model="form.name"
          type="text"
          maxlength="120"
          class="input-field"
          :placeholder="$t('agents_page.check_name_placeholder')"
          autocomplete="off"
          :disabled="saving"
        />
        <p class="mt-1 text-xs text-theme-textLight">{{ $t('agents_page.check_name_hint') }}</p>
      </div>

      <div>
        <label for="check-type" class="form-label">{{ $t('agents_page.check_type') }}</label>
        <select id="check-type" v-model="form.type" class="input-field" :disabled="saving">
          <option value="http">{{ $t('agents_page.check_type_http') }}</option>
          <option value="tcp">{{ $t('agents_page.check_type_tcp') }}</option>
          <option value="ttl">{{ $t('agents_page.check_type_ttl') }}</option>
        </select>
        <p class="mt-1 text-xs text-theme-textLight">
          {{ $t(`agents_page.check_type_${form.type}_hint`) }}
        </p>
      </div>

      <div v-if="form.type === 'http'" class="grid gap-4 sm:grid-cols-[120px_1fr]">
        <div>
          <label for="check-scheme" class="form-label">{{ $t('agents_page.check_scheme') }}</label>
          <select id="check-scheme" v-model="form.scheme" class="input-field" :disabled="saving">
            <option value="http">http</option>
            <option value="https">https</option>
          </select>
        </div>
        <div>
          <label for="check-path" class="form-label">{{ $t('agents_page.check_path') }}</label>
          <input
            id="check-path"
            v-model="form.path"
            type="text"
            required
            maxlength="255"
            class="input-field font-mono"
            placeholder="/healthz"
            autocomplete="off"
            :disabled="saving"
          />
        </div>
      </div>

      <div v-if="form.type !== 'ttl'" class="grid gap-4 sm:grid-cols-3">
        <div>
          <label for="check-port" class="form-label">{{ $t('agents_page.check_port') }}</label>
          <input
            id="check-port"
            v-model.number="form.port"
            type="number"
            min="1"
            max="65535"
            class="input-field"
            :placeholder="$t('agents_page.check_port_placeholder')"
            :disabled="saving"
          />
        </div>
        <div>
          <label for="check-interval" class="form-label">{{
            $t('agents_page.check_interval')
          }}</label>
          <input
            id="check-interval"
            v-model.number="form.interval"
            type="number"
            min="5"
            max="3600"
            required
            class="input-field"
            :disabled="saving"
          />
        </div>
        <div>
          <label for="check-timeout" class="form-label">{{
            $t('agents_page.check_timeout')
          }}</label>
          <input
            id="check-timeout"
            v-model.number="form.timeout"
            type="number"
            min="1"
            max="60"
            required
            class="input-field"
            :disabled="saving"
          />
        </div>
      </div>

      <div v-else>
        <label for="check-ttl" class="form-label">{{ $t('agents_page.check_ttl') }}</label>
        <input
          id="check-ttl"
          v-model.number="form.ttl"
          type="number"
          min="5"
          max="86400"
          required
          class="input-field"
          :disabled="saving"
        />
        <p class="mt-1 text-xs text-theme-textLight">{{ $t('agents_page.check_ttl_hint') }}</p>
      </div>

      <div class="flex gap-3 justify-end">
        <button type="button" class="btn-secondary" :disabled="saving" @click="close">
          {{ $t('common.cancel') }}
        </button>
        <button type="submit" class="btn-primary" :disabled="saving">
          {{ saving ? $t('agents_page.saving') : $t('common.save') }}
        </button>
      </div>
    </form>
  </BaseModal>
</template>

<script setup>
import { computed, reactive, ref, watch } from 'vue'
import BaseModal from '@/comp/BaseModal.vue'
import { useI18n } from 'vue-i18n'
import { agent_check_api } from '@/api'
import { AGENT_NAME_PATTERN } from '@/lib/agent'
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
  agentName: {
    type: String,
    required: true,
  },
  // Existing check template to edit; omit to add a new one.
  check: {
    type: Object,
    default: null,
  },
  // Ids already in use, so adding does not silently replace a check.
  existingIds: {
    type: Array,
    default: () => [],
  },
})

const emit = defineEmits(['close', 'saved'])

const { t } = useI18n()
const isEdit = computed(() => !!props.check)
const saving = ref(false)
const error = ref(null)
const form = reactive({
  checkId: '',
  name: '',
  type: 'http',
  scheme: 'http',
  path: '/healthz',
  port: null,
  interval: 10,
  timeout: 2,
  ttl: 60,
})

watch(
  () => props.open,
  (open) => {
    if (!open) return
    const check = props.check
    error.value = null
    form.checkId = check?.checkId || ''
    form.name = check?.name || ''
    form.type = check?.type || 'http'
    form.scheme = check?.scheme || 'http'
    form.path = check?.path || '/healthz'
    form.port = check?.port || null
    form.interval = check?.interval || 10
    form.timeout = check?.timeout || 2
    form.ttl = check?.ttl || 60
  }
)

function close() {
  if (saving.value) return
  emit('close')
}

function payload() {
  const body = { name: form.name.trim(), type: form.type }
  if (form.type === 'ttl') {
    body.ttl = form.ttl
    return body
  }
  body.interval = form.interval
  body.timeout = form.timeout
  if (form.port) body.port = form.port
  if (form.type === 'http') {
    body.scheme = form.scheme
    body.path = form.path.trim()
  }
  return body
}

async function submit() {
  const checkId = form.checkId.trim()
  if (!AGENT_NAME_PATTERN.test(checkId)) {
    error.value = t('agents_page.check_invalid_id')
    return
  }
  if (!isEdit.value && props.existingIds.includes(checkId)) {
    error.value = t('agents_page.check_already_exists', { id: checkId })
    return
  }
  if (form.type !== 'ttl' && form.timeout >= form.interval) {
    error.value = t('agents_page.check_invalid_timing')
    return
  }
  if (form.type === 'http' && !form.path.trim().startsWith('/')) {
    error.value = t('agents_page.check_invalid_path')
    return
  }

  saving.value = true
  error.value = null

  try {
    const res = await agent_check_api.upsert(props.workspaceId, props.agentName, checkId, payload())
    emit('saved', res.data)
  } catch (err) {
    error.value = errorText(err, t('agents_page.check_failed_save'))
  } finally {
    saving.value = false
  }
}
</script>
