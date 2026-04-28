import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'

const postDetailSource = readFileSync(new URL('./PostDetailView.vue', import.meta.url), 'utf8')
const profileSource = readFileSync(new URL('./ProfileView.vue', import.meta.url), 'utf8')

describe('markdown rendering usage', () => {
  it('renders post detail markdown instead of stored html', () => {
    expect(postDetailSource).toContain(':content="post.content_markdown"')
    expect(postDetailSource).not.toContain('v-html="post.content_html"')
  })

  it('renders activity comment markdown instead of stored html', () => {
    expect(profileSource).toContain(':content="activity.comment.content_markdown"')
    expect(profileSource).not.toContain('v-html="activity.comment.content_html"')
  })
})
