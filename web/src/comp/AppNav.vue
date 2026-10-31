<template>
  <nav class="bg-white border-b border-theme-border">
    <div class="w-full px-4 sm:px-6 lg:px-8">
      <div class="flex justify-between h-14">
        <div class="flex items-center">
          <div class="shrink-0 flex items-center">
            <router-link to="/" class="text-theme-text">
              <AppLogo variant="tile" :alt="$t('common.logo_alt')" class="h-8 w-auto" />
            </router-link>
          </div>
          <div class="hidden md:ml-8 md:flex md:space-x-1">
            <router-link
              v-for="item in navItems"
              :key="item.to"
              :to="item.to"
              :class="[isActive(item.to) ? 'nav-link-active' : 'nav-link']"
            >
              {{ item.label }}
            </router-link>
          </div>
        </div>
        <div class="relative flex items-center" ref="dropdownRef">
          <button
            type="button"
            @click.stop="toggleDropdown"
            class="flex items-center gap-2 rounded-full border border-theme-border bg-white pl-1 pr-2.5 py-1 text-left hover:bg-theme-hover focus:outline-none focus:ring-2 focus:ring-primary-800 focus:ring-offset-1 transition-colors"
            :aria-expanded="open"
            aria-haspopup="true"
            :aria-label="isMobile ? $t('nav.menu') : $t('nav.user_menu')"
          >
            <span
              class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-primary-200 text-sm font-medium text-theme-text"
              aria-hidden="true"
            >
              <img
                :src="user?.avatar ?? ''"
                :alt="user?.email ?? ''"
                class="h-8 w-8 rounded-full"
              />
            </span>
            <svg
              class="h-4 w-4 text-theme-textLight transition-transform"
              :class="{ 'rotate-180': open }"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              aria-hidden="true"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M19 9l-7 7-7-7"
              />
            </svg>
          </button>
          <Transition
            enter-active-class="transition ease-out duration-100"
            enter-from-class="opacity-0 scale-95"
            enter-to-class="opacity-100 scale-100"
            leave-active-class="transition ease-in duration-75"
            leave-from-class="opacity-100 scale-100"
            leave-to-class="opacity-0 scale-95"
          >
            <div
              v-show="open"
              class="absolute right-0 top-full z-50 mt-2 w-56 max-h-[24rem] overflow-y-auto rounded-lg border border-theme-border bg-white py-1 shadow-lg origin-top-right focus:outline-none"
              role="menu"
            >
              <!-- Mobile nav links (only on small screens) -->
              <div v-if="isMobile && navItems.length" class="border-b border-theme-border py-1">
                <router-link
                  v-for="item in navItems"
                  :key="item.to"
                  :to="item.to"
                  :class="[
                    isActive(item.to)
                      ? 'bg-theme-hover text-theme-text'
                      : 'text-theme-text hover:bg-theme-hover',
                  ]"
                  class="block px-4 py-2 text-sm font-medium"
                  role="menuitem"
                  @click="open = false"
                >
                  {{ item.label }}
                </router-link>
              </div>
              <div class="border-b border-theme-border px-4 py-3">
                <p v-if="user?.name" class="text-sm font-semibold text-theme-text truncate">
                  {{ user.name }}
                </p>
                <p
                  class="text-sm truncate"
                  :class="user?.name ? 'text-theme-textLight' : 'text-theme-text'"
                >
                  {{ user?.email ?? '' }}
                </p>
              </div>
              <div class="py-1">
                <router-link
                  to="/profile"
                  class="block px-4 py-2 text-sm text-theme-text hover:bg-theme-hover"
                  role="menuitem"
                  @click="open = false"
                >
                  {{ $t('nav.profile') }}
                </router-link>
                <router-link
                  to="/select-workspace"
                  class="block px-4 py-2 text-sm text-theme-text hover:bg-theme-hover"
                  role="menuitem"
                  @click="open = false"
                >
                  {{ $t('nav.select_workspace') }}
                </router-link>
                <button
                  type="button"
                  class="block w-full px-4 py-2 text-left text-sm text-theme-text hover:bg-theme-hover"
                  role="menuitem"
                  @click="handleLogout"
                >
                  {{ $t('nav.logout') }}
                </button>
              </div>
            </div>
          </Transition>
        </div>
      </div>
    </div>
  </nav>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { auth_api } from '@/api'
import { removeWorkspaceFromStorage, loadWorkspaceFromStorage } from '@/utils/storage'
import { user, clearUser } from '@/lib/auth'
import {
  applyLocale,
  applyTheme,
  applyUserPreferences,
  readStoredLocale,
  readStoredTheme,
} from '@/lib/preferences'
import { canManageWorkspace } from '@/lib/permission'
import { isSaaS } from '@/lib/edition'
import AppLogo from '@/comp/AppLogo.vue'

const route = useRoute()
const router = useRouter()
const dropdownRef = ref(null)
const open = ref(false)
const isMobile = ref(false)

const MOBILE_BREAKPOINT = 768

const { t } = useI18n()

function initPreferences() {
  if (user.value?.theme && user.value?.language) {
    applyUserPreferences(user.value)
    return
  }
  applyTheme(readStoredTheme())
  applyLocale(readStoredLocale())
}

const canManageWorkspaceNav = computed(() => {
  void route.path
  return canManageWorkspace(user.value, loadWorkspaceFromStorage())
})

const navItems = computed(() => {
  return [
    { to: '/dashboard', label: t('nav.dashboard') },
    { to: '/agents', label: t('nav.agents') },
    { to: '/intentions', label: t('nav.intentions') },
    { to: '/kv', label: t('nav.kv') },
    { to: '/traffic', label: t('nav.traffic') },
    { to: '/integrations', label: 'Integrations', requiresManage: true },
    { to: '/events', label: t('nav.events'), requiresManage: true },
    { to: '/billing', label: t('nav.billing'), requiresManage: true, requiresSaaS: true },
    { to: '/members', label: 'Members', requiresManage: true },
    { to: '/settings', label: t('nav.settings'), requiresManage: true },
    { to: '/workspaces', label: 'Workspaces' },
  ].filter(
    (item) =>
      (!item.requiresManage || canManageWorkspaceNav.value) && (!item.requiresSaaS || isSaaS())
  )
})

function isActive(path) {
  return route.path === path || (path !== '/dashboard' && route.path.startsWith(path + '/'))
}

function checkMobile() {
  isMobile.value = window.innerWidth < MOBILE_BREAKPOINT
}

function toggleDropdown() {
  open.value = !open.value
}

async function handleLogout() {
  open.value = false
  removeWorkspaceFromStorage()
  clearUser()

  try {
    await auth_api.logout()
  } catch (err) {
    console.error('Logout failed:', err)
  }
  router.push('/login')
}

function onClickOutside(event) {
  if (dropdownRef.value && !dropdownRef.value.contains(event.target)) {
    open.value = false
  }
}

onMounted(() => {
  initPreferences()
  checkMobile()
  window.addEventListener('resize', checkMobile)
  document.addEventListener('click', onClickOutside)
})

onUnmounted(() => {
  window.removeEventListener('resize', checkMobile)
  document.removeEventListener('click', onClickOutside)
})
</script>
