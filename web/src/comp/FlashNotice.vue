<template>
  <Teleport to="body">
    <Transition name="flash">
      <div
        v-if="message"
        class="fixed top-6 left-1/2 z-[200] -translate-x-1/2 pointer-events-none"
        role="status"
        aria-live="polite"
      >
        <div
          class="flex items-center gap-2 rounded-full border border-theme-border bg-white px-4 py-2 text-sm font-medium text-theme-text shadow-lg"
        >
          <svg
            class="h-4 w-4 shrink-0 text-emerald-600"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            aria-hidden="true"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M5 13l4 4L19 7"
            />
          </svg>
          <span>{{ message }}</span>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup>
import { watch, onUnmounted } from 'vue'

const props = defineProps({
  message: {
    type: String,
    default: null,
  },
  duration: {
    type: Number,
    default: 2500,
  },
})

const emit = defineEmits(['dismiss'])

let timer

watch(
  () => props.message,
  (value) => {
    clearTimeout(timer)
    if (!value) return
    timer = setTimeout(() => emit('dismiss'), props.duration)
  }
)

onUnmounted(() => clearTimeout(timer))
</script>

<style scoped>
.flash-enter-active,
.flash-leave-active {
  transition:
    opacity 0.2s ease,
    transform 0.2s ease;
}

.flash-enter-from,
.flash-leave-to {
  opacity: 0;
  transform: translateY(-0.5rem);
}
</style>
