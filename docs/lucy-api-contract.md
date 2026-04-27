# Decoy Club API Contract for Lucy

metadata:
  source_project: decoy-club
  intended_consumer: lucy
  generated_from:
    - backend/internal/http/router.go
    - backend/internal/*/handler.go
    - backend/internal/*/dto.go
    - frontend/src/api/client.ts
  api_prefix: /api/v1
  default_backend_base_url: http://localhost:8080
  default_api_base_url: http://localhost:8080/api/v1
  auth_scheme: bearer_jwt
  token_ttl: 24h
  time_format: RFC3339 string
  id_format: MongoDB ObjectID hex string
  content_type_default: application/json

## Global Rules

```yaml
authentication:
  header: Authorization
  value_format: "Bearer <token>"
  required_when_endpoint_auth_is: required
  optional_viewer_endpoints:
    behavior: If a valid bearer token is present, response includes viewer-specific fields such as liked_by_viewer and following.
    invalid_optional_token_behavior: ignored by optionalViewerID middleware

error_response:
  content_type: application/json
  schema:
    error: string
  examples:
    - status: 400
      body: { "error": "invalid request body" }
    - status: 401
      body: { "error": "missing authorization token" }
    - status: 401
      body: { "error": "invalid authorization token" }

empty_success_response:
  status: 204
  body: null

static_files:
  route: /uploads/*
  method: GET
  description: Serves uploaded files from configured upload_dir. API upload responses return public_url values under this route.
```

## Endpoint Index

```yaml
endpoints:
  - { method: GET, path: /api/v1/health, auth: none, operation_id: healthCheck }
  - { method: POST, path: /api/v1/auth/register, auth: none, operation_id: register }
  - { method: POST, path: /api/v1/auth/login, auth: none, operation_id: login }
  - { method: GET, path: /api/v1/users/{username}/profile, auth: optional, operation_id: getUserProfile }
  - { method: GET, path: /api/v1/users/{username}/posts, auth: optional, operation_id: listUserPosts }
  - { method: POST, path: /api/v1/users/{username}/follow, auth: required, operation_id: followUser }
  - { method: DELETE, path: /api/v1/users/{username}/follow, auth: required, operation_id: unfollowUser }
  - { method: GET, path: /api/v1/posts, auth: optional, operation_id: listPublicTimeline }
  - { method: POST, path: /api/v1/posts, auth: required, operation_id: createPost }
  - { method: GET, path: /api/v1/posts/{postId}, auth: optional, operation_id: getPost }
  - { method: DELETE, path: /api/v1/posts/{postId}, auth: required, operation_id: deletePost }
  - { method: POST, path: /api/v1/posts/{postId}/like, auth: required, operation_id: likePost }
  - { method: DELETE, path: /api/v1/posts/{postId}/like, auth: required, operation_id: unlikePost }
  - { method: GET, path: /api/v1/posts/{postId}/comments, auth: none, operation_id: listPostComments }
  - { method: POST, path: /api/v1/posts/{postId}/comments, auth: required, operation_id: createPostComment }
  - { method: POST, path: /api/v1/comments/{commentId}/replies, auth: required, operation_id: replyToComment }
  - { method: DELETE, path: /api/v1/comments/{commentId}, auth: required, operation_id: deleteComment }
  - { method: GET, path: /api/v1/topics/trending, auth: none, operation_id: listTrendingTopics }
  - { method: GET, path: /api/v1/notifications/unread-count, auth: required, operation_id: getNotificationUnreadCount }
  - { method: GET, path: /api/v1/notifications, auth: required, operation_id: listNotifications }
  - { method: GET, path: /api/v1/notifications/{notificationId}, auth: required, operation_id: getNotification }
  - { method: POST, path: /api/v1/notifications/read, auth: required, operation_id: markNotificationsRead }
  - { method: POST, path: /api/v1/uploads/images, auth: required, operation_id: uploadImage }
  - { method: GET, path: /api/v1/timeline/following, auth: required, operation_id: listFollowingTimeline }
```

## Schemas

```yaml
schemas:
  User:
    type: object
    fields:
      id: string
      username: string
      avatar_url: string
      bio: string
      post_count: integer
      reply_count: integer
      followers_count: integer
      following_count: integer
      received_like_count: integer
      given_like_count: integer
      created_at: string
      updated_at: string

  Profile:
    type: object
    fields:
      username: string
      avatar_url: string
      bio: string
      post_count: integer
      reply_count: integer
      followers_count: integer
      following_count: integer
      received_like_count: integer
      given_like_count: integer
      following: boolean

  Post:
    type: object
    fields:
      id: string
      author_id: string
      author_username: string optional
      content_markdown: string
      content_html: string
      embedded_images: string[]
      topics: string[]
      like_count: integer
      liked_by_viewer: boolean
      comment_count: integer
      created_at: string
      updated_at: string
      deleted_at: string optional

  Comment:
    type: object
    fields:
      id: string
      post_id: string
      author_id: string
      author_username: string optional
      content_markdown: string
      content_html: string
      parent_comment_id: string optional
      reply_to_user_id: string optional
      is_deleted: boolean
      created_at: string
      updated_at: string
      deleted_at: string optional
      replies: Comment[]
    notes:
      - Deleted comments are returned with content_markdown set to "内容已删除" and content_html set to "<p>内容已删除</p>".
      - Only one level of replies is supported by service validation.

  Topic:
    type: object
    fields:
      id: string
      name: string
      post_count: integer
      updated_at: string
      last_post_at: string

  NotificationType:
    enum:
      - post_liked
      - post_commented
      - comment_replied
      - user_mentioned

  NotificationListItem:
    type: object
    fields:
      id: string
      type: NotificationType
      is_read: boolean
      created_at: string
      actor:
        id: string
        username: string
      post:
        id: string
        author_id: string
        author_username: string
        content_markdown: string optional
        content_preview: string
      comment:
        id: string
        author_id: string
        author_username: string
        content_markdown: string optional
        content_preview: string
        created_at: string
      parent_comment:
        id: string
        author_id: string
        author_username: string
        content_markdown: string optional
        content_preview: string
        created_at: string
      snapshot_flags:
        post_deleted: boolean
        comment_deleted: boolean
        parent_comment_deleted: boolean
    notes:
      - List endpoint omits full content_markdown in post/comment snapshots.
      - Detail endpoint includes full content_markdown in available snapshots.

  UploadedFile:
    type: object
    fields:
      id: string
      file_name: string
      mime_type: string
      size: integer
      public_url: string
      created_at: string
```

## Detailed Endpoints

```yaml
- operation_id: healthCheck
  method: GET
  path: /api/v1/health
  auth: none
  response:
    200:
      status: string
      example: { "status": "ok" }

- operation_id: register
  method: POST
  path: /api/v1/auth/register
  auth: none
  request:
    body:
      username: string
      password: string
  response:
    201:
      user: User
  errors:
    400: invalid request body | username already taken | other validation/storage errors

- operation_id: login
  method: POST
  path: /api/v1/auth/login
  auth: none
  request:
    body:
      username: string
      password: string
  response:
    200:
      token: string
      user: User
  errors:
    400: invalid request body | other login errors
    401: invalid credentials | user not found

- operation_id: getUserProfile
  method: GET
  path: /api/v1/users/{username}/profile
  auth: optional
  path_params:
    username: string
  response:
    200:
      profile: Profile
  errors:
    404: user not found

- operation_id: listUserPosts
  method: GET
  path: /api/v1/users/{username}/posts
  auth: optional
  path_params:
    username: string
  response:
    200:
      posts: Post[]
  implementation_notes:
    - Current backend uses a fixed limit of 20 and exposes no pagination parameters here.
  errors:
    404: user not found

- operation_id: followUser
  method: POST
  path: /api/v1/users/{username}/follow
  auth: required
  path_params:
    username: string
  response:
    200:
      profile: Profile
  errors:
    400: cannot follow self
    401: missing authorization token | invalid authorization token | missing viewer identity
    404: user not found

- operation_id: unfollowUser
  method: DELETE
  path: /api/v1/users/{username}/follow
  auth: required
  path_params:
    username: string
  response:
    200:
      profile: Profile
  errors:
    400: cannot follow self
    401: missing authorization token | invalid authorization token | missing viewer identity
    404: user not found

- operation_id: listPublicTimeline
  method: GET
  path: /api/v1/posts
  auth: optional
  query:
    limit:
      type: integer
      default: 20
      rule: positive integer only; invalid or non-positive values fall back to default
    before:
      type: string
      format: RFC3339
      required: false
    before_id:
      type: string
      required: false
      rule: requires before when provided
  response:
    200:
      posts: Post[]
  errors:
    400: before_id requires before | other listing errors

- operation_id: createPost
  method: POST
  path: /api/v1/posts
  auth: required
  request:
    body:
      content_markdown: string
      embedded_images: string[]
  response:
    201:
      post: Post
  side_effects:
    - Extracts topics from content_markdown.
    - Renders content_html from content_markdown.
    - Creates user_mentioned notifications for valid @username mentions.
  errors:
    400: invalid request body | other create errors
    401: missing authorization token | invalid authorization token | invalid author id | missing viewer identity

- operation_id: getPost
  method: GET
  path: /api/v1/posts/{postId}
  auth: optional
  path_params:
    postId: string
  response:
    200:
      post: Post
  errors:
    404: post not found

- operation_id: deletePost
  method: DELETE
  path: /api/v1/posts/{postId}
  auth: required
  path_params:
    postId: string
  response:
    204: null
  behavior:
    - Soft deletes the post.
    - Only the author can delete through repository checks.
  errors:
    401: missing authorization token | invalid authorization token | missing viewer identity
    404: post not found

- operation_id: likePost
  method: POST
  path: /api/v1/posts/{postId}/like
  auth: required
  path_params:
    postId: string
  response:
    204: null
  side_effects:
    - Creates post_liked notification when the like state changes and actor is not the post author.
  errors:
    401: missing authorization token | invalid authorization token | missing viewer identity
    404: post not found

- operation_id: unlikePost
  method: DELETE
  path: /api/v1/posts/{postId}/like
  auth: required
  path_params:
    postId: string
  response:
    204: null
  errors:
    401: missing authorization token | invalid authorization token | missing viewer identity
    404: post not found

- operation_id: listPostComments
  method: GET
  path: /api/v1/posts/{postId}/comments
  auth: none
  path_params:
    postId: string
  response:
    200:
      comments: Comment[]
  response_shape:
    - comments is a tree of root comments.
    - replies are nested in each root comment's replies array.
  errors:
    400: invalid post id | other listing errors

- operation_id: createPostComment
  method: POST
  path: /api/v1/posts/{postId}/comments
  auth: required
  path_params:
    postId: string
  request:
    body:
      content_markdown: string
  response:
    201:
      comment: Comment
  side_effects:
    - Renders content_html from content_markdown.
    - Creates post_commented notification when commenting on another user's post.
    - Creates user_mentioned notifications for valid @username mentions.
  errors:
    400: invalid request body | invalid post id | invalid comment author id | other create errors
    401: missing authorization token | invalid authorization token | missing viewer identity

- operation_id: replyToComment
  method: POST
  path: /api/v1/comments/{commentId}/replies
  auth: required
  path_params:
    commentId: string
  request:
    body:
      content_markdown: string
  response:
    201:
      comment: Comment
  behavior:
    - Replies target root comments only.
    - The created reply has parent_comment_id set to commentId and reply_to_user_id set to the parent comment author.
  side_effects:
    - Creates comment_replied notification when replying to another user's comment.
    - Creates user_mentioned notifications for valid @username mentions.
  errors:
    400: invalid request body | only one level of replies is supported | invalid comment author id | other reply errors
    401: missing authorization token | invalid authorization token | missing viewer identity
    404: comment not found

- operation_id: deleteComment
  method: DELETE
  path: /api/v1/comments/{commentId}
  auth: required
  path_params:
    commentId: string
  response:
    204: null
  behavior:
    - Soft deletes the comment.
    - Only the author can delete through repository checks.
  errors:
    401: missing authorization token | invalid authorization token | missing viewer identity
    404: comment not found

- operation_id: listTrendingTopics
  method: GET
  path: /api/v1/topics/trending
  auth: none
  query:
    limit:
      type: integer
      default: 10
      rule: positive integer only; invalid or non-positive values fall back to default
  response:
    200:
      topics: Topic[]
  errors:
    400: listing errors

- operation_id: getNotificationUnreadCount
  method: GET
  path: /api/v1/notifications/unread-count
  auth: required
  response:
    200:
      total: integer
      by_type:
        type: object
        additional_properties: integer
        keys: NotificationType
  errors:
    401: missing authorization token | invalid authorization token | missing viewer identity

- operation_id: listNotifications
  method: GET
  path: /api/v1/notifications
  auth: required
  query:
    page:
      type: integer
      default: 1
      rule: positive integer only; invalid or non-positive values fall back to default
    page_size:
      type: integer
      default: 20
      rule: positive integer only; invalid or non-positive values fall back to default
    status:
      type: string
      enum: [unread]
      required: false
      behavior: Only status=unread enables unread-only filtering; any other value means all statuses.
    type:
      type: string
      enum: [likes, comments, mentions]
      required: false
      mapping:
        likes: [post_liked]
        comments: [post_commented, comment_replied]
        mentions: [user_mentioned]
  response:
    200:
      notifications: NotificationListItem[]
  errors:
    401: missing authorization token | invalid authorization token | missing viewer identity

- operation_id: getNotification
  method: GET
  path: /api/v1/notifications/{notificationId}
  auth: required
  path_params:
    notificationId: string
  response:
    200:
      notification: NotificationListItem
  behavior:
    - The notification must belong to the authenticated viewer.
    - Full content_markdown is included for available post/comment snapshots.
  errors:
    401: missing authorization token | invalid authorization token | missing viewer identity
    404: notification not found

- operation_id: markNotificationsRead
  method: POST
  path: /api/v1/notifications/read
  auth: required
  request:
    body:
      notification_ids: string[]
  response:
    204: null
  behavior:
    - Marks only the authenticated viewer's matching notifications as read.
  errors:
    400: invalid request body | other mark-read errors
    401: missing authorization token | invalid authorization token | missing viewer identity

- operation_id: uploadImage
  method: POST
  path: /api/v1/uploads/images
  auth: required
  request:
    content_type: multipart/form-data
    form_fields:
      file:
        type: file
        required: true
        constraints:
          - Content-Type must start with image/
          - Size must be non-zero
  response:
    201:
      file: UploadedFile
  errors:
    400: missing file | invalid owner id | only image uploads are supported | file is empty
    401: missing authorization token | invalid authorization token | missing viewer identity
    500: filesystem/storage errors

- operation_id: listFollowingTimeline
  method: GET
  path: /api/v1/timeline/following
  auth: required
  query:
    page:
      type: integer
      default: 1
      rule: positive integer only; invalid or non-positive values fall back to default
    size:
      type: integer
      default: 20
      rule: positive integer only; invalid or non-positive values fall back to default
  response:
    200:
      posts: Post[]
  errors:
    401: missing authorization token | invalid authorization token | missing viewer identity
```

## Lucy Client Notes

```yaml
recommended_client_defaults:
  api_base_url_env: VITE_API_BASE_URL
  fallback_api_base_url: http://localhost:8080/api/v1
  request_headers:
    json:
      Content-Type: application/json
    authenticated:
      Authorization: "Bearer <token>"
  response_handling:
    - If HTTP status is 204, treat response body as undefined/null.
    - If HTTP status is not 2xx, parse JSON and use error as the user-visible message.
    - Normalize posts embedded_images and topics to empty arrays when the backend returns null or missing values.

frontend_function_map_from_decoy_club:
  register: "POST /auth/register"
  login: "POST /auth/login"
  fetchPublicPosts: "GET /posts"
  fetchFollowingPosts: "GET /timeline/following"
  fetchPost: "GET /posts/{postId}"
  fetchPostComments: "GET /posts/{postId}/comments"
  createPost: "POST /posts"
  createComment: "POST /posts/{postId}/comments"
  replyToComment: "POST /comments/{commentId}/replies"
  likePost: "POST /posts/{postId}/like"
  unlikePost: "DELETE /posts/{postId}/like"
  uploadImage: "POST /uploads/images"
  fetchUnreadCount: "GET /notifications/unread-count"
  fetchNotifications: "GET /notifications"
  fetchTrendingTopics: "GET /topics/trending"
  fetchNotificationDetail: "GET /notifications/{notificationId}"
  markNotificationsRead: "POST /notifications/read"
  fetchProfile: "GET /users/{username}/profile"
  fetchUserPosts: "GET /users/{username}/posts"
  followUser: "POST /users/{username}/follow"
  unfollowUser: "DELETE /users/{username}/follow"
```
