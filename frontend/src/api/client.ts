const API_BASE_URL = (import.meta.env.VITE_API_BASE_URL as string | undefined)?.replace(/\/$/, '') || '/api/v1'

export type ApiPost = {
  id: string
  author_id: string
  author_username?: string
  content_markdown: string
  content_html: string
  embedded_images: string[]
  topics: string[]
  like_count: number
  liked_by_viewer?: boolean
  comment_count: number
  created_at: string
  updated_at: string
}

export function normalizePost(post: ApiPost): ApiPost {
  return {
    ...post,
    embedded_images: Array.isArray(post.embedded_images) ? post.embedded_images : [],
    topics: Array.isArray(post.topics) ? post.topics : [],
  }
}

function normalizePostResponse(response: { post: ApiPost }) {
  return {
    ...response,
    post: normalizePost(response.post),
  }
}

function normalizePostsResponse(response: { posts: ApiPost[] | null }) {
  return {
    ...response,
    posts: Array.isArray(response.posts) ? response.posts.map(normalizePost) : [],
  }
}

export const normalizePostSearchResponse = normalizePostsResponse

export type ApiComment = {
  id: string
  post_id: string
  author_id: string
  author_username?: string
  content_markdown: string
  content_html: string
  parent_comment_id?: string
  reply_to_user_id?: string
  like_count: number
  liked_by_viewer?: boolean
  is_deleted: boolean
  created_at: string
  updated_at: string
  replies: ApiComment[]
}

export function normalizeComment(comment: ApiComment): ApiComment {
  return {
    ...comment,
    like_count: typeof comment.like_count === 'number' ? comment.like_count : 0,
    liked_by_viewer: Boolean(comment.liked_by_viewer),
    replies: Array.isArray(comment.replies) ? comment.replies.map(normalizeComment) : [],
  }
}

function normalizeCommentResponse(response: { comment: ApiComment }) {
  return {
    ...response,
    comment: normalizeComment(response.comment),
  }
}

function normalizeCommentsResponse(response: { comments: ApiComment[] | null }) {
  return {
    ...response,
    comments: Array.isArray(response.comments) ? response.comments.map(normalizeComment) : [],
  }
}

export type ApiUser = {
  id: string
  username: string
  avatar_url?: string
  bio?: string
}

export type ApiPostLiker = {
  id: string
  username: string
  avatar_url?: string
}

export type ApiProfile = {
  username: string
  avatar_url?: string
  bio?: string
  status_text?: string
  status_preset?: string
  post_count: number
  reply_count: number
  followers_count: number
  following_count: number
  received_like_count: number
  given_like_count: number
  following: boolean
}

export function normalizeUserSearchResponse(response: { users: ApiProfile[] | null }) {
  return {
    ...response,
    users: Array.isArray(response.users) ? response.users : [],
  }
}

export type ActivityType = 'views' | 'likes' | 'comments'

export type ApiActivityCounts = {
  views: number
  likes: number
  comments: number
}

export type ApiActivityItem = {
  id: string
  type: 'view' | 'like' | 'comment'
  count: number
  created_at: string
  updated_at: string
  post?: ApiPost
  comment?: ApiComment
}

export function normalizeActivityResponse(response: { activities: ApiActivityItem[] | null }) {
  return {
    ...response,
    activities: Array.isArray(response.activities)
      ? response.activities.map((activity) => ({
        ...activity,
        post: activity.post ? normalizePost(activity.post) : undefined,
      }))
      : [],
  }
}

export type AuthResponse = {
  token: string
  user: ApiUser
}

export type NotificationListItem = {
  id: string
  type: 'post_liked' | 'post_commented' | 'comment_replied' | 'user_mentioned'
  is_read: boolean
  created_at: string
  actor?: { id: string; username: string }
  post?: { id: string; author_id: string; author_username: string; content_markdown?: string; content_preview: string }
  comment?: { id: string; author_id: string; author_username: string; content_markdown?: string; content_preview: string; created_at: string }
  parent_comment?: { id: string; author_id: string; author_username: string; content_markdown?: string; content_preview: string; created_at: string }
}

export type NotificationFilter = 'all' | 'likes' | 'comments' | 'mentions'

export type Topic = {
  id: string
  name: string
  post_count: number
  updated_at: string
  last_post_at: string
}

async function request<T>(path: string, init: RequestInit = {}, token?: string): Promise<T> {
  const headers = new Headers(init.headers)
  if (!headers.has('Content-Type') && !(init.body instanceof FormData)) {
    headers.set('Content-Type', 'application/json')
  }
  if (token) {
    headers.set('Authorization', `Bearer ${token}`)
  }

  const response = await fetch(`${API_BASE_URL}${path}`, { ...init, headers })
  if (!response.ok) {
    let message = `Request failed: ${response.status}`
    try {
      const payload = await response.json()
      if (typeof payload.error === 'string') {
        message = payload.error
      }
    } catch {
      // ignore json parse failures
    }
    throw new Error(message)
  }

  if (response.status === 204) {
    return undefined as T
  }
  return response.json() as Promise<T>
}

export function register(username: string, password: string) {
  return request<{ user: ApiUser }>('/auth/register', {
    method: 'POST',
    body: JSON.stringify({ username, password }),
  })
}

export function login(username: string, password: string) {
  return request<AuthResponse>('/auth/login', {
    method: 'POST',
    body: JSON.stringify({ username, password }),
  })
}

export function fetchPublicPosts() {
  return request<{ posts: ApiPost[] | null }>('/posts').then(normalizePostsResponse)
}

export function fetchFollowingPosts(token: string) {
  return request<{ posts: ApiPost[] | null }>('/timeline/following', {}, token).then(normalizePostsResponse)
}

export function fetchPost(postId: string, token?: string) {
  return request<{ post: ApiPost }>(`/posts/${postId}`, {}, token).then(normalizePostResponse)
}

export function fetchPostLikers(postId: string, limit = 6) {
  return request<{ likers: ApiPostLiker[] | null; total: number }>(`/posts/${postId}/likes?limit=${limit}`)
    .then((response) => ({
      ...response,
      likers: Array.isArray(response.likers) ? response.likers : [],
    }))
}

export function fetchPostComments(postId: string, token?: string) {
  return request<{ comments: ApiComment[] | null }>(`/posts/${postId}/comments`, {}, token).then(normalizeCommentsResponse)
}

export function createPost(token: string, contentMarkdown: string, embeddedImages: string[]) {
  return request<{ post: ApiPost }>('/posts', {
    method: 'POST',
    body: JSON.stringify({
      content_markdown: contentMarkdown,
      embedded_images: embeddedImages,
    }),
  }, token).then(normalizePostResponse)
}

export function createComment(token: string, postId: string, contentMarkdown: string) {
  return request<{ comment: ApiComment }>(`/posts/${postId}/comments`, {
    method: 'POST',
    body: JSON.stringify({ content_markdown: contentMarkdown }),
  }, token).then(normalizeCommentResponse)
}

export function replyToComment(token: string, commentId: string, contentMarkdown: string) {
  return request<{ comment: ApiComment }>(`/comments/${commentId}/replies`, {
    method: 'POST',
    body: JSON.stringify({ content_markdown: contentMarkdown }),
  }, token).then(normalizeCommentResponse)
}

export function likePost(token: string, postId: string) {
  return request<void>(`/posts/${postId}/like`, { method: 'POST' }, token)
}

export function unlikePost(token: string, postId: string) {
  return request<void>(`/posts/${postId}/like`, { method: 'DELETE' }, token)
}

export function likeComment(token: string, commentId: string) {
  return request<void>(`/comments/${commentId}/like`, { method: 'POST' }, token)
}

export function unlikeComment(token: string, commentId: string) {
  return request<void>(`/comments/${commentId}/like`, { method: 'DELETE' }, token)
}

export function uploadImage(token: string, file: File) {
  const formData = new FormData()
  formData.append('file', file)

  return request<{ file: { id: string; public_url: string; file_name: string } }>('/uploads/images', {
    method: 'POST',
    body: formData,
  }, token)
}

export function fetchUnreadCount(token: string) {
  return request<{ total: number; by_type: Record<string, number> }>('/notifications/unread-count', {}, token)
}

export function fetchNotifications(token: string, unreadOnly = false, filter: NotificationFilter = 'all') {
  const params = new URLSearchParams()
  if (unreadOnly) {
    params.set('status', 'unread')
  }
  if (filter !== 'all') {
    params.set('type', filter)
  }
  const query = params.toString() ? `?${params.toString()}` : ''
  return request<{ notifications: NotificationListItem[] }>(`/notifications${query}`, {}, token)
}

export function fetchTrendingTopics(limit = 10) {
  return request<{ topics: Topic[] }>(`/topics/trending?limit=${limit}`)
}

export function fetchNotificationDetail(token: string, notificationId: string) {
  return request<{ notification: NotificationListItem }>(`/notifications/${notificationId}`, {}, token)
}

export function markNotificationsRead(token: string, notificationIds: string[]) {
  return request<void>('/notifications/read', {
    method: 'POST',
    body: JSON.stringify({ notification_ids: notificationIds }),
  }, token)
}

export function fetchProfile(username: string, token?: string) {
  return request<{ profile: ApiProfile }>(`/users/${encodeURIComponent(username)}/profile`, {}, token)
}

export function searchUsers(query: string, token?: string) {
  const params = new URLSearchParams()
  params.set('q', query)
  return request<{ users: ApiProfile[] | null }>(`/users/search?${params.toString()}`, {}, token)
    .then(normalizeUserSearchResponse)
}

export function searchPosts(query: string, token?: string) {
  const params = new URLSearchParams()
  params.set('q', query)
  return request<{ posts: ApiPost[] | null }>(`/posts/search?${params.toString()}`, {}, token)
    .then(normalizePostSearchResponse)
}

export function updateMyStatus(token: string, statusText: string, statusPreset: string) {
  return request<{ profile: ApiProfile }>('/users/me/status', {
    method: 'PUT',
    body: JSON.stringify({
      status_text: statusText,
      status_preset: statusPreset,
    }),
  }, token)
}

export function fetchUserPosts(username: string, token?: string) {
  return request<{ posts: ApiPost[] | null }>(`/users/${encodeURIComponent(username)}/posts`, {}, token).then(normalizePostsResponse)
}

export function fetchActivityCounts(username: string, token?: string) {
  return request<{ counts: ApiActivityCounts }>(`/users/${encodeURIComponent(username)}/activity-counts`, {}, token)
}

export function fetchUserActivity(token: string, username: string, type: ActivityType) {
  return request<{ activities: ApiActivityItem[] | null }>(
    `/users/${encodeURIComponent(username)}/activity?type=${encodeURIComponent(type)}`,
    {},
    token,
  ).then(normalizeActivityResponse)
}

export function followUser(token: string, username: string) {
  return request<{ profile: ApiProfile }>(`/users/${encodeURIComponent(username)}/follow`, { method: 'POST' }, token)
}

export function unfollowUser(token: string, username: string) {
  return request<{ profile: ApiProfile }>(`/users/${encodeURIComponent(username)}/follow`, { method: 'DELETE' }, token)
}

export { API_BASE_URL }
