function escapeHtml(text) {
  return text.replaceAll('&', '&amp;').replaceAll('<', '&lt;').replaceAll('>', '&gt;')
}

function token(cls, value) {
  return `<span class="${cls}">${escapeHtml(value)}</span>`
}

function tokenizeLine(line, specs) {
  let pos = 0
  let html = ''

  while (pos < line.length) {
    const slice = line.slice(pos)
    let matched = false

    for (const spec of specs) {
      const match = slice.match(spec.pattern)
      if (!match || match.index !== 0) continue

      html += spec.cls ? token(spec.cls, match[0]) : escapeHtml(match[0])
      pos += match[0].length
      matched = true
      break
    }

    if (!matched) {
      html += escapeHtml(line[pos])
      pos += 1
    }
  }

  return html
}

const pythonSpecs = [
  { pattern: /^#.*$/, cls: 'token-comment' },
  { pattern: /^"(\\.|[^"])*"/, cls: 'token-string' },
  { pattern: /^'(\\.|[^'])*'/, cls: 'token-string' },
  { pattern: /^(from|import|print)\b/, cls: 'token-keyword' },
  { pattern: /^\d+\b/, cls: 'token-var' },
  { pattern: /^\.[a-z_]+(?=\()/, cls: 'token-fn' },
  { pattern: /^[a-z_][a-z0-9_]*(?=\()/, cls: 'token-fn' },
  { pattern: /^\s+/, cls: null },
]

const goSpecs = [
  { pattern: /^\/\/.*$/, cls: 'token-comment' },
  { pattern: /^"(\\.|[^"])*"/, cls: 'token-string' },
  { pattern: /^(package|import|func|return|var|ctx)\b/, cls: 'token-keyword' },
  { pattern: /^:=|,|\{|\}|\(|\)/, cls: null },
  { pattern: /^\.[A-Z][A-Za-z0-9_]*/, cls: 'token-fn' },
  { pattern: /^sloth\.[A-Za-z0-9_]+/, cls: 'token-fn' },
  { pattern: /^sloth\.[A-Za-z0-9_]+/, cls: 'token-fn' },
  { pattern: /^\s+/, cls: null },
]

const curlSpecs = [
  { pattern: /^#.*$/, cls: 'token-comment' },
  { pattern: /^"(\\.|[^"])*"/, cls: 'token-string' },
  { pattern: /^'(\\.|[^'])*'/, cls: 'token-string' },
  { pattern: /^curl\b/, cls: 'token-fn' },
  { pattern: /^-(X|H|d)\b/, cls: 'token-keyword' },
  { pattern: /^\\$/, cls: null },
  { pattern: /^\s+/, cls: null },
]

const languageSpecs = {
  python: pythonSpecs,
  go: goSpecs,
  curl: curlSpecs,
}

export function highlightCode(code, language) {
  const specs = languageSpecs[language]
  if (!specs) return escapeHtml(code)

  return code
    .split('\n')
    .map((line) => tokenizeLine(line, specs))
    .join('\n')
}
