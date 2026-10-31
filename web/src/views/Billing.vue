<template>
  <div class="min-h-screen bg-theme-bg">
    <AppNav />

    <main class="w-full px-4 py-8 sm:px-6 lg:px-8">
      <div class="mb-8">
        <h1 class="text-3xl font-semibold text-theme-text">{{ $t('billing_page.title') }}</h1>
        <p class="mt-2 text-theme-textLight">{{ $t('billing_page.subtitle') }}</p>
      </div>

      <div
        v-if="errorMessage"
        class="mb-6 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700"
      >
        {{ errorMessage }}
      </div>

      <div v-if="billingLoading" class="space-y-6">
        <div class="h-72 animate-pulse rounded-xl border border-theme-border bg-white" />
        <div class="h-96 animate-pulse rounded-xl border border-theme-border bg-white" />
      </div>

      <template v-else>
        <div class="grid grid-cols-1 gap-6 xl:grid-cols-3">
          <section
            class="overflow-hidden rounded-xl border border-theme-border bg-white shadow-sm xl:col-span-2"
          >
            <div class="border-b border-theme-border bg-primary-50/40 px-6 py-4">
              <p class="text-xs font-medium uppercase tracking-wider text-theme-textLight">
                {{ $t('billing_page.credits_label') }}
              </p>
              <h2 class="mt-1 text-2xl font-semibold text-theme-text">
                {{ formatCompact(tokenBalance) }}
              </h2>
              <p class="mt-1 text-sm text-theme-textLight">
                {{ $t('billing_page.credits_available') }}
              </p>
            </div>

            <div class="p-6">
              <p class="text-sm leading-6 text-theme-textLight">
                {{ $t('billing_page.credits_description', { rate: formatNumber(tokensPerUsd) }) }}
              </p>

              <div class="mt-6 grid grid-cols-2 gap-3 sm:grid-cols-4">
                <button
                  v-for="preset in amountPresets"
                  :key="preset"
                  type="button"
                  class="rounded-lg border px-4 py-3 text-sm font-medium transition-colors"
                  :class="
                    amountCents === preset * 100
                      ? 'border-primary-800 bg-primary-50 text-theme-text'
                      : 'border-theme-border bg-white text-theme-text hover:bg-theme-hover'
                  "
                  @click="selectPreset(preset)"
                >
                  ${{ preset }}
                </button>
              </div>

              <label class="mt-5 block">
                <span class="form-label">{{ $t('billing_page.custom_amount') }}</span>
                <input
                  v-model="customAmount"
                  type="number"
                  min="1"
                  step="1"
                  class="input-field mt-1"
                  :placeholder="$t('billing_page.custom_amount_placeholder')"
                  @input="onCustomAmount"
                />
              </label>

              <p class="mt-3 text-sm text-theme-textLight">
                {{
                  $t('billing_page.tokens_you_get', { tokens: formatCompact(tokensForSelection) })
                }}
              </p>

              <div class="mt-8 flex flex-wrap gap-3">
                <button
                  type="button"
                  class="btn-primary"
                  :disabled="checkoutLoading || tokensForSelection <= 0"
                  @click="buyCredits"
                >
                  {{ checkoutLoading ? $t('billing_page.opening') : $t('billing_page.buy_tokens') }}
                </button>
                <button
                  v-if="canManageStripe"
                  type="button"
                  class="btn-secondary"
                  :disabled="portalLoading"
                  @click="openPortal"
                >
                  {{
                    portalLoading ? $t('billing_page.opening') : $t('billing_page.manage_billing')
                  }}
                </button>
              </div>
            </div>
          </section>

          <aside class="space-y-6">
            <section class="rounded-xl border border-theme-border bg-white p-6 shadow-sm">
              <h2 class="text-lg font-semibold text-theme-text mb-4">
                {{ $t('billing_page.payment') }}
              </h2>
              <p class="text-sm leading-6 text-theme-textLight">
                <template v-if="canManageStripe">
                  {{ $t('billing_page.payment_manage_desc') }}
                </template>
                <template v-else>
                  {{ $t('billing_page.payment_buy_desc') }}
                </template>
              </p>
              <button
                type="button"
                class="btn-secondary mt-4 w-full"
                :class="{ 'cursor-not-allowed opacity-60': !canManageStripe || portalLoading }"
                :disabled="!canManageStripe || portalLoading"
                @click="openPortal"
              >
                {{
                  canManageStripe
                    ? portalLoading
                      ? $t('billing_page.opening')
                      : $t('billing_page.open_billing_portal')
                    : $t('billing_page.available_after_purchase')
                }}
              </button>
            </section>
          </aside>
        </div>

        <section
          class="mt-6 overflow-hidden rounded-xl border border-theme-border bg-white shadow-sm"
        >
          <div class="border-b border-theme-border bg-primary-50/40 px-6 py-4">
            <div class="flex flex-wrap items-start justify-between gap-3">
              <div>
                <h2 class="text-lg font-semibold text-theme-text">
                  {{ $t('billing_page.usage') }}
                </h2>
                <p class="mt-1 text-sm text-theme-textLight">
                  {{ $t('billing_page.usage_subtitle') }}
                </p>
              </div>
              <span
                class="rounded-full bg-theme-hover px-3 py-1 text-xs font-medium text-theme-textLight"
              >
                {{ $t('billing_page.token_balance_label') }}
              </span>
            </div>
          </div>

          <div class="space-y-6 px-6 py-5">
            <div v-for="metric in usageMetrics" :key="metric.id">
              <div class="flex flex-wrap items-start justify-between gap-3">
                <div>
                  <div class="flex flex-wrap items-center gap-2">
                    <p class="text-sm font-medium text-theme-text">{{ metric.label }}</p>
                    <span
                      class="rounded px-1.5 py-0.5 text-[10px] font-medium uppercase tracking-wide"
                      :class="metric.metricBadgeClass"
                    >
                      {{ metric.badgeShort }}
                    </span>
                  </div>
                  <p v-if="metric.hint" class="mt-0.5 text-xs text-theme-textLight">
                    {{ metric.hint }}
                  </p>
                </div>
                <p class="text-sm font-medium tabular-nums text-theme-text">
                  <span>{{ metric.displayUsed }}</span>
                  <span v-if="!metric.infoOnly" class="font-normal text-theme-textLight">
                    / {{ metric.displayLimit }}</span
                  >
                </p>
              </div>

              <div
                v-if="!metric.infoOnly"
                class="mt-3 h-2 overflow-hidden rounded-full bg-theme-hover"
              >
                <div
                  class="h-full rounded-full transition-all duration-300"
                  :style="metric.barStyle"
                />
              </div>

              <p v-if="metric.infoOnly" class="mt-2 text-xs text-theme-textLight">
                {{ metric.infoSuffix }}
              </p>
              <p v-else class="mt-2 text-xs text-theme-textLight">
                {{ metric.percentLabel }} {{ metric.percentSuffix }}
              </p>
            </div>
          </div>

          <div class="border-t border-theme-border bg-theme-hover/30 px-6 py-4">
            <p class="text-xs leading-5 text-theme-textLight">
              <span class="font-medium text-theme-text">{{
                $t('billing_page.usage_footer_monthly')
              }}</span>
              {{ $t('billing_page.usage_reset_on', { date: usagePeriodReset }) }}
              <span class="font-medium text-theme-text">{{
                $t('billing_page.usage_footer_ongoing')
              }}</span>
              {{ $t('billing_page.usage_ongoing_note') }}
            </p>
          </div>
        </section>
      </template>
    </main>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppNav from '@/comp/AppNav.vue'
import { billing_api } from '@/api'
import { useWorkspaceContext } from '@/lib/permission'
import { errorText } from '@/lib/error'

const { t, locale } = useI18n()
const { currentWorkspace } = useWorkspaceContext()

const billingLoading = ref(false)
const portalLoading = ref(false)
const checkoutLoading = ref(false)
const errorMessage = ref('')
const amountCents = ref(1000)
const customAmount = ref('')
const amountPresets = [10, 25, 50, 100]
const billingStatus = ref({
  providerCustomerId: '',
  aiTokensBalance: 0,
  tokensPerUsd: 20000,
  updatedAt: '',
})

const usageData = ref({
  used: {
    workspaceMembers: 0,
    aiTokens: 0,
    aiCost: 0,
  },
  limits: {
    workspaceMembers: 3,
    aiTokens: 500_000,
  },
  periodReset: '',
})

const tokenBalance = computed(() => billingStatus.value.aiTokensBalance || 0)
const tokensPerUsd = computed(() => billingStatus.value.tokensPerUsd || 20000)
const tokensForSelection = computed(() =>
  Math.floor((amountCents.value * tokensPerUsd.value) / 100)
)

const canManageStripe = computed(() => !!billingStatus.value.providerCustomerId)

function selectPreset(dollars) {
  amountCents.value = dollars * 100
  customAmount.value = ''
}

function onCustomAmount() {
  const dollars = Number(customAmount.value)
  amountCents.value = dollars >= 1 ? Math.round(dollars * 100) : 0
}

const usagePeriodReset = computed(() => {
  if (usageData.value.periodReset) {
    return formatDate(usageData.value.periodReset)
  }

  const now = new Date()
  const next = new Date(now.getFullYear(), now.getMonth() + 1, 1)
  return next.toLocaleDateString(locale.value, { month: 'short', day: 'numeric', year: 'numeric' })
})

const formatNumber = (value) => value.toLocaleString(locale.value, { maximumFractionDigits: 0 })

const formatCompact = (value) => {
  if (value >= 1_000_000) return `${(value / 1_000_000).toFixed(1).replace(/\.0$/, '')}M`
  if (value >= 1_000) return `${(value / 1_000).toFixed(1).replace(/\.0$/, '')}K`
  return formatNumber(value)
}

const formatUsd = (value) => {
  const fractionDigits = value > 0 && value < 0.01 ? 6 : 2
  return new Intl.NumberFormat(locale.value, {
    style: 'currency',
    currency: 'USD',
    minimumFractionDigits: fractionDigits,
    maximumFractionDigits: fractionDigits,
  }).format(value)
}

const usagePercent = (used, limit) => {
  if (!limit) return 0
  return Math.min(100, Math.round((used / limit) * 100))
}

const usageBarColor = (percent) => {
  const p = Math.min(100, Math.max(0, percent)) / 100
  return `color-mix(in srgb, var(--primary-400) ${Math.round((1 - p) * 100)}%, var(--theme-text))`
}

const usageBarStyle = (percent) => ({
  width: `${Math.min(100, Math.max(0, percent))}%`,
  backgroundColor: usageBarColor(percent),
})

const buildUsageMetrics = (limits, used) => {
  const definitions = [
    {
      id: 'aiTokens',
      resetPolicy: 'monthly',
      label: t('billing_page.metrics.ai_tokens_label'),
      hint: t('billing_page.metrics.ai_tokens_hint'),
      used: used.aiTokens,
      limit: limits.aiTokens,
      displayUsed: formatCompact(used.aiTokens),
      displayLimit: formatCompact(limits.aiTokens),
    },
    {
      id: 'workspaceMembers',
      resetPolicy: 'capacity',
      label: t('billing_page.metrics.workspace_members_label'),
      hint: t('billing_page.metrics.workspace_members_hint'),
      used: used.workspaceMembers,
      limit: limits.workspaceMembers,
      displayUsed: formatNumber(used.workspaceMembers),
      displayLimit: formatNumber(limits.workspaceMembers),
    },
    {
      id: 'aiCost',
      resetPolicy: 'monthly',
      infoOnly: true,
      label: t('billing_page.metrics.ai_cost_label'),
      hint: t('billing_page.metrics.ai_cost_hint'),
      used: used.aiCost,
      limit: 1,
      displayUsed: formatUsd(used.aiCost),
      displayLimit: '',
      infoSuffix: t('billing_page.metric_suffix_ai_cost'),
    },
  ]

  const displayByPolicy = {
    monthly: {
      badgeShort: t('billing_page.metric_badge_monthly'),
      metricBadgeClass: 'bg-primary-50 text-theme-textLight',
      percentSuffix: t('billing_page.metric_suffix_monthly'),
    },
    capacity: {
      badgeShort: t('billing_page.metric_badge_ongoing'),
      metricBadgeClass: 'bg-theme-hover text-theme-textLight',
      percentSuffix: t('billing_page.metric_suffix_capacity'),
    },
  }

  return definitions.map((metric) => {
    const percent = usagePercent(metric.used, metric.limit)
    const display = displayByPolicy[metric.resetPolicy]
    return {
      ...metric,
      ...display,
      percent,
      percentLabel: `${percent}%`,
      barStyle: usageBarStyle(percent),
    }
  })
}

const usageMetrics = computed(() => {
  return buildUsageMetrics(usageData.value.limits, usageData.value.used)
})

const formatDate = (value) => {
  if (!value) return ''
  return new Date(value).toLocaleDateString(locale.value, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  })
}

const loadBilling = async () => {
  billingLoading.value = true
  errorMessage.value = ''
  try {
    const [statusRes, usageRes] = await Promise.all([
      billing_api.status(currentWorkspace.id),
      billing_api.usage(currentWorkspace.id),
    ])
    billingStatus.value = statusRes.data
    usageData.value = usageRes.data
  } catch (err) {
    errorMessage.value = errorText(err, t('billing_page.failed_load'))
  } finally {
    billingLoading.value = false
  }
}

const buyCredits = async () => {
  checkoutLoading.value = true
  errorMessage.value = ''
  try {
    const res = await billing_api.checkout(currentWorkspace.id, { amountCents: amountCents.value })
    window.location.href = res.data.url
  } catch (err) {
    errorMessage.value = errorText(err, t('billing_page.failed_checkout'))
  } finally {
    checkoutLoading.value = false
  }
}

const openPortal = async () => {
  if (!canManageStripe.value) return

  portalLoading.value = true
  errorMessage.value = ''
  try {
    const res = await billing_api.portal(currentWorkspace.id)
    window.location.href = res.data.url
  } catch (err) {
    errorMessage.value = errorText(err, t('billing_page.failed_portal'))
  } finally {
    portalLoading.value = false
  }
}

onMounted(loadBilling)
</script>
