import { computed } from 'vue'
import { loadUserFromStorage, loadWorkspaceFromStorage } from '@/utils/storage'

export const WORKSPACE_KEY_PERMISSION_GROUPS = [
  {
    labelKey: 'workspace_settings_page.perm_group_workspace',
    permissions: [
      { value: 'CAN_GET_WORKSPACE', labelKey: 'workspace_settings_page.perm_get_workspace' },
      { value: 'CAN_UPDATE_WORKSPACE', labelKey: 'workspace_settings_page.perm_update_workspace' },
      { value: 'CAN_DELETE_WORKSPACE', labelKey: 'workspace_settings_page.perm_delete_workspace' },
    ],
  },
  {
    labelKey: 'workspace_settings_page.perm_group_members',
    permissions: [
      {
        value: 'CAN_LIST_WORKSPACE_MEMBERS',
        labelKey: 'workspace_settings_page.perm_list_workspace_members',
      },
      {
        value: 'CAN_UPDATE_MEMBER_ROLE',
        labelKey: 'workspace_settings_page.perm_update_member_role',
      },
      { value: 'CAN_REMOVE_MEMBER', labelKey: 'workspace_settings_page.perm_remove_member' },
    ],
  },
  {
    labelKey: 'workspace_settings_page.perm_group_invites',
    permissions: [
      {
        value: 'CAN_LIST_WORKSPACE_INVITES',
        labelKey: 'workspace_settings_page.perm_list_workspace_invites',
      },
      {
        value: 'CAN_GET_WORKSPACE_INVITE',
        labelKey: 'workspace_settings_page.perm_get_workspace_invite',
      },
      { value: 'CAN_REMOVE_INVITE', labelKey: 'workspace_settings_page.perm_remove_invite' },
    ],
  },
  {
    labelKey: 'workspace_settings_page.perm_group_billing',
    permissions: [
      {
        value: 'CAN_GET_WORKSPACE_BILLING',
        labelKey: 'workspace_settings_page.perm_get_workspace_billing',
      },
      {
        value: 'CAN_UPDATE_WORKSPACE_BILLING',
        labelKey: 'workspace_settings_page.perm_update_workspace_billing',
      },
    ],
  },
  {
    labelKey: 'workspace_settings_page.perm_group_access_keys',
    permissions: [
      {
        value: 'CAN_LIST_WORKSPACE_ACCESS_KEYS',
        labelKey: 'workspace_settings_page.perm_list_workspace_access_keys',
      },
      {
        value: 'CAN_CREATE_WORKSPACE_ACCESS_KEY',
        labelKey: 'workspace_settings_page.perm_create_workspace_access_key',
      },
      {
        value: 'CAN_GET_WORKSPACE_ACCESS_KEY',
        labelKey: 'workspace_settings_page.perm_get_workspace_access_key',
      },
      {
        value: 'CAN_DELETE_WORKSPACE_ACCESS_KEY',
        labelKey: 'workspace_settings_page.perm_delete_workspace_access_key',
      },
    ],
  },
  {
    labelKey: 'workspace_settings_page.perm_group_agents',
    permissions: [
      { value: 'CAN_LIST_AGENTS', labelKey: 'workspace_settings_page.perm_list_agents' },
      { value: 'CAN_GET_AGENT', labelKey: 'workspace_settings_page.perm_get_agent' },
      { value: 'CAN_UPSERT_AGENT', labelKey: 'workspace_settings_page.perm_upsert_agent' },
      { value: 'CAN_DELETE_AGENT', labelKey: 'workspace_settings_page.perm_delete_agent' },
    ],
  },
  {
    labelKey: 'workspace_settings_page.perm_group_instances',
    permissions: [
      { value: 'CAN_LIST_INSTANCES', labelKey: 'workspace_settings_page.perm_list_instances' },
      { value: 'CAN_GET_INSTANCE', labelKey: 'workspace_settings_page.perm_get_instance' },
      {
        value: 'CAN_REGISTER_INSTANCE',
        labelKey: 'workspace_settings_page.perm_register_instance',
      },
      { value: 'CAN_REPORT_CHECK', labelKey: 'workspace_settings_page.perm_report_check' },
    ],
  },
  {
    labelKey: 'workspace_settings_page.perm_group_intentions',
    permissions: [
      { value: 'CAN_LIST_INTENTIONS', labelKey: 'workspace_settings_page.perm_list_intentions' },
      { value: 'CAN_GET_INTENTION', labelKey: 'workspace_settings_page.perm_get_intention' },
      { value: 'CAN_CREATE_INTENTION', labelKey: 'workspace_settings_page.perm_create_intention' },
      { value: 'CAN_UPDATE_INTENTION', labelKey: 'workspace_settings_page.perm_update_intention' },
      { value: 'CAN_DELETE_INTENTION', labelKey: 'workspace_settings_page.perm_delete_intention' },
      { value: 'CAN_CHECK_INTENTION', labelKey: 'workspace_settings_page.perm_check_intention' },
    ],
  },
  {
    labelKey: 'workspace_settings_page.perm_group_integrations',
    permissions: [
      {
        value: 'CAN_LIST_INTEGRATIONS',
        labelKey: 'workspace_settings_page.perm_list_integrations',
      },
      { value: 'CAN_GET_INTEGRATION', labelKey: 'workspace_settings_page.perm_get_integration' },
      {
        value: 'CAN_CREATE_INTEGRATION',
        labelKey: 'workspace_settings_page.perm_create_integration',
      },
      {
        value: 'CAN_UPDATE_INTEGRATION',
        labelKey: 'workspace_settings_page.perm_update_integration',
      },
      {
        value: 'CAN_DELETE_INTEGRATION',
        labelKey: 'workspace_settings_page.perm_delete_integration',
      },
    ],
  },
  {
    labelKey: 'workspace_settings_page.perm_group_events',
    permissions: [
      { value: 'CAN_LIST_EVENTS', labelKey: 'workspace_settings_page.perm_list_events' },
      { value: 'CAN_GET_EVENT', labelKey: 'workspace_settings_page.perm_get_event' },
    ],
  },
  {
    labelKey: 'workspace_settings_page.perm_group_key_value',
    permissions: [
      { value: 'CAN_LIST_KEY_VALUE', labelKey: 'workspace_settings_page.perm_list_key_value' },
      { value: 'CAN_GET_KEY_VALUE', labelKey: 'workspace_settings_page.perm_get_key_value' },
      { value: 'CAN_PUT_KEY_VALUE', labelKey: 'workspace_settings_page.perm_put_key_value' },
      { value: 'CAN_DELETE_KEY_VALUE', labelKey: 'workspace_settings_page.perm_delete_key_value' },
    ],
  },
  {
    labelKey: 'workspace_settings_page.perm_group_traffic',
    permissions: [
      { value: 'CAN_LIST_TRAFFIC', labelKey: 'workspace_settings_page.perm_list_traffic' },
      { value: 'CAN_GET_TRAFFIC', labelKey: 'workspace_settings_page.perm_get_traffic' },
    ],
  },
]

export const WORKSPACE_KEY_PERMISSIONS = WORKSPACE_KEY_PERMISSION_GROUPS.flatMap(
  (group) => group.permissions
)

export const PLATFORM_ROLE_ADMIN = 'admin'
export const PLATFORM_ROLE_REGULAR = 'regular'
export const WORKSPACE_ROLE_READONLY = 'readonly'
export const WORKSPACE_ROLE_ADMIN = 'admin'
export const WORKSPACE_ROLE_OWNER = 'owner'
export const WORKSPACE_ROLE_REGULAR = 'regular'

const PLATFORM_ROLES = new Set([PLATFORM_ROLE_ADMIN, PLATFORM_ROLE_REGULAR])

const WORKSPACE_MEMBERSHIP_ROLES = new Set([
  WORKSPACE_ROLE_READONLY,
  WORKSPACE_ROLE_REGULAR,
  WORKSPACE_ROLE_ADMIN,
  WORKSPACE_ROLE_OWNER,
])

export function getPlatformRole(user) {
  const r = user?.role
  if (r == null || r === '') return null
  return PLATFORM_ROLES.has(r) ? r : null
}

export function getWorkspaceMembershipRole(workspace) {
  const r = workspace?.role
  if (r == null || r === '') return null
  return WORKSPACE_MEMBERSHIP_ROLES.has(r) ? r : null
}

export function canManageWorkspace(user, workspace) {
  if (!workspace?.id) return false
  return workspace.role === WORKSPACE_ROLE_ADMIN || workspace.role === WORKSPACE_ROLE_OWNER
}

export function useWorkspaceContext() {
  const currentUser = loadUserFromStorage()
  const currentWorkspace = loadWorkspaceFromStorage()
  const canManage = computed(() => canManageWorkspace(currentUser, currentWorkspace))

  return { currentUser, currentWorkspace, canManage }
}
