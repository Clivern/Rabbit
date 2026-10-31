<template>
  <BaseModal
    :open="open"
    :title="$t('agents_page.identity_add_title')"
    :description="$t('agents_page.identity_modal_desc')"
    :error="error"
    :busy="saving"
    @close="close"
  >
    <form class="mt-4 space-y-4" @submit.prevent="submit">
      <div>
        <label for="identity-name" class="form-label">{{ $t('agents_page.name') }}</label>
        <input
          id="identity-name"
          v-model="form.name"
          type="text"
          required
          maxlength="100"
          class="input-field font-mono"
          :placeholder="$t('agents_page.identity_name_placeholder')"
          autocomplete="off"
          :disabled="saving"
        />
        <p class="mt-1 text-xs text-theme-textLight">{{ $t('agents_page.identity_name_hint') }}</p>
      </div>

      <div>
        <label for="identity-type" class="form-label">{{ $t('agents_page.identity_type') }}</label>
        <select id="identity-type" v-model="form.type" class="input-field" :disabled="saving">
          <option value="api_key">{{ $t('agents_page.identity_type_api_key') }}</option>
          <option value="basic_auth">{{ $t('agents_page.identity_type_basic_auth') }}</option>
          <option value="client_credentials">
            {{ $t('agents_page.identity_type_client_credentials') }}
          </option>
        </select>
        <p class="mt-1 text-xs text-theme-textLight">
          {{ $t(`agents_page.identity_type_${form.type}_hint`) }}
        </p>
      </div>

      <div v-if="form.type === 'basic_auth'">
        <label for="identity-username" class="form-label">{{
          $t('agents_page.identity_username')
        }}</label>
        <input
          id="identity-username"
          v-model="form.username"
          type="text"
          required
          maxlength="100"
          class="input-field font-mono"
          :placeholder="agentName"
          autocomplete="off"
          :disabled="saving"
        />
      </div>

      <div v-if="form.type === 'client_credentials'">
        <label for="identity-token-ttl" class="form-label">{{
          $t('agents_page.identity_token_ttl')
        }}</label>
        <input
          id="identity-token-ttl"
          v-model.number="form.tokenTtl"
          type="number"
          min="60"
          max="86400"
          required
          class="input-field"
          :disabled="saving"
        />
        <p class="mt-1 text-xs text-theme-textLight">
          {{ $t('agents_page.identity_token_ttl_hint') }}
        </p>
      </div>

      <div>
        <label for="identity-secret" class="form-label">{{
          $t('agents_page.identity_secret')
        }}</label>
        <input
          id="identity-secret"
          v-model="form.secret"
          type="password"
          minlength="8"
          maxlength="512"
          class="input-field font-mono"
          :placeholder="$t('agents_page.identity_secret_placeholder')"
          autocomplete="new-password"
          :disabled="saving"
        />
        <p class="mt-1 text-xs text-theme-textLight">
          {{ $t('agents_page.identity_secret_hint') }}
        </p>
      </div>

      <div class="flex gap-3 justify-end">
        <button type="button" class="btn-secondary" :disabled="saving" @click="close">
          {{ $t('common.cancel') }}
        </button>
        <button type="submit" class="btn-primary" :disabled="saving">
          {{ saving ? $t('agents_page.saving') : $t('common.create') }}
        </button>
      </div>
    </form>
  </BaseModal>
</template>

<script setup>
import { reactive, ref, watch } from 'vue'
import BaseModal from '@/comp/BaseModal.vue'
import { useI18n } from 'vue-i18n'
import { agent_identity_api } from '@/api'
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
})

const emit = defineEmits(['close', 'saved'])

const { t } = useI18n()
const saving = ref(false)
const error = ref(null)
const form = reactive({
  name: '',
  type: 'api_key',
  username: '',
  tokenTtl: 300,
  secret: '',
})

watch(
  () => props.open,
  (open) => {
    if (!open) return
    error.value = null
    form.name = ''
    form.type = 'api_key'
    form.username = props.agentName
    form.tokenTtl = 300
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

  const body = { name: form.name.trim(), type: form.type }
  if (form.type === 'basic_auth') body.username = form.username.trim()
  if (form.type === 'client_credentials') body.tokenTtl = form.tokenTtl
  if (form.secret) body.secret = form.secret

  try {
    const res = await agent_identity_api.create(props.workspaceId, props.agentName, body)
    emit('saved', res.data)
  } catch (err) {
    error.value = errorText(err, t('agents_page.identity_failed_save'))
  } finally {
    saving.value = false
  }
}
</script>
