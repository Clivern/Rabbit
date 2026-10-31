<template>
  <BaseModal
    :open="open"
    :title="title"
    :error="error"
    :busy="busy"
    size="sm"
    @close="emit('cancel')"
  >
    <p class="mt-2 text-sm text-theme-textLight">{{ message }}</p>
    <slot />
    <div class="mt-6 flex gap-3 justify-end">
      <button type="button" class="btn-secondary" :disabled="busy" @click="cancel">
        {{ $t('common.cancel') }}
      </button>
      <button
        type="button"
        class="btn-primary bg-red-600 hover:bg-red-700 focus:ring-red-500"
        :disabled="busy || disabled"
        @click="emit('confirm')"
      >
        {{ busy ? busyLabel : confirmLabel }}
      </button>
    </div>
  </BaseModal>
</template>

<script setup>
import BaseModal from '@/comp/BaseModal.vue'

const props = defineProps({
  open: { type: Boolean, default: false },
  title: { type: String, required: true },
  message: { type: String, default: '' },
  error: { type: String, default: null },
  busy: { type: Boolean, default: false },
  disabled: { type: Boolean, default: false },
  confirmLabel: { type: String, required: true },
  busyLabel: { type: String, required: true },
})

const emit = defineEmits(['confirm', 'cancel'])

function cancel() {
  if (!props.busy) emit('cancel')
}
</script>
