<template>
  <section class="rounded-xl border border-theme-border bg-white shadow-sm overflow-hidden">
    <div class="border-b border-theme-border px-5 py-4">
      <h2 class="text-lg font-semibold text-theme-text">{{ $t('agents_page.card') }}</h2>
      <p class="mt-1 text-sm text-theme-textLight">{{ $t('agents_page.card_desc') }}</p>
    </div>
    <div v-if="!agent.card" class="p-5">
      <div class="flex items-start gap-3 rounded-lg border border-dashed border-theme-border p-4">
        <span
          class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-theme-hover text-theme-textLight"
        >
          <svg
            class="h-4 w-4"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            stroke-width="2"
            aria-hidden="true"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              d="M12 6v6l4 2m6-2a10 10 0 11-20 0 10 10 0 0120 0z"
            />
          </svg>
        </span>
        <div class="min-w-0">
          <p class="text-sm font-medium text-theme-text">{{ $t('agents_page.card_pending') }}</p>
          <p class="mt-1 text-sm text-theme-textLight">{{ $t('agents_page.card_pending_desc') }}</p>
        </div>
      </div>
      <dl class="mt-4 space-y-4 text-sm">
        <div>
          <dt class="text-xs font-medium uppercase tracking-wider text-theme-textLight">
            {{ $t('agents_page.card_path') }}
          </dt>
          <dd class="mt-1.5">
            <code
              class="block break-all rounded-md border border-theme-border bg-theme-hover px-3 py-2 text-xs font-mono text-theme-text"
              >{{ agent.cardPath }}</code
            >
          </dd>
        </div>
        <div>
          <dt class="text-xs font-medium uppercase tracking-wider text-theme-textLight">
            {{ $t('agents_page.lb_strategy') }}
          </dt>
          <dd class="mt-1.5 text-theme-text">
            {{ $t('agents_page.lb_strategy_' + (agent.lbStrategy || 'round_robin')) }}
          </dd>
        </div>
      </dl>
    </div>
    <dl v-if="agent.card" class="space-y-4 px-5 py-4 text-sm">
      <div v-if="card.url">
        <dt class="text-theme-textLight">{{ $t('agents_page.gateway_url') }}</dt>
        <dd class="mt-1 flex items-start gap-2">
          <code
            class="min-w-0 flex-1 break-all rounded bg-theme-hover px-2 py-1 text-xs font-mono text-theme-text"
            >{{ card.url }}</code
          >
          <button
            type="button"
            class="shrink-0 text-xs font-medium text-primary-600 hover:text-primary-700"
            @click="copy(card.url, 'url')"
          >
            {{ copiedField === 'url' ? $t('agents_page.copied') : $t('agents_page.copy') }}
          </button>
        </dd>
      </div>
      <div v-if="card.provider">
        <dt class="text-theme-textLight">{{ $t('agents_page.provider') }}</dt>
        <dd class="mt-1 text-theme-text">{{ card.provider }}</dd>
      </div>
      <div>
        <dt class="text-theme-textLight">{{ $t('agents_page.input_modes') }}</dt>
        <dd class="mt-1 flex flex-wrap gap-1.5">
          <span
            v-for="mode in card.inputModes"
            :key="mode"
            class="rounded bg-theme-hover px-1.5 py-0.5 text-xs font-mono text-theme-text"
            >{{ mode }}</span
          >
          <span v-if="!card.inputModes.length" class="text-theme-textLight">—</span>
        </dd>
      </div>
      <div>
        <dt class="text-theme-textLight">{{ $t('agents_page.output_modes') }}</dt>
        <dd class="mt-1 flex flex-wrap gap-1.5">
          <span
            v-for="mode in card.outputModes"
            :key="mode"
            class="rounded bg-theme-hover px-1.5 py-0.5 text-xs font-mono text-theme-text"
            >{{ mode }}</span
          >
          <span v-if="!card.outputModes.length" class="text-theme-textLight">—</span>
        </dd>
      </div>
      <div>
        <dt class="text-theme-textLight">{{ $t('agents_page.capabilities') }}</dt>
        <dd class="mt-1 flex flex-wrap gap-1.5">
          <span
            v-for="capability in card.capabilities"
            :key="capability"
            class="rounded-full bg-primary-50 px-2 py-0.5 text-xs font-medium text-primary-700"
          >
            {{ capability }}
          </span>
          <span v-if="!card.capabilities.length" class="text-theme-textLight">—</span>
        </dd>
      </div>
      <div>
        <dt class="text-theme-textLight">{{ $t('agents_page.card_path') }}</dt>
        <dd class="mt-1">
          <code
            class="break-all rounded bg-theme-hover px-2 py-1 text-xs font-mono text-theme-text"
            >{{ agent.cardPath }}</code
          >
        </dd>
      </div>
      <div>
        <dt class="text-theme-textLight">{{ $t('agents_page.lb_strategy') }}</dt>
        <dd class="mt-1 text-theme-text">
          {{ $t('agents_page.lb_strategy_' + (agent.lbStrategy || 'round_robin')) }}
        </dd>
      </div>
      <div class="grid grid-cols-2 gap-4">
        <div>
          <dt class="text-theme-textLight">{{ $t('agents_page.updated') }}</dt>
          <dd class="mt-1 text-theme-text">{{ formatRelative(agent.updatedAt) }}</dd>
        </div>
        <div class="min-w-0">
          <dt class="text-theme-textLight">{{ $t('agents_page.checksum') }}</dt>
          <dd
            class="mt-1 truncate text-xs font-mono text-theme-textLight"
            :title="agent.cardChecksum"
          >
            {{ agent.cardChecksum }}
          </dd>
        </div>
      </div>
    </dl>
    <details v-if="agent.card" class="border-t border-theme-border">
      <summary
        class="cursor-pointer px-5 py-3 text-sm font-medium text-theme-text hover:bg-theme-hover"
      >
        {{ $t('agents_page.raw_card') }}
      </summary>
      <pre
        class="max-h-96 overflow-auto bg-theme-hover px-5 py-4 text-xs leading-5 font-mono text-theme-text"
        >{{ rawCard }}</pre>
    </details>
  </section>
</template>

<script setup>
import { computed } from 'vue'
import { useCopy } from '@/lib/clipboard'
import { parseCard } from '@/lib/agent'
import { formatRelative } from '@/utils/date'

const props = defineProps({
  agent: { type: Object, required: true },
})

const { copiedField, copy } = useCopy()

const card = computed(() => parseCard(props.agent.card))
const rawCard = computed(() => JSON.stringify(props.agent.card ?? {}, null, 2))
</script>
