<template>
  <BaseModal
    :open="open"
    :title="
      current ? $t('agents_page.protection_replace_title') : $t('agents_page.protection_add_title')
    "
    :description="
      current ? $t('agents_page.protection_replace_desc') : $t('agents_page.protection_modal_desc')
    "
    :error="error"
    :busy="saving"
    @close="close"
  >
    <form class="mt-4 space-y-4" @submit.prevent="submit">
      <div>
        <label for="protection-type" class="form-label">{{
          $t('agents_page.protection_type')
        }}</label>
        <select id="protection-type" v-model="form.type" class="input-field" :disabled="saving">
          <option value="api_key">{{ $t('agents_page.protection_type_api_key') }}</option>
          <option value="basic_auth">{{ $t('agents_page.protection_type_basic_auth') }}</option>
        </select>
        <p class="mt-1 text-xs text-theme-textLight">
          {{ $t(`agents_page.protection_type_${form.type}_hint`) }}
        </p>
      </div>

      <div v-if="form.type === 'api_key'">
        <label for="protection-header" class="form-label">{{
          $t('agents_page.protection_header')
        }}</label>
        <input
          id="protection-header"
          v-model="form.header"
          type="text"
          maxlength="100"
          class="input-field font-mono"
          placeholder="X-API-Key"
          autocomplete="off"
          :disabled="saving"
        />
      </div>

      <div v-if="form.type === 'basic_auth'">
        <label for="protection-username" class="form-label">{{
          $t('agents_page.protection_username')
        }}</label>
        <input
          id="protection-username"
          v-model="form.username"
          type="text"
          required
          maxlength="100"
          class="input-field font-mono"
          placeholder="sloth"
          autocomplete="off"
          :disabled="saving"
        />
      </div>

      <div>
        <label for="protection-secret" class="form-label">{{
          $t('agents_page.protection_secret')
        }}</label>
        <input
          id="protection-secret"
          v-model="form.secret"
          type="password"
          minlength="8"
          maxlength="512"
          class="input-field font-mono"
          :placeholder="$t('agents_page.protection_secret_placeholder')"
          autocomplete="new-password"
          :disabled="saving"
        />
        <p class="mt-1 text-xs text-theme-textLight">
          {{ $t('agents_page.protection_secret_hint') }}
        </p>
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
import { reactive, ref, watch } from 'vue'
import BaseModal from '@/comp/BaseModal.vue'
import { useI18n } from 'vue-i18n'
import { agent_protection_api } from '@/api'
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
  current: {
    type: Object,
    default: null,
  },
})

const emit = defineEmits(['close', 'saved'])

const { t } = useI18n()
const saving = ref(false)
const error = ref(null)
const form = reactive({
  type: 'api_key',
  header: '',
  username: '',
  secret: '',
})

watch(
  () => props.open,
  (open) => {
    if (!open) return
    error.value = null
    form.type = props.current?.type || 'api_key'
    form.header = props.current?.header || 'X-API-Key'
    form.username = props.current?.username || 'sloth'
    form.secret = ''
  }
)

function close() {
  if (saving.value) return
  emit('close')
}

async function submit() {
  saving.value = true
  error.value = null

  const body = { type: form.type }
  if (form.type === 'api_key') body.header = form.header.trim()
  if (form.type === 'basic_auth') body.username = form.username.trim()
  if (form.secret) body.secret = form.secret

  try {
    const res = await agent_protection_api.save(props.workspaceId, props.agentName, body)
    emit('saved', res.data)
  } catch (err) {
    error.value = errorText(err, t('agents_page.protection_failed_save'))
  } finally {
    saving.value = false
  }
}
</script>
