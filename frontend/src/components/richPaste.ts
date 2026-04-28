const SAFE_LINK_PROTOCOLS = new Set(['http:', 'https:', 'mailto:'])

type ClipboardFileItem = {
  kind: string
  type: string
  getAsFile: () => File | null
}

type ClipboardImageSource = {
  files?: ArrayLike<File>
  items?: ArrayLike<ClipboardFileItem>
} | null | undefined

export function convertRichHtmlToMarkdown(html: string) {
  let markdown = html
    .replace(/\r\n/g, '\n')
    .replace(/<!--[\s\S]*?-->/g, '')
    .replace(/<script[\s\S]*?<\/script>/gi, '')
    .replace(/<style[\s\S]*?<\/style>/gi, '')
    .replace(/<br\s*\/?>/gi, '\n')

  markdown = markdown.replace(/<h([1-6])\b[^>]*>([\s\S]*?)<\/h\1>/gi, (_match, level: string, content: string) => {
    return `\n\n${'#'.repeat(Number(level))} ${convertInlineHtml(content)}\n\n`
  })

  markdown = markdown.replace(/<li\b[^>]*>([\s\S]*?)<\/li>/gi, (_match, content: string) => {
    return `\n- ${convertInlineHtml(content)}`
  })

  markdown = markdown.replace(/<\/?(ul|ol)\b[^>]*>/gi, '\n')
  markdown = markdown.replace(/<blockquote\b[^>]*>([\s\S]*?)<\/blockquote>/gi, (_match, content: string) => {
    const quote = convertRichHtmlToMarkdown(content)
      .split('\n')
      .map((line) => line.trim())
      .filter(Boolean)
      .map((line) => `> ${line}`)
      .join('\n')
    return quote ? `\n\n${quote}\n\n` : ''
  })

  markdown = markdown.replace(/<(p|div)\b[^>]*>([\s\S]*?)<\/\1>/gi, (_match, _tag: string, content: string) => {
    const line = convertInlineHtml(content)
    return line ? `\n\n${line}\n\n` : '\n\n'
  })

  markdown = convertInlineHtml(markdown)
  return normalizeMarkdown(markdown)
}

export function getClipboardImageFiles(source: ClipboardImageSource) {
  const images: File[] = []
  const seen = new Set<File>()

  if (source?.files) {
    for (let index = 0; index < source.files.length; index++) {
      const file = source.files[index]
      if (file?.type.startsWith('image/') && !seen.has(file)) {
        images.push(file)
        seen.add(file)
      }
    }
  }

  if (source?.items) {
    for (let index = 0; index < source.items.length; index++) {
      const item = source.items[index]
      if (item?.kind !== 'file' || !item.type.startsWith('image/')) {
        continue
      }
      const file = item.getAsFile()
      if (file && !seen.has(file)) {
        images.push(file)
        seen.add(file)
      }
    }
  }

  return images
}

function convertInlineHtml(html: string): string {
  let text = html
    .replace(/<code\b[^>]*>([\s\S]*?)<\/code>/gi, (_match, content: string) => {
      return `\`${decodeHtml(stripTags(content)).replace(/`/g, '\\`')}\``
    })
    .replace(/<a\b[^>]*href=(["'])(.*?)\1[^>]*>([\s\S]*?)<\/a>/gi, (_match, _quote: string, href: string, content: string) => {
      const label = convertInlineHtml(content)
      return isSafeHref(href) ? `[${label}](${decodeHtml(href)})` : label
    })
    .replace(/<(strong|b)\b[^>]*>([\s\S]*?)<\/\1>/gi, (_match, _tag: string, content: string) => {
      return `**${convertInlineHtml(content)}**`
    })
    .replace(/<(em|i)\b[^>]*>([\s\S]*?)<\/\1>/gi, (_match, _tag: string, content: string) => {
      return `*${convertInlineHtml(content)}*`
    })
    .replace(/<img\b[^>]*alt=(["'])(.*?)\1[^>]*>/gi, (_match, _quote: string, alt: string) => {
      return decodeHtml(alt)
    })

  text = stripTags(text)
  return decodeHtml(text)
    .replace(/\u00a0/g, ' ')
    .replace(/[ \t\f\v]+/g, ' ')
    .replace(/[ \t]*\n[ \t]*/g, '\n')
    .trim()
}

function stripTags(value: string) {
  return value.replace(/<[^>]*>/g, '')
}

function decodeHtml(value: string) {
  return value.replace(/&(#x?[0-9a-f]+|[a-z]+);/gi, (entity, body: string) => {
    if (body[0] === '#') {
      const isHex = body[1]?.toLowerCase() === 'x'
      const codePoint = Number.parseInt(body.slice(isHex ? 2 : 1), isHex ? 16 : 10)
      return Number.isFinite(codePoint) ? String.fromCodePoint(codePoint) : entity
    }

    const named: Record<string, string> = {
      amp: '&',
      apos: "'",
      gt: '>',
      lt: '<',
      nbsp: ' ',
      quot: '"',
    }
    return named[body.toLowerCase()] ?? entity
  })
}

function isSafeHref(href: string) {
  try {
    const url = new URL(decodeHtml(href), 'https://decoy.local')
    return SAFE_LINK_PROTOCOLS.has(url.protocol)
  } catch {
    return false
  }
}

function normalizeMarkdown(markdown: string) {
  return markdown
    .split('\n')
    .map((line) => line.trimEnd())
    .join('\n')
    .replace(/\n{3,}/g, '\n\n')
    .trim()
}
