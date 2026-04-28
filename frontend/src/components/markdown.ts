import MarkdownIt from 'markdown-it'

const markdown = new MarkdownIt({
  breaks: true,
  html: false,
  linkify: true,
})

const defaultLinkOpen = markdown.renderer.rules.link_open ?? ((tokens, idx, options, _env, self) => self.renderToken(tokens, idx, options))

markdown.renderer.rules.link_open = (tokens, idx, options, env, self) => {
  const token = tokens[idx]
  const targetIndex = token.attrIndex('target')
  const relIndex = token.attrIndex('rel')

  if (targetIndex < 0) {
    token.attrPush(['target', '_blank'])
  } else {
    token.attrs![targetIndex][1] = '_blank'
  }

  if (relIndex < 0) {
    token.attrPush(['rel', 'nofollow noopener noreferrer'])
  } else {
    token.attrs![relIndex][1] = 'nofollow noopener noreferrer'
  }

  return defaultLinkOpen(tokens, idx, options, env, self)
}

markdown.renderer.rules.image = (tokens, idx) => {
  return markdown.utils.escapeHtml(tokens[idx].content)
}

export function renderMarkdown(content: string) {
  return markdown.render(neutralizeRawHtml(content))
}

function neutralizeRawHtml(content: string) {
  return content
    .replace(/<\s*([a-z][\w-]*)(?:\s[^>]*)?>/gi, '<$1>')
    .replace(/<\s*\/\s*([a-z][\w-]*)\s*>/gi, '</$1>')
}
