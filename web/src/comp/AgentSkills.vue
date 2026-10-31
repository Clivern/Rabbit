<template>
  <section class="rounded-xl border border-theme-border bg-white shadow-sm overflow-hidden">
    <div class="border-b border-theme-border px-6 py-4">
      <h2 class="text-lg font-semibold text-theme-text">{{ $t('agents_page.skills') }}</h2>
      <p class="mt-1 text-sm text-theme-textLight">{{ $t('agents_page.skills_desc') }}</p>
    </div>
    <ul v-if="skills.length" class="divide-y divide-theme-border">
      <li v-for="skill in skills" :key="skill.id" class="px-6 py-4">
        <div class="flex flex-wrap items-center gap-2">
          <span v-if="skill.name" class="text-sm font-medium text-theme-text">{{
            skill.name
          }}</span>
          <span
            class="rounded bg-theme-hover px-1.5 py-0.5 text-xs font-mono text-theme-textLight"
            >{{ skill.id }}</span
          >
        </div>
        <p v-if="skill.description" class="mt-1 text-sm text-theme-textLight">
          {{ skill.description }}
        </p>
        <div v-if="skill.tags.length" class="mt-2 flex flex-wrap gap-1.5">
          <span
            v-for="tag in skill.tags"
            :key="tag"
            class="rounded-full border border-theme-border px-2 py-0.5 text-xs text-theme-textLight"
          >
            {{ tag }}
          </span>
        </div>
      </li>
    </ul>
    <p v-else class="p-8 text-center text-sm text-theme-textLight">
      {{ $t('agents_page.no_skills') }}
    </p>
  </section>
</template>

<script setup>
defineProps({
  skills: { type: Array, required: true },
})
</script>
