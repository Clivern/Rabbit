import axios from 'axios'
import { removeUserFromStorage, removeWorkspaceFromStorage } from '@/utils/storage'

const api = axios.create({
  baseURL: '/api/v1',
  timeout: 10000,
})

// The user's chosen language, sent so API messages come back translated
function localeHeaders() {
  try {
    const locale = typeof localStorage !== 'undefined' && localStorage.getItem('_flocale')
    if (locale === 'en' || locale === 'fr') {
      return { 'Accept-Language': locale }
    }
  } catch (_) {}
  return {}
}

// Request interceptor
api.interceptors.request.use((config) => {
  Object.assign(config.headers, localeHeaders())
  return config
})

// Endpoints and pages where a 401 does not redirect to login
const noRedirectUrls = ['public/action/logout']
const publicPages = ['/', '/login', '/setup']

// Response interceptor
api.interceptors.response.use(
  (response) => response,
  (error) => {
    const { pathname, search } = window.location

    if (
      error.response?.status === 401 &&
      !noRedirectUrls.includes(error.config?.url) &&
      !publicPages.includes(pathname)
    ) {
      removeUserFromStorage()
      removeWorkspaceFromStorage()
      window.location.href = `/login?redirect=${encodeURIComponent(pathname + search)}`
    }
    return Promise.reject(error)
  }
)

// API endpoints
export const health_api = {
  check: () => api.get('public/_health'),
  ready: () => api.get('public/_ready'),
}

// API endpoints for the current user's session and profile
export const auth_api = {
  logout: () => api.post('public/action/logout'),
  me: () => api.get('/me'),
  getProfile: () => api.get('/action/profile'),
  updateProfile: (data) => api.put('/action/profile', data),
}

// API endpoints for the first-run setup
export const setup_api = {
  install: (data) => api.post('public/action/setup', data),
  checkInstalled: () => api.get('public/action/setup/status'),
}

// API endpoints for settings
export const settings_api = {
  getSettings: () => api.get('/action/settings'),
  updateSettings: (data) => api.put('/action/settings', data),
}

// API endpoints for workspaces (projects)
export const workspace_api = {
  list: (params) => api.get('/workspaces', { params }),
  get: (id) => api.get(`/workspaces/${id}`),
  create: (data) => api.post('/workspaces', data),
  update: (id, data) => api.put(`/workspaces/${id}`, data),
  delete: (id) => api.delete(`/workspaces/${id}`),
}

// API endpoints for workspace members
export const workspace_member_api = {
  list: (workspaceId, params) => api.get(`/workspaces/${workspaceId}/members`, { params }),
  updateRole: (workspaceId, userId, data) =>
    api.put(`/workspaces/${workspaceId}/members/${userId}`, data),
  delete: (workspaceId, userId) => api.delete(`/workspaces/${workspaceId}/members/${userId}`),
}

// API endpoints for workspace invites
export const workspace_invite_api = {
  list: (workspaceId, params) => api.get(`/workspaces/${workspaceId}/invites`, { params }),
  create: (workspaceId, data) => api.post(`/workspaces/${workspaceId}/invites`, data),
  delete: (workspaceId, id) => api.delete(`/workspaces/${workspaceId}/invites/${id}`),
}

// API endpoints for workspace billing
export const billing_api = {
  status: (workspaceId) => api.get(`/workspaces/${workspaceId}/billing`),
  usage: (workspaceId) => api.get(`/workspaces/${workspaceId}/billing/usage`),
  checkout: (workspaceId, data) => api.post(`/workspaces/${workspaceId}/billing/checkout`, data),
  portal: (workspaceId) => api.post(`/workspaces/${workspaceId}/billing/portal`),
}

// API endpoints for workspace access keys
export const workspace_access_key_api = {
  list: (workspaceId, params) => api.get(`/workspaces/${workspaceId}/keys`, { params }),
  create: (workspaceId, data) => api.post(`/workspaces/${workspaceId}/keys`, data),
  delete: (workspaceId, id) => api.delete(`/workspaces/${workspaceId}/keys/${id}`),
}

// API endpoints for workspace agents
export const agent_api = {
  list: (workspaceId, params) => api.get(`/workspaces/${workspaceId}/agents`, { params }),
  get: (workspaceId, name) =>
    api.get(`/workspaces/${workspaceId}/agents/${encodeURIComponent(name)}`),
  upsert: (workspaceId, name, data) =>
    api.put(`/workspaces/${workspaceId}/agents/${encodeURIComponent(name)}`, data),
  delete: (workspaceId, name) =>
    api.delete(`/workspaces/${workspaceId}/agents/${encodeURIComponent(name)}`),
}

// API endpoints for agent health checks
export const agent_check_api = {
  list: (workspaceId, name) =>
    api.get(`/workspaces/${workspaceId}/agents/${encodeURIComponent(name)}/checks`),
  upsert: (workspaceId, name, checkId, data) =>
    api.put(
      `/workspaces/${workspaceId}/agents/${encodeURIComponent(name)}/checks/${encodeURIComponent(checkId)}`,
      data
    ),
  delete: (workspaceId, name, checkId) =>
    api.delete(
      `/workspaces/${workspaceId}/agents/${encodeURIComponent(name)}/checks/${encodeURIComponent(checkId)}`
    ),
}

// API endpoints for agent identities
export const agent_identity_api = {
  list: (workspaceId, name) =>
    api.get(`/workspaces/${workspaceId}/agents/${encodeURIComponent(name)}/identities`),
  create: (workspaceId, name, data) =>
    api.post(`/workspaces/${workspaceId}/agents/${encodeURIComponent(name)}/identities`, data),
  update: (workspaceId, name, id, data) =>
    api.put(`/workspaces/${workspaceId}/agents/${encodeURIComponent(name)}/identities/${id}`, data),
  delete: (workspaceId, name, id) =>
    api.delete(`/workspaces/${workspaceId}/agents/${encodeURIComponent(name)}/identities/${id}`),
}

// API endpoints for agent protection
export const agent_protection_api = {
  get: (workspaceId, name) =>
    api.get(`/workspaces/${workspaceId}/agents/${encodeURIComponent(name)}/protection`),
  save: (workspaceId, name, data) =>
    api.put(`/workspaces/${workspaceId}/agents/${encodeURIComponent(name)}/protection`, data),
  updateStatus: (workspaceId, name, data) =>
    api.put(
      `/workspaces/${workspaceId}/agents/${encodeURIComponent(name)}/protection/status`,
      data
    ),
  delete: (workspaceId, name) =>
    api.delete(`/workspaces/${workspaceId}/agents/${encodeURIComponent(name)}/protection`),
}

// API endpoints for agent instances
export const agent_instance_api = {
  list: (workspaceId, name, params) =>
    api.get(`/workspaces/${workspaceId}/agents/${encodeURIComponent(name)}/instances`, { params }),
  delete: (workspaceId, name, instanceId) =>
    api.delete(
      `/workspaces/${workspaceId}/agents/${encodeURIComponent(name)}/instances/${encodeURIComponent(instanceId)}`
    ),
}

// API endpoints for workspace intentions
export const intention_api = {
  list: (workspaceId, params) => api.get(`/workspaces/${workspaceId}/intentions`, { params }),
  create: (workspaceId, data) => api.put(`/workspaces/${workspaceId}/intentions`, data),
  update: (workspaceId, id, data) => api.put(`/workspaces/${workspaceId}/intentions/${id}`, data),
  delete: (workspaceId, id) => api.delete(`/workspaces/${workspaceId}/intentions/${id}`),
}

// API endpoints for workspace event integrations
export const integration_api = {
  list: (workspaceId, params) => api.get(`/workspaces/${workspaceId}/integrations`, { params }),
  create: (workspaceId, data) => api.post(`/workspaces/${workspaceId}/integrations`, data),
  update: (workspaceId, id, data) => api.put(`/workspaces/${workspaceId}/integrations/${id}`, data),
  delete: (workspaceId, id) => api.delete(`/workspaces/${workspaceId}/integrations/${id}`),
}

// API endpoints for workspace events
export const event_api = {
  list: (workspaceId, params) => api.get(`/workspaces/${workspaceId}/events`, { params }),
  get: (workspaceId, id) => api.get(`/workspaces/${workspaceId}/events/${id}`),
}

// API endpoints for workspace gateway traffic
export const traffic_api = {
  list: (workspaceId, params) => api.get(`/workspaces/${workspaceId}/traffic`, { params }),
  get: (workspaceId, id) => api.get(`/workspaces/${workspaceId}/traffic/${id}`),
}

// API endpoints for the current user's dashboard chats
export const chat_api = {
  list: (workspaceId) => api.get(`/workspaces/${workspaceId}/chats`),
  create: (workspaceId) => api.post(`/workspaces/${workspaceId}/chats`),
  get: (workspaceId, id) => api.get(`/workspaces/${workspaceId}/chats/${id}`),
  delete: (workspaceId, id) => api.delete(`/workspaces/${workspaceId}/chats/${id}`),
  send: (workspaceId, id, data, onEvent) =>
    streamEvents(`/workspaces/${workspaceId}/chats/${id}/messages`, data, onEvent),
}

// POSTs data and calls onEvent(name, data) for each server-sent event until the stream ends.
// Axios cannot read a streamed body in the browser and EventSource only does GET, so this uses fetch.
// A refused request throws an axios-shaped error, so errorText reads its message.
async function streamEvents(url, data, onEvent) {
  const res = await fetch(`${api.defaults.baseURL}${url}`, {
    method: 'POST',
    credentials: 'same-origin',
    headers: {
      'Content-Type': 'application/json',
      Accept: 'text/event-stream',
      ...localeHeaders(),
    },
    body: JSON.stringify(data),
  })

  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw { response: { status: res.status, data: body } }
  }

  const reader = res.body.pipeThrough(new TextDecoderStream()).getReader()
  let buffer = ''

  for (;;) {
    const { value, done } = await reader.read()
    if (done) return

    buffer += value
    const events = buffer.split('\n\n')
    buffer = events.pop()

    for (const raw of events) {
      const name = raw.match(/^event: (.*)$/m)?.[1]
      const payload = raw.match(/^data: (.*)$/m)?.[1]
      if (name && payload) onEvent(name, JSON.parse(payload))
    }
  }
}

// API endpoints for the current user's workspace notifications
export const notification_api = {
  list: (workspaceId, params) => api.get(`/workspaces/${workspaceId}/notifications`, { params }),
  countUnread: (workspaceId) => api.get(`/workspaces/${workspaceId}/notifications/unread`),
  markRead: (workspaceId, id) => api.post(`/workspaces/${workspaceId}/notifications/${id}/read`),
  markAllRead: (workspaceId) => api.post(`/workspaces/${workspaceId}/notifications/read`),
  dismiss: (workspaceId, id) => api.post(`/workspaces/${workspaceId}/notifications/${id}/dismiss`),
}

// Keys may contain slashes, so they are not encoded
export const kv_api = {
  list: (workspaceId, params) => api.get(`/workspaces/${workspaceId}/kv`, { params }),
  get: (workspaceId, key) => api.get(`/workspaces/${workspaceId}/kv/${key}`),
  put: (workspaceId, key, data) => api.put(`/workspaces/${workspaceId}/kv/${key}`, data),
  delete: (workspaceId, key) => api.delete(`/workspaces/${workspaceId}/kv/${key}`),
}

// API endpoints for current user's API keys
export const api_keys_api = {
  list: (params) => api.get('/apiKeys', { params }),
  create: (data) => api.post('/apiKeys', data),
  delete: (id) => api.delete(`/apiKeys/${id}`),
}
