<template>
  <div class="min-h-screen flex items-center justify-center bg-theme-bg px-4">
    <div class="max-w-sm w-full">
      <div class="text-center mb-10">
        <router-link to="/" class="inline-flex justify-center text-theme-text">
          <AppLogo :alt="$t('common.logo_alt')" class="h-28 w-auto" />
        </router-link>
      </div>

      <div class="text-center mb-8">
        <h1 class="text-2xl font-semibold text-theme-text tracking-tight">
          {{ $t('login.title') }}
        </h1>
        <p class="mt-2 text-sm text-theme-textLight leading-relaxed">
          {{ $t('login.subtitle') }}
        </p>
      </div>

      <div v-if="error" class="rounded-md border border-red-200 bg-red-50 p-3 mb-5">
        <p class="text-sm text-red-800 text-center">
          {{ error }}
        </p>
      </div>

      <div class="space-y-3">
        <button
          type="button"
          class="btn-primary w-full py-3 flex items-center justify-center gap-2.5 disabled:opacity-50 disabled:cursor-not-allowed"
          :disabled="loading"
          @click="startOAuth('github')"
        >
          <svg
            v-if="loadingProvider !== 'github'"
            class="h-5 w-5 shrink-0"
            viewBox="0 0 24 24"
            fill="currentColor"
            aria-hidden="true"
          >
            <path
              fill-rule="evenodd"
              clip-rule="evenodd"
              d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z"
            />
          </svg>
          <svg
            v-else
            class="animate-spin h-5 w-5 shrink-0"
            fill="none"
            viewBox="0 0 24 24"
            aria-hidden="true"
          >
            <circle
              class="opacity-25"
              cx="12"
              cy="12"
              r="10"
              stroke="currentColor"
              stroke-width="4"
            />
            <path
              class="opacity-75"
              fill="currentColor"
              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
            />
          </svg>
          <span>{{
            loadingProvider === 'github' ? $t('login.submitting') : $t('login.continue_github')
          }}</span>
        </button>

        <button
          type="button"
          class="btn-secondary w-full py-3 flex items-center justify-center gap-2.5 disabled:opacity-50 disabled:cursor-not-allowed"
          :disabled="loading"
          @click="startOAuth('google')"
        >
          <svg
            v-if="loadingProvider !== 'google'"
            class="h-5 w-5 shrink-0"
            viewBox="0 0 24 24"
            aria-hidden="true"
          >
            <path
              fill="#4285F4"
              d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
            />
            <path
              fill="#34A853"
              d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
            />
            <path
              fill="#FBBC05"
              d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"
            />
            <path
              fill="#EA4335"
              d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
            />
          </svg>
          <svg
            v-else
            class="animate-spin h-5 w-5 shrink-0"
            fill="none"
            viewBox="0 0 24 24"
            aria-hidden="true"
          >
            <circle
              class="opacity-25"
              cx="12"
              cy="12"
              r="10"
              stroke="currentColor"
              stroke-width="4"
            />
            <path
              class="opacity-75"
              fill="currentColor"
              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
            />
          </svg>
          <span>{{
            loadingProvider === 'google' ? $t('login.submitting') : $t('login.continue_google')
          }}</span>
        </button>
      </div>

      <p class="text-center text-xs text-theme-textLight mt-10">
        {{ $t('common.copyright') }}
      </p>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { auth_api, setup_api } from '@/api'
import { saveUser } from '@/lib/auth'
import { applyUserPreferences } from '@/lib/preferences'
import AppLogo from '@/comp/AppLogo.vue'
import { errorText } from '@/lib/error'

const { t } = useI18n()
const router = useRouter()
const route = useRoute()

const loading = ref(false)
const loadingProvider = ref(null)
const error = ref(null)

function getRedirectTarget() {
  const redirect = route.query.redirect
  if (typeof redirect === 'string' && redirect.startsWith('/') && !redirect.startsWith('//')) {
    return redirect
  }
  return '/select-workspace'
}

const startOAuth = (provider) => {
  loading.value = true
  loadingProvider.value = provider

  const redirect = route.query.redirect
  if (typeof redirect === 'string' && redirect.startsWith('/') && !redirect.startsWith('//')) {
    window.location.href = `/api/v1/public/action/oauth/${provider}?redirect=${encodeURIComponent(redirect)}`
    return
  }

  window.location.href = `/api/v1/public/action/oauth/${provider}`
}

const finishOAuthLogin = async () => {
  loading.value = true
  error.value = null

  try {
    const response = await auth_api.getProfile()
    saveUser(response.data?.user)
    applyUserPreferences(response.data?.user)
    router.replace(getRedirectTarget())
  } catch (err) {
    error.value = errorText(err, t('login.oauth_failed'))
    loading.value = false
    loadingProvider.value = null
  }
}

onMounted(async () => {
  try {
    const response = await setup_api.checkInstalled()
    if (!response.data.installed) {
      router.push('/setup')
      return
    }
  } catch (err) {
    console.error('Failed to check setup status:', err)
  }

  if (route.query.oauth_error) {
    error.value = t('login.oauth_failed')
    return
  }

  if (route.query.oauth === 'github' || route.query.oauth === 'google') {
    await finishOAuthLogin()
  }
})
</script>
