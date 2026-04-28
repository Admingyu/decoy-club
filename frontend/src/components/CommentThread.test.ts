import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'

const source = readFileSync(new URL('./CommentThread.vue', import.meta.url), 'utf8')

describe('CommentThread actions', () => {
  it('uses post action styling for comment like and reply actions', () => {
    expect(source).toContain('post-action post-action--button')
    expect(source).not.toContain('ghost-button')
  })

  it('renders markdown instead of stored comment html', () => {
    expect(source).toContain(':content="comment.content_markdown"')
    expect(source).not.toContain('v-html="comment.content_html"')
  })
})
