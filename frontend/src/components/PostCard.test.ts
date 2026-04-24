import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'

const source = readFileSync(new URL('./PostCard.vue', import.meta.url), 'utf8')

describe('PostCard markup', () => {
  it('does not wrap raw HTML post content in a RouterLink', () => {
    const contentIndex = source.indexOf('v-html="post.content_html"')
    const openingRouterLink = source.lastIndexOf('<RouterLink', contentIndex)
    const closingRouterLink = source.indexOf('</RouterLink>', openingRouterLink)

    expect(contentIndex).toBeGreaterThan(-1)
    expect(openingRouterLink < contentIndex && contentIndex < closingRouterLink).toBe(false)
  })
})
