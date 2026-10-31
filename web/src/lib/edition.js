import { setup_api } from '@/api'

const OSS = 'oss'
const SAAS = 'saas'

let edition = OSS
let loaded = false
let pending = null

export async function loadEdition() {
  if (loaded) {
    return edition
  }
  if (pending) {
    return pending
  }

  pending = setup_api
    .checkInstalled()
    .then((res) => {
      edition = res.data?.edition === SAAS ? SAAS : OSS
      loaded = true
      return edition
    })
    .catch(() => {
      edition = OSS
      loaded = true
      return edition
    })
    .finally(() => {
      pending = null
    })

  return pending
}

export function isSaaS() {
  return edition === SAAS
}
