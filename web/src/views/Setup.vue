<template>
  <div class="min-h-screen flex items-center justify-center bg-theme-bg px-4 py-12">
    <div class="max-w-md w-full">
      <div class="text-center mb-10">
        <div class="flex justify-center mb-6">
          <AppLogo :alt="$t('common.logo_alt')" class="h-24 w-auto text-theme-text" />
        </div>
        <h1 class="text-2xl font-semibold text-theme-text mb-2">{{ $t('setup.title') }}</h1>
        <p class="text-sm text-theme-textLight">{{ $t('setup.subtitle') }}</p>
      </div>

      <div class="bg-white rounded-lg border border-theme-border p-8 shadow-sm">
        <form class="space-y-5" @submit.prevent="handleSetup">
          <div>
            <label for="platform-email" class="block text-sm font-medium text-theme-text mb-2">
              {{ $t('setup.platform_email') }}
            </label>
            <input
              id="platform-email"
              v-model="form.platformEmail"
              type="email"
              required
              class="input-field"
              :placeholder="$t('setup.platform_email_placeholder')"
              :disabled="loading"
            />
            <p class="text-xs text-theme-textLight mt-1.5">{{ $t('setup.platform_email_hint') }}</p>
          </div>

          <div v-if="error" class="rounded-md border border-red-200 bg-red-50 p-3">
            <p class="text-sm text-red-800">
              {{ error }}
            </p>
          </div>

          <div>
            <button
              type="submit"
              class="w-full btn-primary py-2.5 disabled:opacity-50 disabled:cursor-not-allowed"
              :disabled="loading"
            >
              <span v-if="!loading">{{ $t('setup.submit') }}</span>
              <span v-else class="flex items-center justify-center">
                <svg
                  class="animate-spin -ml-1 mr-3 h-5 w-5 text-white"
                  fill="none"
                  viewBox="0 0 24 24"
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
                {{ $t('setup.submitting') }}
              </span>
            </button>
          </div>
        </form>
      </div>

      <p class="text-center text-xs text-theme-textLight mt-8">
        {{ $t('common.copyright') }}
      </p>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { setup_api } from '@/api'
import { showFlash } from '@/lib/flash'
import AppLogo from '@/comp/AppLogo.vue'
import { errorText } from '@/lib/error'

const { t } = useI18n()
const router = useRouter()
const form = reactive({
  platformEmail: '',
})

const loading = ref(false)
const error = ref(null)

onMounted(async () => {
  try {
    const response = await setup_api.checkInstalled()
    if (response.data.installed) {
      router.push('/login')
    }
  } catch (err) {
    console.error('Failed to check setup status:', err)
  }
})

const handleSetup = async () => {
  loading.value = true
  error.value = null

  try {
    await setup_api.install({
      platformEmail: form.platformEmail,
    })

    showFlash(t('common.saved'))

    setTimeout(() => {
      router.push('/login')
    }, 2000)
  } catch (err) {
    error.value = errorText(err, t('setup.error_default'))
  } finally {
    loading.value = false
  }
}
</script>
