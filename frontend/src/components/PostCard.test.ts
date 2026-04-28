import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'

const source = readFileSync(new URL('./PostCard.vue', import.meta.url), 'utf8')

describe('PostCard markup', () => {
  it('renders markdown content outside of RouterLink navigation', () => {
    const contentIndex = source.indexOf(':content="post.content_markdown"')
    const openingRouterLink = source.lastIndexOf('<RouterLink', contentIndex)
    const closingRouterLink = source.indexOf('</RouterLink>', openingRouterLink)

    expect(contentIndex).toBeGreaterThan(-1)
    expect(openingRouterLink < contentIndex && contentIndex < closingRouterLink).toBe(false)
  })

  it('does not render stored post html directly', () => {
    expect(source).not.toContain('v-html="post.content_html"')
  })
})
