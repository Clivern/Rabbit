<template>
  <span
    class="inline-flex shrink-0 items-center justify-center rounded-full bg-[var(--btn-primary-bg)] text-white"
    :class="small ? 'h-7 w-7' : 'h-10 w-10'"
    aria-hidden="true"
  >
    <svg :class="small ? 'h-3.5 w-3.5' : 'h-5 w-5'" viewBox="0 0 24 24" fill="currentColor">
      <path
        d="M12 2.5l2.2 6.3 6.3 2.2-6.3 2.2L12 19.5l-2.2-6.3L3.5 11l6.3-2.2zM19 16l.9 2.1 2.1.9-2.1.9L19 22l-.9-2.1-2.1-.9 2.1-.9z"
      />
    </svg>
  </span>
</template>

<script setup>
defineProps({
  small: { type: Boolean, default: false },
})
</script>
