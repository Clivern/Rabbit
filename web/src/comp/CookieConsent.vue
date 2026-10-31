<template>
  <Teleport to="body">
    <!-- Banner -->
    <Transition name="cookie-banner">
      <div
        v-if="showCookieBanner"
        class="fixed inset-x-0 bottom-0 z-[150] p-4 sm:p-6 pointer-events-none"
        role="dialog"
        aria-labelledby="cookie-banner-title"
        aria-describedby="cookie-banner-desc"
      >
        <div
          class="cookie-consent-panel pointer-events-auto mx-auto flex max-w-4xl flex-col gap-4 rounded-lg border border-theme-border p-5 shadow-lg sm:flex-row sm:items-center sm:gap-6 sm:p-6"
        >
          <div class="min-w-0 flex-1">
            <p id="cookie-banner-title" class="text-sm font-semibold text-theme-text">
              {{ $t('cookies.banner_title') }}
            </p>
            <p id="cookie-banner-desc" class="mt-1.5 text-sm leading-relaxed text-theme-textLight">
              {{ $t('cookies.banner_desc') }}
            </p>
          </div>
          <div class="flex shrink-0 flex-col gap-2 sm:flex-row sm:items-center">
            <button
              type="button"
              class="btn-secondary whitespace-nowrap px-4 py-2"
              @click="onCustomize"
            >
              {{ $t('cookies.customize') }}
            </button>
            <button
              type="button"
              class="btn-secondary whitespace-nowrap px-4 py-2"
              @click="onReject"
            >
              {{ $t('cookies.reject_optional') }}
            </button>
            <button
              type="button"
              class="btn-primary whitespace-nowrap px-4 py-2"
              @click="onAcceptAll"
            >
              {{ $t('cookies.accept_all') }}
            </button>
          </div>
        </div>
      </div>
    </Transition>

    <!-- Preferences modal -->
    <Transition name="modal">
      <div
        v-if="showCookiePreferences"
        class="fixed inset-0 z-[160] flex items-end justify-center p-4 sm:items-center sm:p-6"
        role="dialog"
        aria-labelledby="cookie-prefs-title"
        aria-modal="true"
      >
        <div
          class="absolute inset-0 bg-primary-900/40 backdrop-blur-[2px]"
          aria-hidden="true"
          @click="closePreferences"
        />
        <div
          class="cookie-consent-panel relative w-full max-w-lg rounded-lg border border-theme-border p-6 shadow-xl"
        >
          <div class="flex items-start justify-between gap-4">
            <div>
              <h2 id="cookie-prefs-title" class="text-lg font-semibold text-theme-text">
                {{ $t('cookies.prefs_title') }}
              </h2>
              <p class="mt-1 text-sm text-theme-textLight">
                {{ $t('cookies.prefs_desc') }}
              </p>
            </div>
            <button
              type="button"
              class="rounded-md p-1 text-theme-textLight transition hover:bg-theme-hover hover:text-theme-text"
              :aria-label="$t('common.close')"
              @click="closePreferences"
            >
              <svg
                class="h-5 w-5"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                aria-hidden="true"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M6 18L18 6M6 6l12 12"
                />
              </svg>
            </button>
          </div>

          <ul class="mt-6 space-y-4">
            <li class="rounded-lg border border-theme-border p-4">
              <div class="flex items-start justify-between gap-4">
                <div>
                  <p class="text-sm font-medium text-theme-text">
                    {{ $t('cookies.essential_title') }}
                  </p>
                  <p class="mt-1 text-sm text-theme-textLight">
                    {{ $t('cookies.essential_desc') }}
                  </p>
                </div>
                <span
                  class="shrink-0 rounded-full bg-theme-hover px-2.5 py-0.5 text-xs font-medium text-theme-textLight"
                >
                  {{ $t('cookies.always_on') }}
                </span>
              </div>
            </li>

            <li class="rounded-lg border border-theme-border p-4">
              <div class="flex items-start justify-between gap-4">
                <div>
                  <p class="text-sm font-medium text-theme-text">
                    {{ $t('cookies.preferences_title') }}
                  </p>
                  <p class="mt-1 text-sm text-theme-textLight">
                    {{ $t('cookies.preferences_desc') }}
                  </p>
                </div>
                <label class="relative inline-flex shrink-0 cursor-pointer items-center">
                  <input v-model="draft.preferences" type="checkbox" class="peer sr-only" />
                  <span
                    class="relative inline-block h-6 w-11 rounded-full bg-theme-border transition peer-checked:bg-primary-800 after:absolute after:left-0.5 after:top-0.5 after:h-5 after:w-5 after:rounded-full after:bg-white after:shadow-sm after:transition after:content-[''] peer-checked:after:translate-x-5"
                    aria-hidden="true"
                  />
                </label>
              </div>
            </li>

            <li class="rounded-lg border border-theme-border p-4">
              <div class="flex items-start justify-between gap-4">
                <div>
                  <p class="text-sm font-medium text-theme-text">
                    {{ $t('cookies.analytics_title') }}
                  </p>
                  <p class="mt-1 text-sm text-theme-textLight">
                    {{ $t('cookies.analytics_desc') }}
                  </p>
                </div>
                <label class="relative inline-flex shrink-0 cursor-pointer items-center">
                  <input v-model="draft.analytics" type="checkbox" class="peer sr-only" />
                  <span
                    class="relative inline-block h-6 w-11 rounded-full bg-theme-border transition peer-checked:bg-primary-800 after:absolute after:left-0.5 after:top-0.5 after:h-5 after:w-5 after:rounded-full after:bg-white after:shadow-sm after:transition after:content-[''] peer-checked:after:translate-x-5"
                    aria-hidden="true"
                  />
                </label>
              </div>
            </li>
          </ul>

          <div class="mt-6 flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
            <button type="button" class="btn-secondary px-4 py-2" @click="closePreferences">
              {{ $t('common.cancel') }}
            </button>
            <button type="button" class="btn-primary px-4 py-2" @click="onSavePreferences">
              {{ $t('cookies.save_preferences') }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup>
import { reactive, watch } from 'vue'
import {
  acceptAllCookies,
  consentCategoriesForEdit,
  hasConsent,
  initCookieConsent,
  rejectOptionalCookies,
  saveConsent,
  showCookieBanner,
  showCookiePreferences,
} from '@/lib/cookies'

const draft = reactive({
  preferences: false,
  analytics: false,
})

watch(showCookiePreferences, (open) => {
  if (!open) return
  const categories = consentCategoriesForEdit()
  draft.preferences = categories.preferences
  draft.analytics = categories.analytics
})

function onAcceptAll() {
  acceptAllCookies()
}

function onReject() {
  rejectOptionalCookies()
}

function onCustomize() {
  showCookieBanner.value = false
  showCookiePreferences.value = true
}

function closePreferences() {
  showCookiePreferences.value = false
  if (!hasConsent()) {
    showCookieBanner.value = true
  }
}

function onSavePreferences() {
  saveConsent({
    preferences: draft.preferences,
    analytics: draft.analytics,
  })
}

initCookieConsent()
</script>

<style scoped>
.cookie-consent-panel {
  background-color: var(--surface);
}

.cookie-banner-enter-active,
.cookie-banner-leave-active {
  transition:
    opacity 0.25s ease,
    transform 0.25s ease;
}

.cookie-banner-enter-from,
.cookie-banner-leave-to {
  opacity: 0;
  transform: translateY(1rem);
}
</style>
