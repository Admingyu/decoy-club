import { describe, expect, it } from 'vitest'
import {
  normalizeActivityResponse,
  normalizeComment,
  normalizePost,
  normalizeUserSearchResponse,
  type ApiComment,
  type ApiPost,
} from './client'

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

describe('normalizeActivityResponse', () => {
  it('normalizes null activity lists from existing API data', () => {
    expect(normalizeActivityResponse({ activities: null }).activities).toEqual([])
  })
})

describe('normalizeUserSearchResponse', () => {
  it('normalizes null user search lists from existing API data', () => {
    expect(normalizeUserSearchResponse({ users: null }).users).toEqual([])
  })
})

describe('normalizeComment', () => {
  it('normalizes missing comment like state and reply lists from existing API data', () => {
    const comment = {
      id: 'comment-1',
      post_id: 'post-1',
      author_id: 'author-1',
      content_markdown: 'hello',
      content_html: '<p>hello</p>',
      is_deleted: false,
      created_at: '2026-04-24T00:00:00Z',
      updated_at: '2026-04-24T00:00:00Z',
      replies: null,
    } as unknown as ApiComment

    expect(normalizeComment(comment)).toMatchObject({
      like_count: 0,
      liked_by_viewer: false,
      replies: [],
    })
  })
})
