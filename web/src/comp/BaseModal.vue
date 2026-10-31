<template>
  <Teleport to="body">
    <Transition name="modal">
      <div
        v-if="open"
        class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40"
        role="dialog"
        aria-modal="true"
        :aria-labelledby="titleId"
        @click.self="close"
      >
        <div
          ref="panel"
          class="bg-white rounded-lg shadow-xl border border-theme-border w-full p-6 max-h-[90vh] overflow-y-auto"
          :class="size === 'sm' ? 'max-w-sm' : 'max-w-lg'"
          @keydown.tab="trapFocus"
        >
          <h2 :id="titleId" class="text-lg font-semibold text-theme-text">{{ title }}</h2>
          <p v-if="description" class="mt-1 text-sm text-theme-textLight">{{ description }}</p>

          <div
            v-if="error"
            class="mt-4 rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-800"
            role="alert"
          >
            {{ error }}
          </div>

          <slot />
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup>
import { nextTick, onBeforeUnmount, ref, useId, watch } from 'vue'

const props = defineProps({
  open: { type: Boolean, default: false },
  title: { type: String, required: true },
  description: { type: String, default: '' },
  error: { type: String, default: null },
  busy: { type: Boolean, default: false },
  size: { type: String, default: 'lg' },
})

const emit = defineEmits(['close'])

const FOCUSABLE =
  'a[href], button:not([disabled]), input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])'

const titleId = useId()
const panel = ref(null)
let opener = null

function close() {
  if (!props.busy) emit('close')
}

function onKeydown(e) {
  if (e.key === 'Escape') close()
}

// Keep Tab and Shift+Tab cycling inside the panel while the modal is open.
function trapFocus(e) {
  const items = [...panel.value.querySelectorAll(FOCUSABLE)]
  const first = items[0]
  const last = items[items.length - 1]

  if (e.shiftKey && document.activeElement === first) {
    e.preventDefault()
    last.focus()
  } else if (!e.shiftKey && document.activeElement === last) {
    e.preventDefault()
    first.focus()
  }
}

function release() {
  document.body.style.overflow = ''
  document.removeEventListener('keydown', onKeydown)
  opener?.focus()
  opener = null
}

// Lock page scroll, listen for Escape and move focus in on open; undo all of it and hand focus back on close.
watch(
  () => props.open,
  async (open) => {
    if (!open) {
      if (opener) release()
      return
    }

    opener = document.activeElement
    document.body.style.overflow = 'hidden'
    document.addEventListener('keydown', onKeydown)

    await nextTick()
    panel.value.querySelector(FOCUSABLE)?.focus()
  },
  { immediate: true }
)

onBeforeUnmount(() => {
  if (opener) release()
})
</script>
