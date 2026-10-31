import { Marked } from 'marked'

// Links may only go to the web, email or a dashboard path; anything else (javascript:, data:) stays plain text.
const SAFE_LINK = /^(https?:|mailto:|\/(?!\/))/i

function escape(text) {
  return text
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;')
}

// Text we render can quote workspace data other people wrote, so the renderer itself is the sanitizer:
// raw HTML is shown as text, unsafe links lose their link and images show only their alt text.
const markdown = new Marked({ gfm: true, breaks: true })

markdown.use({
  renderer: {
    html({ text }) {
      return escape(text)
    },
    link({ href, title, tokens }) {
      const label = this.parser.parseInline(tokens)
      if (!SAFE_LINK.test(href)) return label

      const titleAttr = title ? ` title="${escape(title)}"` : ''
      const target = href.startsWith('/') ? '' : ' target="_blank" rel="noopener noreferrer"'

      return `<a href="${escape(href)}"${titleAttr}${target}>${label}</a>`
    },
    image({ text }) {
      return escape(text)
    },
  },
})

// Renders untrusted Markdown to HTML that is safe to put in v-html.
export function renderMarkdown(text) {
  return markdown.parse(text)
}

// Strips Markdown marks for one-line previews, keeping link labels and code text.
export function plainText(text) {
  return text
    .replace(/!?\[([^\]]*)\]\([^)]*\)/g, '$1')
    .replace(/[*`#>|~]/g, '')
    .replace(/\s+/g, ' ')
    .trim()
}
