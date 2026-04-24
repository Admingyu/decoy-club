import { describe, expect, it } from 'vitest'
import { normalizePost, type ApiPost } from './client'

const basePost: ApiPost = {
  id: 'post-1',
  author_id: 'author-1',
  content_markdown: 'hello',
  content_html: '<p>hello</p>',
  embedded_images: [],
  like_count: 0,
  comment_count: 0,
  created_at: '2026-04-24T00:00:00Z',
  updated_at: '2026-04-24T00:00:00Z',
}

describe('normalizePost', () => {
  it('normalizes null embedded image lists from existing API data', () => {
    const post = { ...basePost, embedded_images: null } as unknown as ApiPost

    expect(normalizePost(post).embedded_images).toEqual([])
  })
})
