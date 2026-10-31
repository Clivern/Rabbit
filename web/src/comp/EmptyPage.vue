<template>
  <div class="min-h-screen bg-theme-bg">
    <AppNav />

    <main class="w-full px-4 sm:px-6 lg:px-8 py-8">
      <header class="mb-8">
        <h1 class="text-2xl sm:text-3xl font-semibold text-theme-text tracking-tight">
          {{ title }}
        </h1>
      </header>

      <section class="rounded-xl border border-theme-border bg-white p-8 shadow-sm">
        <p class="text-sm text-theme-textLight">This page is empty for now.</p>
      </section>
    </main>
  </div>
</template>

<script setup>
import AppNav from '@/comp/AppNav.vue'

defineProps({
  title: {
    type: String,
    required: true,
  },
})
</script>
