import { createRouter, createWebHistory } from 'vue-router'
import { loadUserFromStorage, loadWorkspaceFromStorage } from '@/utils/storage'
import { canManageWorkspace } from '@/lib/permission'
import { loadEdition, isSaaS } from '@/lib/edition'

const routes = [
  {
    path: '/',
    redirect: '/login',
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
    meta: { requiresGuest: true, title: 'Login', description: 'Login to platform' },
  },
  {
    path: '/setup',
    name: 'Setup',
    component: () => import('@/views/Setup.vue'),
    meta: { requiresGuest: true, title: 'Setup', description: 'Setup platform' },
  },
  {
    path: '/dashboard',
    name: 'Dashboard',
    component: () => import('@/views/Dashboard.vue'),
    meta: {
      requiresAuth: true,
      requiresWorkspace: true,
      title: 'Dashboard',
      description: 'Platform dashboard',
    },
  },
  {
    path: '/agents',
    name: 'Agents',
    component: () => import('@/views/Agents.vue'),
    meta: {
      requiresAuth: true,
      requiresWorkspace: true,
      title: 'Agents',
      description: 'Registered agents',
    },
  },
  {
    path: '/agents/:name',
    name: 'Agent',
    component: () => import('@/views/Agent.vue'),
    meta: {
      requiresAuth: true,
      requiresWorkspace: true,
      title: 'Agent',
      description: 'Agent card and instances',
    },
  },
  {
    path: '/intentions',
    name: 'Intentions',
    component: () => import('@/views/Intentions.vue'),
    meta: {
      requiresAuth: true,
      requiresWorkspace: true,
      title: 'Intentions',
      description: 'Who may call whom',
    },
  },
  {
    path: '/kv',
    name: 'KV',
    component: () => import('@/views/KV.vue'),
    meta: {
      requiresAuth: true,
      requiresWorkspace: true,
      title: 'KV',
      description: 'Workspace key-value config',
    },
  },
  {
    path: '/traffic',
    name: 'Traffic',
    component: () => import('@/views/Traffic.vue'),
    meta: {
      requiresAuth: true,
      requiresWorkspace: true,
      title: 'Traffic',
      description: 'Gateway calls',
    },
  },
  {
    path: '/select-workspace',
    name: 'SwitchWorkspaces',
    component: () => import('@/views/SelectWorkspace.vue'),
    meta: {
      requiresAuth: true,
      title: 'Switch Workspaces',
      description: 'Choose or create a workspace',
    },
  },
  {
    path: '/workspaces',
    name: 'Workspaces',
    component: () => import('@/views/Workspaces.vue'),
    meta: {
      requiresAuth: true,
      requiresWorkspace: true,
      title: 'Workspaces',
      description: 'Workspaces',
    },
  },
  {
    path: '/profile',
    name: 'Profile',
    component: () => import('@/views/Profile.vue'),
    meta: { requiresAuth: true, title: 'Profile', description: 'Your account profile' },
  },
  {
    path: '/settings',
    name: 'Settings',
    component: () => import('@/views/Settings.vue'),
    meta: {
      requiresAuth: true,
      requiresWorkspace: true,
      requiresWorkspaceAdmin: true,
      title: 'Settings',
      description: 'Settings',
    },
  },
  {
    path: '/members',
    name: 'Members',
    component: () => import('@/views/Members.vue'),
    meta: {
      requiresAuth: true,
      requiresWorkspace: true,
      requiresWorkspaceAdmin: true,
      title: 'Members',
      description: 'Members',
    },
  },
  {
    path: '/billing',
    name: 'Billing',
    component: () => import('@/views/Billing.vue'),
    meta: {
      requiresAuth: true,
      requiresWorkspace: true,
      requiresWorkspaceAdmin: true,
      requiresSaaS: true,
      title: 'Billing',
      description: 'Billing',
    },
  },
  {
    path: '/integrations',
    name: 'Integrations',
    component: () => import('@/views/Integrations.vue'),
    meta: {
      requiresAuth: true,
      requiresWorkspace: true,
      requiresWorkspaceAdmin: true,
      title: 'Integrations',
      description: 'Integrations',
    },
  },
  {
    path: '/events',
    name: 'Events',
    component: () => import('@/views/Events.vue'),
    meta: {
      requiresAuth: true,
      requiresWorkspace: true,
      requiresWorkspaceAdmin: true,
      title: 'Events',
      description: 'Workspace events',
    },
  },
  {
    path: '/404',
    name: 'NotFound',
    component: () => import('@/views/NotFound.vue'),
    meta: { title: 'Page not found' },
  },
  {
    path: '/500',
    name: 'ServerError',
    component: () => import('@/views/ServerError.vue'),
    meta: { title: 'Server error' },
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFoundCatchAll',
    component: () => import('@/views/NotFound.vue'),
    meta: { title: 'Page not found' },
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior(to, _from, savedPosition) {
    if (savedPosition) {
      return savedPosition
    }
    if (to.hash) {
      return { el: to.hash, behavior: 'smooth' }
    }
    return { top: 0 }
  },
})

function isSafeRedirect(redirect) {
  return typeof redirect === 'string' && redirect.startsWith('/') && !redirect.startsWith('//')
}

// Navigation guard - reads auth state directly from storage
router.beforeEach(async (to) => {
  await loadEdition()

  const currentUser = loadUserFromStorage()
  const currentWorkspace = loadWorkspaceFromStorage()
  const isAuthenticated = !!currentUser

  if (to.meta.requiresSaaS && !isSaaS()) {
    return '/404'
  }
  if (to.meta.requiresAuth && !isAuthenticated) {
    return { path: '/login', query: { redirect: to.fullPath } }
  }
  if (to.meta.requiresGuest && isAuthenticated) {
    const redirect = to.query.redirect
    return isSafeRedirect(redirect) ? redirect : '/select-workspace'
  }
  if (to.meta.requiresWorkspace && !currentWorkspace) {
    return '/select-workspace'
  }
  if (to.meta.requiresWorkspaceAdmin && !canManageWorkspace(currentUser, currentWorkspace)) {
    return '/404'
  }
})

export default router
