// Mirrors the server-side agent and instance name pattern.
export const AGENT_NAME_PATTERN = /^[A-Za-z0-9][A-Za-z0-9._-]{0,99}$/

// Where instances serve their A2A card unless the agent sets another path.
export const AGENT_CARD_PATH = '/.well-known/agent-card.json'

export function healthClass(health) {
  return (
    {
      passing: 'bg-emerald-50 text-emerald-800',
      unknown: 'bg-slate-100 text-slate-700',
      warning: 'bg-amber-50 text-amber-800',
      critical: 'bg-red-50 text-red-800',
    }[health] || 'bg-theme-hover text-theme-textLight'
  )
}

export function healthDotClass(health) {
  return (
    {
      passing: 'bg-emerald-500',
      unknown: 'bg-slate-400',
      warning: 'bg-amber-500',
      critical: 'bg-red-500',
    }[health] || 'bg-theme-textLight'
  )
}

function asArray(value) {
  return Array.isArray(value) ? value : []
}

// Pulls the fields the UI shows out of an A2A agent card.
export function parseCard(card) {
  const body = card && typeof card === 'object' ? card : {}
  const capabilities =
    body.capabilities && typeof body.capabilities === 'object' ? body.capabilities : {}

  return {
    description: typeof body.description === 'string' ? body.description : '',
    url:
      typeof body.url === 'string'
        ? body.url
        : String(asArray(body.supportedInterfaces)[0]?.url ?? ''),
    provider: body.provider?.organization || '',
    skills: asArray(body.skills)
      .filter((skill) => skill && typeof skill === 'object')
      .map((skill) => ({
        id: String(skill.id ?? ''),
        name: String(skill.name ?? ''),
        description: String(skill.description ?? ''),
        tags: asArray(skill.tags).map(String),
      })),
    inputModes: asArray(body.defaultInputModes).map(String),
    outputModes: asArray(body.defaultOutputModes).map(String),
    capabilities: Object.keys(capabilities).filter((key) => capabilities[key] === true),
  }
}

export function countByStatus(instances) {
  const counts = { critical: 0, unknown: 0, warning: 0, passing: 0 }
  asArray(instances).forEach((instance) => {
    counts[instance.status] = (counts[instance.status] || 0) + 1
  })

  return counts
}
