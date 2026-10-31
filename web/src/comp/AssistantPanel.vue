<template>
  <div>
    <Transition
      enter-active-class="transition duration-200 ease-out motion-reduce:transition-none"
      enter-from-class="translate-y-4 scale-95 opacity-0"
      leave-active-class="transition duration-150 ease-in motion-reduce:transition-none"
      leave-to-class="translate-y-4 scale-95 opacity-0"
    >
      <section
        v-if="open"
        :aria-label="$t('assistant_panel.title')"
        class="fixed inset-0 z-40 flex origin-bottom-right flex-col overflow-hidden bg-white pb-[env(safe-area-inset-bottom)] sm:inset-x-auto sm:bottom-24 sm:left-auto sm:right-6 sm:top-auto sm:h-[min(720px,calc(100dvh-8rem))] sm:w-[480px] sm:rounded-3xl sm:pb-0 sm:shadow-[0_5px_40px_rgba(0,0,0,0.16)] lg:w-[560px]"
        @keydown.esc="open = false"
      >
        <template v-if="view === 'list'">
          <header
            class="flex items-center gap-2 border-b border-theme-border py-3 pl-4 pr-3 sm:gap-3 sm:py-4 sm:pl-5 sm:pr-4"
          >
            <AssistantAvatar />
            <div class="min-w-0 grow">
              <h2 class="text-base font-semibold text-theme-text">
                {{ $t('assistant_panel.title') }}
              </h2>
              <p class="mt-0.5 hidden text-sm text-theme-textLight sm:block">
                {{ $t('assistant_panel.subtitle') }}
              </p>
            </div>
            <button
              type="button"
              class="btn-primary inline-flex h-9 shrink-0 items-center gap-1.5 rounded-full px-3.5"
              :disabled="creating"
              @click="newChat"
            >
              <svg
                class="h-3.5 w-3.5"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                aria-hidden="true"
              >
                <path d="M12 5v14M5 12h14" />
              </svg>
              {{ $t('assistant_panel.new') }}
            </button>
            <button
              type="button"
              class="inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-lg text-theme-textLight hover:bg-theme-hover"
              :aria-label="$t('assistant_panel.close')"
              @click="open = false"
            >
              <svg
                class="h-[18px] w-[18px]"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                aria-hidden="true"
              >
                <path d="M6 6l12 12M18 6L6 18" />
              </svg>
            </button>
          </header>

          <div class="grow overflow-y-auto">
            <p
              v-if="errorMessage"
              class="m-4 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800"
            >
              {{ errorMessage }}
            </p>
            <p v-if="loading && !chats.length" class="p-6 text-center text-sm text-theme-textLight">
              {{ $t('assistant_panel.loading') }}
            </p>
            <p v-else-if="!chats.length" class="p-6 text-center text-sm text-theme-textLight">
              {{ $t('assistant_panel.no_chats') }}
            </p>
            <div
              v-for="item in chats"
              :key="item.id"
              class="group flex items-center border-b border-theme-border hover:bg-theme-hover"
            >
              <button
                type="button"
                class="flex min-w-0 grow flex-col gap-1 py-3.5 pl-4 pr-2 text-left sm:pl-5"
                @click="openChat(item.id)"
              >
                <span class="flex w-full items-center gap-2">
                  <span class="min-w-0 grow truncate text-sm font-semibold text-theme-text">{{
                    item.title || $t('assistant_panel.untitled')
                  }}</span>
                  <span v-if="item.status !== 'idle'" :class="statusPillClass(item.status)">{{
                    $t(`assistant_panel.status_${item.status}`)
                  }}</span>
                  <time
                    class="whitespace-nowrap text-xs text-theme-textLight"
                    :datetime="item.updatedAt"
                    :title="formatTimestamp(item.updatedAt)"
                    >{{ formatRelative(item.updatedAt) }}</time
                  >
                </span>
                <span class="w-full truncate text-sm text-theme-textLight">{{
                  item.preview ? plainText(item.preview) : $t('assistant_panel.no_messages')
                }}</span>
              </button>
              <button
                type="button"
                class="mr-3 inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-lg text-theme-textLight hover:bg-red-50 hover:text-red-700 focus-visible:opacity-100 sm:opacity-0 sm:group-hover:opacity-100"
                :aria-label="
                  $t('assistant_panel.delete_chat', {
                    title: item.title || $t('assistant_panel.untitled'),
                  })
                "
                @click="deleteTarget = item"
              >
                <svg
                  class="h-4 w-4"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  aria-hidden="true"
                >
                  <path
                    d="M4 7h16M10 11v6M14 11v6M6 7l1 12a2 2 0 0 0 2 2h6a2 2 0 0 0 2-2l1-12M9 7V4h6v3"
                  />
                </svg>
              </button>
            </div>
          </div>

          <footer
            class="flex items-center gap-2 border-t border-theme-border px-4 py-3 text-xs text-theme-textLight sm:px-5"
          >
            <svg
              class="h-3.5 w-3.5 shrink-0"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
              aria-hidden="true"
            >
              <path d="M12 3l8 3v6c0 4.5-3.4 8.3-8 9-4.6-.7-8-4.5-8-9V6z" />
            </svg>
            <span>{{ $t('assistant_panel.footer') }}</span>
          </footer>
        </template>

        <template v-else>
          <header
            class="flex items-center gap-1 border-b border-theme-border py-2.5 pl-1 pr-3 sm:gap-2 sm:py-3 sm:pl-2 sm:pr-4"
          >
            <button
              type="button"
              class="inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-lg text-theme-text hover:bg-theme-hover"
              :aria-label="$t('assistant_panel.back')"
              @click="showList"
            >
              <svg
                class="h-[18px] w-[18px]"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
                aria-hidden="true"
              >
                <path d="M15 18l-6-6 6-6" />
              </svg>
            </button>
            <AssistantAvatar small />
            <div class="flex min-w-0 grow items-center gap-2">
              <h2 class="truncate text-[15px] font-semibold text-theme-text">
                {{ active?.title || $t('assistant_panel.untitled') }}
              </h2>
              <span
                v-if="active && active.status !== 'idle'"
                :class="statusPillClass(active.status)"
                >{{ $t(`assistant_panel.status_${active.status}`) }}</span
              >
            </div>
            <button
              type="button"
              class="inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-lg text-theme-textLight hover:bg-red-50 hover:text-red-700"
              :aria-label="
                $t('assistant_panel.delete_chat', {
                  title: active?.title || $t('assistant_panel.untitled'),
                })
              "
              @click="deleteTarget = active"
            >
              <svg
                class="h-4 w-4"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
                aria-hidden="true"
              >
                <path
                  d="M4 7h16M10 11v6M14 11v6M6 7l1 12a2 2 0 0 0 2 2h6a2 2 0 0 0 2-2l1-12M9 7V4h6v3"
                />
              </svg>
            </button>
            <button
              type="button"
              class="inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-lg text-theme-textLight hover:bg-theme-hover"
              :aria-label="$t('assistant_panel.close')"
              @click="open = false"
            >
              <svg
                class="h-[18px] w-[18px]"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                aria-hidden="true"
              >
                <path d="M6 6l12 12M18 6L6 18" />
              </svg>
            </button>
          </header>

          <div
            ref="threadRef"
            class="flex grow flex-col gap-3 overflow-y-auto px-4 py-4 sm:px-5"
            aria-live="polite"
          >
            <div class="flex max-w-[88%] items-end gap-2 self-start">
              <AssistantAvatar small />
              <div
                v-if="greeting"
                class="whitespace-pre-wrap rounded-2xl rounded-bl-md bg-theme-hover px-4 py-2.5 text-sm leading-relaxed text-theme-text"
              >
                <span class="sr-only">{{ greetingText }}</span>
                <span aria-hidden="true">{{ greeting }}</span>
              </div>
              <div
                v-else
                class="flex items-center gap-1 rounded-2xl rounded-bl-md bg-theme-hover px-4 py-3.5"
                aria-hidden="true"
              >
                <span
                  v-for="dot in 3"
                  :key="dot"
                  class="h-1.5 w-1.5 animate-bounce rounded-full bg-primary-500 motion-reduce:animate-none"
                  :style="{ animationDelay: `${(dot - 1) * 150}ms` }"
                ></span>
              </div>
            </div>

            <Transition
              enter-active-class="transition duration-200 ease-out motion-reduce:transition-none"
              enter-from-class="translate-y-2 opacity-0"
            >
              <div
                v-if="active && !active.messages?.length && greeting === greetingText"
                class="flex flex-wrap justify-center gap-2"
              >
                <button
                  v-for="key in SUGGESTIONS"
                  :key="key"
                  type="button"
                  class="rounded-full border border-theme-border bg-white px-3.5 py-1.5 text-sm text-theme-text transition-colors hover:border-primary-300 hover:bg-theme-hover"
                  @click="useSuggestion(key)"
                >
                  {{ $t(`assistant_panel.${key}`) }}
                </button>
              </div>
            </Transition>

            <template v-for="message in active?.messages || []" :key="message.id">
              <div
                v-if="message.kind === 'user'"
                class="max-w-[80%] self-end whitespace-pre-wrap rounded-2xl rounded-br-md bg-[var(--btn-primary-bg)] px-4 py-2.5 text-sm leading-relaxed text-white"
              >
                {{ message.content }}
              </div>

              <div v-else class="flex max-w-[88%] items-end gap-2 self-start">
                <AssistantAvatar small />
                <!-- eslint-disable-next-line vue/no-v-html -- renderMarkdown escapes HTML and drops unsafe links -->
                <div
                  class="markdown-content assistant-markdown min-w-0 rounded-2xl rounded-bl-md bg-theme-hover px-4 py-2.5 text-sm text-theme-text"
                  v-html="renderMarkdown(message.content)"
                ></div>
              </div>
            </template>

            <template v-if="active?.status === 'running'">
              <div class="flex max-w-[88%] items-end gap-2 self-start">
                <AssistantAvatar small />
                <!-- eslint-disable-next-line vue/no-v-html -- renderMarkdown escapes HTML and drops unsafe links -->
                <div
                  v-if="streamed"
                  class="markdown-content assistant-markdown min-w-0 rounded-2xl rounded-bl-md bg-theme-hover px-4 py-2.5 text-sm text-theme-text"
                  v-html="renderMarkdown(streamed)"
                ></div>
                <div
                  v-else
                  class="flex items-center gap-1 rounded-2xl rounded-bl-md bg-theme-hover px-4 py-3.5"
                  role="status"
                >
                  <span class="sr-only">{{ $t('assistant_panel.thinking') }}</span>
                  <span
                    v-for="dot in 3"
                    :key="dot"
                    class="h-1.5 w-1.5 animate-bounce rounded-full bg-primary-500 motion-reduce:animate-none"
                    :style="{ animationDelay: `${(dot - 1) * 150}ms` }"
                  ></span>
                </div>
              </div>
            </template>
            <p
              v-if="active?.status === 'failed'"
              class="self-start rounded-lg border border-red-200 bg-red-50 px-3.5 py-2.5 text-[13px] text-red-800"
            >
              {{ $t('assistant_panel.failed_answer') }}
            </p>
            <p
              v-if="errorMessage"
              class="self-stretch rounded-lg border border-red-200 bg-red-50 px-3.5 py-2.5 text-[13px] text-red-800"
            >
              {{ errorMessage }}
            </p>
          </div>

          <form
            class="m-3 mt-0 flex items-end gap-2 rounded-2xl border border-theme-border bg-white p-1.5 pl-3 focus-within:border-primary-500"
            @submit.prevent="send"
          >
            <label for="assistant-input" class="sr-only">{{
              $t('assistant_panel.message_label')
            }}</label>
            <textarea
              id="assistant-input"
              ref="inputRef"
              v-model="draft"
              rows="2"
              required
              maxlength="4000"
              class="max-h-32 grow resize-none border-0 bg-transparent py-2 text-base leading-snug text-theme-text placeholder:text-primary-400 focus:outline-none sm:text-sm"
              :placeholder="$t('assistant_panel.placeholder')"
              @keydown.enter.exact.prevent="$event.target.form.requestSubmit()"
            ></textarea>
            <button
              type="submit"
              class="btn-primary inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-xl p-0"
              :disabled="sending"
              :aria-label="$t('assistant_panel.send')"
            >
              <svg
                class="h-[18px] w-[18px]"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
                aria-hidden="true"
              >
                <path d="M5 12h14M13 6l6 6-6 6" />
              </svg>
            </button>
          </form>
        </template>
      </section>
    </Transition>

    <ConfirmModal
      :open="!!deleteTarget"
      :title="$t('assistant_panel.delete_title')"
      :message="
        $t('assistant_panel.delete_confirm', {
          title: deleteTarget?.title || $t('assistant_panel.untitled'),
        })
      "
      :error="deleteError"
      :busy="deleting"
      :confirm-label="$t('assistant_panel.delete')"
      :busy-label="$t('assistant_panel.deleting')"
      @cancel="closeDelete"
      @confirm="deleteChat"
    />

    <button
      type="button"
      class="fixed bottom-[calc(1.5rem+env(safe-area-inset-bottom))] right-4 z-40 h-14 w-14 items-center justify-center rounded-full bg-[var(--btn-primary-bg)] text-white shadow-[0_4px_20px_rgba(0,0,0,0.2)] transition-transform duration-200 hover:scale-105 hover:bg-[var(--btn-primary-hover)] active:scale-95 motion-reduce:transition-none sm:right-6"
      :class="open ? 'hidden sm:inline-flex' : 'inline-flex'"
      :aria-label="open ? $t('assistant_panel.close') : $t('assistant_panel.open')"
      :aria-expanded="open"
      @click="open = !open"
    >
      <Transition
        mode="out-in"
        enter-active-class="transition duration-150 ease-out motion-reduce:transition-none"
        enter-from-class="rotate-[-90deg] scale-50 opacity-0"
        leave-active-class="transition duration-100 ease-in motion-reduce:transition-none"
        leave-to-class="rotate-90 scale-50 opacity-0"
      >
        <svg
          v-if="open"
          key="close"
          class="h-[22px] w-[22px]"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          aria-hidden="true"
        >
          <path d="M6 6l12 12M18 6L6 18" />
        </svg>
        <svg
          v-else
          key="open"
          class="h-6 w-6"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
          aria-hidden="true"
        >
          <path d="M21 12a8 8 0 0 1-11.6 7.1L4 20l1-4.6A8 8 0 1 1 21 12z" />
        </svg>
      </Transition>
    </button>
  </div>
</template>

<script setup>
import { nextTick, onBeforeUnmount, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { chat_api } from '@/api'
import AssistantAvatar from '@/comp/AssistantAvatar.vue'
import ConfirmModal from '@/comp/ConfirmModal.vue'
import { useWorkspaceContext } from '@/lib/permission'
import { errorText } from '@/lib/error'
import { playReply, playSend } from '@/lib/sound'
import { plainText, renderMarkdown } from '@/lib/markdown'
import { formatTimestamp, formatRelative } from '@/utils/date'

const SUGGESTIONS = ['suggestion_agents', 'suggestion_health', 'suggestion_workspace']
const GREETING_DOTS_MS = 400
const GREETING_LETTER_MS = 10
const POLL_MS = 1500

const { t } = useI18n()
const { currentUser, currentWorkspace } = useWorkspaceContext()

const open = ref(false)
const view = ref('list')
const chats = ref([])
const active = ref(null)
const draft = ref('')
const streamed = ref('')
const greeting = ref('')
const greetingText = ref('')
const loading = ref(false)
const creating = ref(false)
const sending = ref(false)
const errorMessage = ref(null)
const deleteTarget = ref(null)
const deleting = ref(false)
const deleteError = ref(null)
const threadRef = ref(null)
const inputRef = ref(null)

let pollTimer = null
let greetingTimer = null

async function loadChats() {
  loading.value = true
  errorMessage.value = null

  try {
    const res = await chat_api.list(currentWorkspace.id)
    chats.value = res.data.chats || []
  } catch (err) {
    errorMessage.value = errorText(err, t('assistant_panel.failed_load'))
  } finally {
    loading.value = false
  }
}

async function newChat() {
  creating.value = true
  errorMessage.value = null

  try {
    const res = await chat_api.create(currentWorkspace.id)
    showThread(res.data)
  } catch (err) {
    errorMessage.value = errorText(err, t('assistant_panel.failed_load'))
  } finally {
    creating.value = false
  }
}

async function openChat(id) {
  errorMessage.value = null

  try {
    const res = await chat_api.get(currentWorkspace.id, id)
    showThread(res.data)
  } catch (err) {
    errorMessage.value = errorText(err, t('assistant_panel.failed_load'))
  }
}

function showThread(chat) {
  active.value = chat
  view.value = 'thread'
  greet(!chat.messages?.length)
  followRunning()
  scrollToEnd()

  // On touch screens focusing would pop the keyboard over the messages; let the user tap the input.
  if (!window.matchMedia('(pointer: coarse)').matches) {
    nextTick(() => inputRef.value?.focus())
  }
}

function showList() {
  stopPolling()
  stopGreeting()
  active.value = null
  errorMessage.value = null
  view.value = 'list'
  loadChats()
}

function greet(animate) {
  stopGreeting()
  greetingText.value = currentUser?.name
    ? t('assistant_panel.greeting', { name: currentUser.name.split(' ')[0] })
    : t('assistant_panel.greeting_anonymous')

  if (!animate || window.matchMedia('(prefers-reduced-motion: reduce)').matches) {
    greeting.value = greetingText.value
    return
  }

  greeting.value = ''
  const letters = Array.from(greetingText.value)
  let typed = 0

  greetingTimer = setTimeout(function typeNext() {
    typed++
    greeting.value = letters.slice(0, typed).join('')
    if (typed < letters.length) greetingTimer = setTimeout(typeNext, GREETING_LETTER_MS)
  }, GREETING_DOTS_MS)
}

function stopGreeting() {
  clearTimeout(greetingTimer)
  greetingTimer = null
}

// Streams the answer into the thread as it is written; the server stores it once it is complete.
async function send() {
  // Called inside the click, so the browser lets the reply chime play later too.
  playSend()
  sending.value = true
  errorMessage.value = null
  streamed.value = ''

  try {
    await chat_api.send(currentWorkspace.id, active.value.id, { content: draft.value }, onEvent)
  } catch (err) {
    errorMessage.value = errorText(err, t('assistant_panel.failed_send'))
  } finally {
    sending.value = false
    streamed.value = ''
    // A stream cut short leaves the answer running on the server, so follow it until it is stored.
    followRunning()
  }
}

function onEvent(name, data) {
  // The user went back to the list or deleted the chat; the server still stores the answer.
  if (view.value !== 'thread') return

  if (name === 'chat') {
    active.value = data
    draft.value = ''
  } else if (name === 'delta') {
    if (!streamed.value) playReply()
    streamed.value += data.content
  } else {
    active.value = data.id ? data : { ...active.value, status: 'failed' }
  }

  scrollToEnd()
}

async function deleteChat() {
  deleting.value = true
  deleteError.value = null

  try {
    const id = deleteTarget.value.id
    await chat_api.delete(currentWorkspace.id, id)
    chats.value = chats.value.filter((item) => item.id !== id)
    closeDelete()
    if (active.value?.id === id) showList()
  } catch (err) {
    deleteError.value = errorText(err, t('assistant_panel.failed_delete'))
  } finally {
    deleting.value = false
  }
}

function closeDelete() {
  deleteTarget.value = null
  deleteError.value = null
}

function useSuggestion(key) {
  draft.value = t(`assistant_panel.${key}`)
  inputRef.value?.focus()
}

// While the assistant answers in the background, reload the thread until it is done.
function followRunning() {
  stopPolling()
  if (active.value?.status !== 'running') return

  pollTimer = setTimeout(async () => {
    const id = active.value?.id
    const res = await chat_api.get(currentWorkspace.id, id)
    if (active.value?.id !== id) return

    const grew = res.data.messages?.length !== active.value.messages?.length
    active.value = res.data
    if (grew) scrollToEnd()
    followRunning()
  }, POLL_MS)
}

function stopPolling() {
  clearTimeout(pollTimer)
  pollTimer = null
}

function scrollToEnd() {
  nextTick(() => {
    if (threadRef.value) threadRef.value.scrollTop = threadRef.value.scrollHeight
  })
}

function statusPillClass(status) {
  const base = 'inline-flex whitespace-nowrap rounded-full px-2 py-0.5 text-[11px] font-medium'
  return status === 'failed'
    ? `${base} bg-red-50 text-red-800`
    : `${base} bg-theme-hover text-primary-700`
}

watch(open, (isOpen) => {
  if (!isOpen) {
    stopPolling()
    return
  }

  if (view.value === 'list') loadChats()
  else followRunning()
})

onBeforeUnmount(() => {
  stopPolling()
  stopGreeting()
})
</script>
