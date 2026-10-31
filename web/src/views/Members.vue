<template>
  <div class="min-h-screen bg-theme-bg">
    <AppNav />

    <main class="w-full py-8 px-4 sm:px-6 lg:px-8">
      <div class="mb-8">
        <h1 class="text-3xl font-semibold text-theme-text">{{ $t('members_page.title') }}</h1>
        <p class="text-theme-textLight mt-2">{{ $t('members_page.subtitle') }}</p>
      </div>

      <div
        v-if="errorMessage"
        class="mb-6 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800"
      >
        {{ errorMessage }}
      </div>

      <div class="space-y-8">
        <section class="rounded-xl border border-theme-border bg-white shadow-sm overflow-hidden">
          <div class="border-b border-theme-border px-6 py-4">
            <h2 class="text-lg font-semibold text-theme-text">
              {{ $t('members_page.invite_member') }}
            </h2>
            <p class="mt-1 text-sm text-theme-textLight">{{ $t('members_page.invite_desc') }}</p>
          </div>
          <form class="grid gap-4 p-6 md:grid-cols-[1fr_180px_auto]" @submit.prevent="createInvite">
            <div>
              <label for="invite-email" class="form-label">{{ $t('members_page.email') }}</label>
              <input
                id="invite-email"
                v-model="inviteForm.email"
                type="email"
                required
                class="input-field"
                :placeholder="$t('members_page.email_placeholder')"
                :disabled="inviteLoading"
              />
            </div>
            <div>
              <label for="invite-role" class="form-label">{{ $t('members_page.role') }}</label>
              <select
                id="invite-role"
                v-model="inviteForm.role"
                class="input-field"
                :disabled="inviteLoading"
              >
                <option value="regular">{{ $t('members_page.role_regular') }}</option>
                <option value="admin">{{ $t('members_page.role_admin') }}</option>
                <option value="readonly">{{ $t('members_page.role_readonly') }}</option>
              </select>
            </div>
            <div class="flex items-end">
              <button
                type="submit"
                class="btn-primary w-full md:w-auto disabled:opacity-50 disabled:cursor-not-allowed"
                :disabled="inviteLoading"
              >
                {{ inviteLoading ? $t('members_page.sending') : $t('members_page.send_invite') }}
              </button>
            </div>
          </form>
        </section>

        <section class="rounded-xl border border-theme-border bg-white shadow-sm overflow-hidden">
          <div class="border-b border-theme-border px-6 py-4">
            <h2 class="text-lg font-semibold text-theme-text">{{ $t('members_page.members') }}</h2>
          </div>
          <div v-if="loadingMembers" class="p-12 text-center">
            <svg
              class="animate-spin h-8 w-8 mx-auto text-theme-text"
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
            <p class="text-theme-textLight mt-3">{{ $t('members_page.loading_members') }}</p>
          </div>
          <div
            v-else-if="members.length === 0"
            class="p-12 text-center text-sm text-theme-textLight"
          >
            {{ $t('members_page.no_members') }}
          </div>
          <div v-else class="overflow-x-auto">
            <table class="min-w-full divide-y divide-theme-border">
              <thead class="bg-theme-hover">
                <tr>
                  <th
                    class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider"
                  >
                    {{ $t('members_page.name') }}
                  </th>
                  <th
                    class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider"
                  >
                    {{ $t('members_page.email') }}
                  </th>
                  <th
                    class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider"
                  >
                    {{ $t('members_page.role') }}
                  </th>
                  <th
                    class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider"
                  >
                    {{ $t('members_page.joined') }}
                  </th>
                  <th
                    class="px-6 py-3 text-right text-xs font-medium text-theme-textLight uppercase tracking-wider"
                  >
                    {{ $t('members_page.actions') }}
                  </th>
                </tr>
              </thead>
              <tbody class="bg-white divide-y divide-theme-border">
                <tr v-for="member in members" :key="member.id" class="hover:bg-theme-hover">
                  <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-theme-text">
                    {{ member.name }}
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap text-sm text-theme-textLight">
                    {{ member.email }}
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap text-sm">
                    <select
                      v-if="canEditMember(member)"
                      :value="member.role"
                      class="input-field py-1.5 text-sm"
                      :disabled="updatingMemberId === member.userId"
                      @change="updateMemberRole(member, $event.target.value)"
                    >
                      <option value="regular">{{ $t('members_page.role_regular') }}</option>
                      <option value="admin">{{ $t('members_page.role_admin') }}</option>
                      <option value="readonly">{{ $t('members_page.role_readonly') }}</option>
                    </select>
                    <span v-else class="capitalize text-theme-text">{{
                      formatRole(member.role)
                    }}</span>
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap text-sm text-theme-textLight">
                    {{ formatDate(member.createdAt) }}
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap text-right text-sm">
                    <button
                      type="button"
                      class="text-red-600 hover:text-red-700 hover:underline disabled:opacity-50 disabled:cursor-not-allowed"
                      :disabled="!canEditMember(member) || deletingMemberId === member.userId"
                      @click="deleteMember(member)"
                    >
                      {{
                        deletingMemberId === member.userId
                          ? $t('members_page.removing')
                          : $t('members_page.remove')
                      }}
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </section>

        <section class="rounded-xl border border-theme-border bg-white shadow-sm overflow-hidden">
          <div class="border-b border-theme-border px-6 py-4">
            <h2 class="text-lg font-semibold text-theme-text">
              {{ $t('members_page.pending_invites') }}
            </h2>
          </div>
          <div v-if="loadingInvites" class="p-12 text-center">
            <svg
              class="animate-spin h-8 w-8 mx-auto text-theme-text"
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
            <p class="text-theme-textLight mt-3">{{ $t('members_page.loading_invites') }}</p>
          </div>
          <div
            v-else-if="invites.length === 0"
            class="p-12 text-center text-sm text-theme-textLight"
          >
            {{ $t('members_page.no_invites') }}
          </div>
          <div v-else class="overflow-x-auto">
            <table class="min-w-full divide-y divide-theme-border">
              <thead class="bg-theme-hover">
                <tr>
                  <th
                    class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider"
                  >
                    {{ $t('members_page.email') }}
                  </th>
                  <th
                    class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider"
                  >
                    {{ $t('members_page.role') }}
                  </th>
                  <th
                    class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider"
                  >
                    {{ $t('members_page.status') }}
                  </th>
                  <th
                    class="px-6 py-3 text-left text-xs font-medium text-theme-textLight uppercase tracking-wider"
                  >
                    {{ $t('members_page.expires') }}
                  </th>
                  <th
                    class="px-6 py-3 text-right text-xs font-medium text-theme-textLight uppercase tracking-wider"
                  >
                    {{ $t('members_page.actions') }}
                  </th>
                </tr>
              </thead>
              <tbody class="bg-white divide-y divide-theme-border">
                <tr v-for="invite in invites" :key="invite.id" class="hover:bg-theme-hover">
                  <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-theme-text">
                    {{ invite.email }}
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap text-sm text-theme-text capitalize">
                    {{ formatRole(invite.role) }}
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap text-sm text-theme-text capitalize">
                    {{ invite.status }}
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap text-sm text-theme-textLight">
                    {{ formatDate(invite.expiresAt) }}
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap text-right text-sm">
                    <button
                      type="button"
                      class="text-red-600 hover:text-red-700 hover:underline disabled:opacity-50 disabled:cursor-not-allowed"
                      :disabled="deletingInviteId === invite.id"
                      @click="deleteInvite(invite)"
                    >
                      {{
                        deletingInviteId === invite.id
                          ? $t('members_page.cancelling')
                          : $t('members_page.cancel_invite')
                      }}
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </section>
      </div>
    </main>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import AppNav from '@/comp/AppNav.vue'
import { showFlash } from '@/lib/flash'
import { workspace_invite_api, workspace_member_api } from '@/api'
import { useWorkspaceContext } from '@/lib/permission'
import { useI18n } from 'vue-i18n'
import { errorText } from '@/lib/error'
import { formatDate } from '@/utils/date'

const { t } = useI18n()
const { currentUser, currentWorkspace } = useWorkspaceContext()

const members = ref([])
const invites = ref([])
const loadingMembers = ref(false)
const loadingInvites = ref(false)
const inviteLoading = ref(false)
const updatingMemberId = ref(null)
const deletingMemberId = ref(null)
const deletingInviteId = ref(null)
const errorMessage = ref(null)

const inviteForm = reactive({
  email: '',
  role: 'regular',
})

function formatRole(role) {
  if (role === 'regular') return t('members_page.role_regular')
  if (role === 'admin') return t('members_page.role_admin')
  if (role === 'owner') return t('members_page.role_owner')
  if (role === 'readonly') return t('members_page.role_readonly')
  return role
}

function canEditMember(member) {
  if (member.role === 'owner') return false
  return member.userId !== currentUser?.id
}

async function loadMembers() {
  loadingMembers.value = true
  errorMessage.value = null

  try {
    const res = await workspace_member_api.list(currentWorkspace.id, { limit: 100, offset: 0 })
    members.value = res.data.members || []
  } catch (err) {
    errorMessage.value = errorText(err, t('members_page.failed_load_members'))
    members.value = []
  } finally {
    loadingMembers.value = false
  }
}

async function loadInvites() {
  loadingInvites.value = true
  errorMessage.value = null

  try {
    const res = await workspace_invite_api.list(currentWorkspace.id, { limit: 100, offset: 0 })
    invites.value = (res.data.invites || []).filter((invite) => invite.status === 'pending')
  } catch (err) {
    errorMessage.value = errorText(err, t('members_page.failed_load_invites'))
    invites.value = []
  } finally {
    loadingInvites.value = false
  }
}

async function createInvite() {
  inviteLoading.value = true
  errorMessage.value = null

  try {
    await workspace_invite_api.create(currentWorkspace.id, {
      email: inviteForm.email.trim(),
      role: inviteForm.role,
    })
    inviteForm.email = ''
    inviteForm.role = 'regular'
    await loadInvites()
    showFlash(t('common.sent'))
  } catch (err) {
    errorMessage.value = errorText(err, t('members_page.failed_create_invite'))
  } finally {
    inviteLoading.value = false
  }
}

async function updateMemberRole(member, role) {
  if (role === member.role) return
  updatingMemberId.value = member.userId
  errorMessage.value = null

  try {
    await workspace_member_api.updateRole(currentWorkspace.id, member.userId, { role })
    await loadMembers()
    showFlash(t('common.saved'))
  } catch (err) {
    errorMessage.value = errorText(err, t('members_page.failed_update_member'))
    await loadMembers()
  } finally {
    updatingMemberId.value = null
  }
}

async function deleteMember(member) {
  if (!window.confirm(t('members_page.remove_confirm', { email: member.email }))) return
  deletingMemberId.value = member.userId
  errorMessage.value = null

  try {
    await workspace_member_api.delete(currentWorkspace.id, member.userId)
    await loadMembers()
    showFlash(t('common.deleted'))
  } catch (err) {
    errorMessage.value = errorText(err, t('members_page.failed_remove_member'))
  } finally {
    deletingMemberId.value = null
  }
}

async function deleteInvite(invite) {
  if (!window.confirm(t('members_page.cancel_invite_confirm', { email: invite.email }))) return
  deletingInviteId.value = invite.id
  errorMessage.value = null

  try {
    await workspace_invite_api.delete(currentWorkspace.id, invite.id)
    await loadInvites()
    showFlash(t('common.cancelled'))
  } catch (err) {
    errorMessage.value = errorText(err, t('members_page.failed_cancel_invite'))
  } finally {
    deletingInviteId.value = null
  }
}

onMounted(() => {
  loadMembers()
  loadInvites()
})
</script>
