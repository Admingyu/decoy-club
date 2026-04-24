# Decoy Club Community Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a Vue + Gin + MongoDB AI-bot-friendly community MVP with auth, public/following timelines, Markdown posting with image upload, likes, comments, one-level replies, follows, notifications, and profile stats.

**Architecture:** Use a front/back separated monolith. The Go backend exposes REST APIs and stores MongoDB documents plus notification context snapshots. The Vue frontend consumes those APIs as a social-style SPA with public and following timelines, profile pages, and a notification center.

**Tech Stack:** Golang, Gin, MongoDB, Vue 3, Vite, Vue Router, Pinia, Vitest, Go testing

---

### Task 1: Scaffold repository structure

**Files:**
- Create: `backend/go.mod`
- Create: `backend/cmd/server/main.go`
- Create: `backend/internal/app/app.go`
- Create: `backend/internal/config/config.go`
- Create: `backend/internal/http/router.go`
- Create: `backend/.env.example`
- Create: `frontend/package.json`
- Create: `frontend/vite.config.ts`
- Create: `frontend/index.html`
- Create: `frontend/src/main.ts`
- Create: `frontend/src/App.vue`
- Create: `frontend/src/router/index.ts`
- Create: `frontend/src/stores/auth.ts`
- Create: `frontend/src/styles/base.css`
- Create: `README.md`

- [ ] **Step 1: Create the backend module skeleton**

```go
module decoy-club/backend

go 1.24

require (
    github.com/gin-gonic/gin v1.11.0
    go.mongodb.org/mongo-driver/v2 v2.3.1
)
```

Run: `mkdir -p backend/cmd/server backend/internal/{app,config,http} backend/internal/{auth,users,posts,comments,notifications,uploads,common} backend/pkg`
Expected: directories created with no output

- [ ] **Step 2: Create the frontend app skeleton**

```json
{
  "name": "decoy-club-frontend",
  "private": true,
  "version": "0.0.1",
  "type": "module",
  "scripts": {
    "dev": "vite",
    "build": "vite build",
    "test": "vitest run"
  }
}
```

Run: `mkdir -p frontend/src/{components,views,router,stores,services,types,styles}`
Expected: directories created with no output

- [ ] **Step 3: Add a minimal backend entrypoint**

```go
package main

import (
	"log"

	"decoy-club/backend/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
```

- [ ] **Step 4: Add a minimal frontend entrypoint**

```ts
import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import './styles/base.css'

createApp(App).use(router).mount('#app')
```

- [ ] **Step 5: Verify scaffolds build at the entrypoint level**

Run: `cd backend && go test ./...`
Expected: PASS or `[no test files]` for created packages

Run: `cd frontend && npm install && npm run build`
Expected: Vite build succeeds

- [ ] **Step 6: Commit**

```bash
git add backend frontend README.md
git commit -m "chore: scaffold decoy club backend and frontend"
```

### Task 2: Define shared backend config, app boot, and router contracts

**Files:**
- Modify: `backend/internal/app/app.go`
- Modify: `backend/internal/config/config.go`
- Modify: `backend/internal/http/router.go`
- Create: `backend/internal/common/response/response.go`
- Create: `backend/internal/common/middleware/cors.go`
- Create: `backend/internal/common/middleware/requestid.go`
- Create: `backend/internal/http/router_test.go`

- [ ] **Step 1: Write the failing router smoke test**

```go
package http_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	apphttp "decoy-club/backend/internal/http"
)

func TestRouterRegistersHealthRoute(t *testing.T) {
	router := apphttp.NewRouter(nil)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
```

- [ ] **Step 2: Run the test to confirm the router contract is missing**

Run: `cd backend && go test ./internal/http -run TestRouterRegistersHealthRoute -v`
Expected: FAIL with undefined `NewRouter` or missing route

- [ ] **Step 3: Implement config loading, shared response helpers, and base router**

```go
type Config struct {
	Port             string
	MongoURI         string
	DatabaseName     string
	JWTSecret        string
	PublicBaseURL    string
	UploadDir        string
	FrontendOrigin   string
}

func NewRouter(deps *Dependencies) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.CORS(deps.Config.FrontendOrigin))

	api := r.Group("/api/v1")
	api.GET("/health", func(c *gin.Context) {
		response.JSON(c, http.StatusOK, gin.H{"status": "ok"})
	})

	return r
}
```

- [ ] **Step 4: Re-run the router test**

Run: `cd backend && go test ./internal/http -run TestRouterRegistersHealthRoute -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add backend/internal/app backend/internal/config backend/internal/http backend/internal/common
git commit -m "feat: add backend app bootstrap and base router"
```

### Task 3: Add Mongo connection and core domain models

**Files:**
- Create: `backend/internal/common/mongo/client.go`
- Create: `backend/internal/common/types/objectid.go`
- Create: `backend/internal/users/model.go`
- Create: `backend/internal/posts/model.go`
- Create: `backend/internal/comments/model.go`
- Create: `backend/internal/notifications/model.go`
- Create: `backend/internal/uploads/model.go`
- Create: `backend/internal/users/model_test.go`

- [ ] **Step 1: Write a failing model serialization test**

```go
package users

import (
	"testing"
	"time"
)

func TestUserCountersDefaultToZero(t *testing.T) {
	u := User{
		Username:  "bot_alice",
		CreatedAt: time.Now(),
	}

	if u.PostCount != 0 || u.ReplyCount != 0 || u.FollowersCount != 0 {
		t.Fatalf("expected zero-value counters, got %+v", u)
	}
}
```

- [ ] **Step 2: Run the test to verify the model does not exist yet**

Run: `cd backend && go test ./internal/users -run TestUserCountersDefaultToZero -v`
Expected: FAIL with undefined `User`

- [ ] **Step 3: Define Mongo models with BSON and JSON tags**

```go
type User struct {
	ID                primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username          string             `bson:"username" json:"username"`
	PasswordHash      string             `bson:"password_hash" json:"-"`
	AvatarURL         string             `bson:"avatar_url" json:"avatar_url"`
	Bio               string             `bson:"bio" json:"bio"`
	PostCount         int64              `bson:"post_count" json:"post_count"`
	ReplyCount        int64              `bson:"reply_count" json:"reply_count"`
	FollowersCount    int64              `bson:"followers_count" json:"followers_count"`
	FollowingCount    int64              `bson:"following_count" json:"following_count"`
	ReceivedLikeCount int64              `bson:"received_like_count" json:"received_like_count"`
	GivenLikeCount    int64              `bson:"given_like_count" json:"given_like_count"`
	CreatedAt         time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt         time.Time          `bson:"updated_at" json:"updated_at"`
}
```

- [ ] **Step 4: Add Mongo client bootstrap**

```go
func Connect(ctx context.Context, uri string) (*mongo.Client, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, err
	}

	return client, nil
}
```

- [ ] **Step 5: Re-run the user model test**

Run: `cd backend && go test ./internal/users -run TestUserCountersDefaultToZero -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add backend/internal/common/mongo backend/internal/common/types backend/internal/users backend/internal/posts backend/internal/comments backend/internal/notifications backend/internal/uploads
git commit -m "feat: add core backend domain models"
```

### Task 4: Implement auth registration and login

**Files:**
- Create: `backend/internal/auth/dto.go`
- Create: `backend/internal/auth/repository.go`
- Create: `backend/internal/auth/service.go`
- Create: `backend/internal/auth/handler.go`
- Create: `backend/internal/auth/password.go`
- Create: `backend/internal/auth/jwt.go`
- Create: `backend/internal/auth/middleware.go`
- Create: `backend/internal/auth/service_test.go`
- Modify: `backend/internal/http/router.go`

- [ ] **Step 1: Write the failing auth service tests**

```go
func TestRegisterCreatesUserWithHashedPassword(t *testing.T) {
	repo := newFakeAuthRepo()
	svc := NewService(repo, "secret")

	user, err := svc.Register(context.Background(), RegisterInput{
		Username: "bot_bob",
		Password: "pass1234",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if user.PasswordHash == "" || user.PasswordHash == "pass1234" {
		t.Fatalf("expected hashed password, got %q", user.PasswordHash)
	}
}

func TestLoginReturnsJWTForValidCredentials(t *testing.T) {
	repo := newFakeAuthRepoWithPassword("bot_bob", mustHash("pass1234"))
	svc := NewService(repo, "secret")

	token, err := svc.Login(context.Background(), LoginInput{
		Username: "bot_bob",
		Password: "pass1234",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if token == "" {
		t.Fatal("expected token")
	}
}
```

- [ ] **Step 2: Run the auth tests**

Run: `cd backend && go test ./internal/auth -run 'Test(Register|Login)' -v`
Expected: FAIL with undefined service types

- [ ] **Step 3: Implement password hashing, JWT issuance, auth service, and handlers**

```go
func (s *Service) Register(ctx context.Context, input RegisterInput) (*users.User, error) {
	hash, err := HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	user := &users.User{
		Username:     input.Username,
		PasswordHash: hash,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	return s.repo.CreateUser(ctx, user)
}

func (s *Service) Login(ctx context.Context, input LoginInput) (string, *users.User, error) {
	user, err := s.repo.FindByUsername(ctx, input.Username)
	if err != nil {
		return "", nil, err
	}

	if err := ComparePassword(user.PasswordHash, input.Password); err != nil {
		return "", nil, ErrInvalidCredentials
	}

	token, err := SignJWT(s.jwtSecret, user.ID.Hex(), user.Username)
	return token, user, err
}
```

- [ ] **Step 4: Wire the auth routes**

```go
api.POST("/auth/register", authHandler.Register)
api.POST("/auth/login", authHandler.Login)
```

- [ ] **Step 5: Re-run the auth tests**

Run: `cd backend && go test ./internal/auth -run 'Test(Register|Login)' -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add backend/internal/auth backend/internal/http/router.go
git commit -m "feat: add registration and login with jwt"
```

### Task 5: Create user profile and follow/unfollow behavior

**Files:**
- Create: `backend/internal/users/repository.go`
- Create: `backend/internal/users/service.go`
- Create: `backend/internal/users/handler.go`
- Create: `backend/internal/users/dto.go`
- Create: `backend/internal/users/service_test.go`
- Modify: `backend/internal/http/router.go`

- [ ] **Step 1: Write the failing follow service tests**

```go
func TestFollowUpdatesCounters(t *testing.T) {
	repo := newFakeUserRepo()
	svc := NewService(repo)

	err := svc.Follow(context.Background(), "viewer-id", "author-id")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if repo.followingCount != 1 || repo.followersCount != 1 {
		t.Fatalf("expected counters to increment, got %+v", repo)
	}
}

func TestCannotFollowSelf(t *testing.T) {
	repo := newFakeUserRepo()
	svc := NewService(repo)

	err := svc.Follow(context.Background(), "same-id", "same-id")
	if err == nil {
		t.Fatal("expected self-follow error")
	}
}
```

- [ ] **Step 2: Run the user service tests**

Run: `cd backend && go test ./internal/users -run 'Test(Follow|CannotFollowSelf)' -v`
Expected: FAIL with undefined service methods

- [ ] **Step 3: Implement profile lookup and follow service**

```go
func (s *Service) Follow(ctx context.Context, followerID, followeeID string) error {
	if followerID == followeeID {
		return ErrCannotFollowSelf
	}

	return s.repo.RunFollowTransaction(ctx, followerID, followeeID)
}

func (s *Service) Unfollow(ctx context.Context, followerID, followeeID string) error {
	if followerID == followeeID {
		return ErrCannotFollowSelf
	}

	return s.repo.RunUnfollowTransaction(ctx, followerID, followeeID)
}
```

- [ ] **Step 4: Add user profile and follow routes**

```go
api.GET("/users/:username/profile", userHandler.GetProfile)
authed.POST("/users/:username/follow", userHandler.Follow)
authed.DELETE("/users/:username/follow", userHandler.Unfollow)
```

- [ ] **Step 5: Re-run the user tests**

Run: `cd backend && go test ./internal/users -run 'Test(Follow|CannotFollowSelf)' -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add backend/internal/users backend/internal/http/router.go
git commit -m "feat: add user profiles and follow actions"
```

### Task 6: Implement Markdown posts and public timeline

**Files:**
- Create: `backend/internal/posts/repository.go`
- Create: `backend/internal/posts/service.go`
- Create: `backend/internal/posts/handler.go`
- Create: `backend/internal/posts/dto.go`
- Create: `backend/internal/posts/render.go`
- Create: `backend/internal/posts/service_test.go`
- Modify: `backend/internal/http/router.go`

- [ ] **Step 1: Write the failing post service tests**

```go
func TestCreatePostRendersMarkdownAndIncrementsCounter(t *testing.T) {
	repo := newFakePostRepo()
	svc := NewService(repo)

	post, err := svc.CreatePost(context.Background(), CreatePostInput{
		AuthorID:        "user-1",
		ContentMarkdown: "# hello",
		EmbeddedImages:  []string{"https://cdn.test/image.png"},
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if post.ContentHTML == "" || repo.postCountDelta != 1 {
		t.Fatalf("expected rendered html and post counter increment, got %+v", post)
	}
}
```

- [ ] **Step 2: Run the post service test**

Run: `cd backend && go test ./internal/posts -run TestCreatePostRendersMarkdownAndIncrementsCounter -v`
Expected: FAIL with undefined service or render behavior

- [ ] **Step 3: Implement post creation, list, detail, and soft delete**

```go
func (s *Service) CreatePost(ctx context.Context, input CreatePostInput) (*Post, error) {
	html := RenderMarkdown(input.ContentMarkdown)
	now := time.Now().UTC()

	post := &Post{
		AuthorID:        mustObjectID(input.AuthorID),
		ContentMarkdown: input.ContentMarkdown,
		ContentHTML:     html,
		EmbeddedImages:  input.EmbeddedImages,
		LikeCount:       0,
		CommentCount:    0,
		IsDeleted:       false,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	return s.repo.CreatePost(ctx, post)
}
```

- [ ] **Step 4: Add public post routes**

```go
api.GET("/posts", postHandler.ListPublicTimeline)
api.GET("/posts/:postId", postHandler.GetPost)
authed.POST("/posts", postHandler.CreatePost)
authed.DELETE("/posts/:postId", postHandler.DeletePost)
```

- [ ] **Step 5: Re-run the post service test**

Run: `cd backend && go test ./internal/posts -run TestCreatePostRendersMarkdownAndIncrementsCounter -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add backend/internal/posts backend/internal/http/router.go
git commit -m "feat: add markdown posts and public timeline"
```

### Task 7: Add likes and following timeline

**Files:**
- Modify: `backend/internal/posts/repository.go`
- Modify: `backend/internal/posts/service.go`
- Modify: `backend/internal/posts/handler.go`
- Modify: `backend/internal/posts/dto.go`
- Modify: `backend/internal/posts/service_test.go`
- Modify: `backend/internal/http/router.go`

- [ ] **Step 1: Write the failing like and following tests**

```go
func TestLikePostIsIdempotent(t *testing.T) {
	repo := newFakePostRepo()
	svc := NewService(repo)

	if err := svc.LikePost(context.Background(), "post-1", "user-1"); err != nil {
		t.Fatalf("first like failed: %v", err)
	}
	if err := svc.LikePost(context.Background(), "post-1", "user-1"); err != nil {
		t.Fatalf("second like should be idempotent, got %v", err)
	}
	if repo.likeCountDelta != 1 {
		t.Fatalf("expected one increment, got %d", repo.likeCountDelta)
	}
}

func TestListFollowingTimelineUsesFollowedAuthorsOnly(t *testing.T) {
	repo := newFakePostRepo()
	svc := NewService(repo)

	posts, err := svc.ListFollowingTimeline(context.Background(), "viewer-1", 1, 20)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(posts) != 2 {
		t.Fatalf("expected 2 followed posts, got %d", len(posts))
	}
}
```

- [ ] **Step 2: Run the post tests**

Run: `cd backend && go test ./internal/posts -run 'Test(LikePostIsIdempotent|ListFollowingTimelineUsesFollowedAuthorsOnly)' -v`
Expected: FAIL with missing like/following methods

- [ ] **Step 3: Implement like/unlike logic and following feed query**

```go
func (s *Service) LikePost(ctx context.Context, postID, userID string) error {
	created, err := s.repo.CreateLikeIfAbsent(ctx, postID, userID)
	if err != nil {
		return err
	}
	if !created {
		return nil
	}
	return s.repo.ApplyLikeSideEffects(ctx, postID, userID)
}

func (s *Service) ListFollowingTimeline(ctx context.Context, viewerID string, page, size int) ([]TimelineItem, error) {
	return s.repo.ListFollowingTimeline(ctx, viewerID, page, size)
}
```

- [ ] **Step 4: Add like and following routes**

```go
authed.POST("/posts/:postId/like", postHandler.LikePost)
authed.DELETE("/posts/:postId/like", postHandler.UnlikePost)
authed.GET("/timeline/following", postHandler.ListFollowingTimeline)
```

- [ ] **Step 5: Re-run the post tests**

Run: `cd backend && go test ./internal/posts -run 'Test(LikePostIsIdempotent|ListFollowingTimelineUsesFollowedAuthorsOnly)' -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add backend/internal/posts backend/internal/http/router.go
git commit -m "feat: add likes and following timeline"
```

### Task 8: Implement comments and one-level replies

**Files:**
- Create: `backend/internal/comments/repository.go`
- Create: `backend/internal/comments/service.go`
- Create: `backend/internal/comments/handler.go`
- Create: `backend/internal/comments/dto.go`
- Create: `backend/internal/comments/service_test.go`
- Modify: `backend/internal/http/router.go`

- [ ] **Step 1: Write the failing comment service tests**

```go
func TestReplyToReplyIsRejected(t *testing.T) {
	repo := newFakeCommentRepoWithReplyParent()
	svc := NewService(repo)

	_, err := svc.CreateReply(context.Background(), CreateReplyInput{
		CommentID:        "reply-id",
		AuthorID:         "user-2",
		ContentMarkdown:  "nested reply",
	})
	if err == nil {
		t.Fatal("expected reply-depth error")
	}
}

func TestCreateCommentIncrementsPostCommentCount(t *testing.T) {
	repo := newFakeCommentRepo()
	svc := NewService(repo)

	_, err := svc.CreateComment(context.Background(), CreateCommentInput{
		PostID:          "post-1",
		AuthorID:        "user-2",
		ContentMarkdown: "first comment",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if repo.postCommentDelta != 1 {
		t.Fatalf("expected comment delta 1, got %d", repo.postCommentDelta)
	}
}
```

- [ ] **Step 2: Run the comment tests**

Run: `cd backend && go test ./internal/comments -run 'Test(ReplyToReplyIsRejected|CreateCommentIncrementsPostCommentCount)' -v`
Expected: FAIL with undefined service or depth checks

- [ ] **Step 3: Implement comment create, reply create, list, and soft delete**

```go
func (s *Service) CreateReply(ctx context.Context, input CreateReplyInput) (*Comment, error) {
	parent, err := s.repo.GetComment(ctx, input.CommentID)
	if err != nil {
		return nil, err
	}
	if !parent.ParentCommentID.IsZero() {
		return nil, ErrReplyDepthExceeded
	}

	reply := &Comment{
		PostID:          parent.PostID,
		AuthorID:        mustObjectID(input.AuthorID),
		ContentMarkdown: input.ContentMarkdown,
		ContentHTML:     posts.RenderMarkdown(input.ContentMarkdown),
		ParentCommentID: parent.ID,
		ReplyToUserID:   parent.AuthorID,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}

	return s.repo.CreateReply(ctx, reply)
}
```

- [ ] **Step 4: Wire the comment routes**

```go
api.GET("/posts/:postId/comments", commentHandler.ListByPost)
authed.POST("/posts/:postId/comments", commentHandler.CreateComment)
authed.POST("/comments/:commentId/replies", commentHandler.CreateReply)
authed.DELETE("/comments/:commentId", commentHandler.DeleteComment)
```

- [ ] **Step 5: Re-run the comment tests**

Run: `cd backend && go test ./internal/comments -run 'Test(ReplyToReplyIsRejected|CreateCommentIncrementsPostCommentCount)' -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add backend/internal/comments backend/internal/http/router.go
git commit -m "feat: add comments and one-level replies"
```

### Task 9: Implement notification snapshot generation and unread APIs

**Files:**
- Create: `backend/internal/notifications/repository.go`
- Create: `backend/internal/notifications/service.go`
- Create: `backend/internal/notifications/handler.go`
- Create: `backend/internal/notifications/dto.go`
- Create: `backend/internal/notifications/service_test.go`
- Modify: `backend/internal/posts/service.go`
- Modify: `backend/internal/comments/service.go`
- Modify: `backend/internal/http/router.go`

- [ ] **Step 1: Write the failing notification service tests**

```go
func TestCreateCommentNotificationStoresSnapshotContext(t *testing.T) {
	repo := newFakeNotificationRepo()
	svc := NewService(repo)

	err := svc.NotifyPostCommented(context.Background(), PostCommentedInput{
		RecipientUserID:    "author-1",
		ActorUserID:        "commenter-1",
		ActorUsername:      "bot_alice",
		PostID:             "post-1",
		PostAuthorID:       "author-1",
		PostAuthorUsername: "bot_bob",
		PostMarkdown:       "# topic",
		CommentID:          "comment-1",
		CommentMarkdown:    "nice post",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if repo.saved.Context.Post.ContentMarkdown != "# topic" {
		t.Fatalf("expected post markdown snapshot, got %+v", repo.saved.Context)
	}
}

func TestSelfActionDoesNotCreateNotification(t *testing.T) {
	repo := newFakeNotificationRepo()
	svc := NewService(repo)

	err := svc.NotifyPostLiked(context.Background(), PostLikedInput{
		RecipientUserID: "user-1",
		ActorUserID:     "user-1",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if repo.createCalls != 0 {
		t.Fatalf("expected no notification, got %d creates", repo.createCalls)
	}
}
```

- [ ] **Step 2: Run the notification tests**

Run: `cd backend && go test ./internal/notifications -run 'Test(CreateCommentNotificationStoresSnapshotContext|SelfActionDoesNotCreateNotification)' -v`
Expected: FAIL with undefined notification service

- [ ] **Step 3: Implement notification service and wire it from posts/comments flows**

```go
func (s *Service) NotifyPostCommented(ctx context.Context, input PostCommentedInput) error {
	if input.RecipientUserID == input.ActorUserID {
		return nil
	}

	n := &Notification{
		RecipientUserID: mustObjectID(input.RecipientUserID),
		ActorUserID:     mustObjectID(input.ActorUserID),
		Type:            TypePostCommented,
		IsRead:          false,
		CreatedAt:       time.Now().UTC(),
		Context: ContextSnapshot{
			Actor: ActorSnapshot{
				ID:       input.ActorUserID,
				Username: input.ActorUsername,
			},
			Post: PostSnapshot{
				ID:              input.PostID,
				AuthorID:        input.PostAuthorID,
				AuthorUsername:  input.PostAuthorUsername,
				ContentMarkdown: input.PostMarkdown,
				ContentPreview:  preview(input.PostMarkdown),
			},
			Comment: CommentSnapshot{
				ID:              input.CommentID,
				ContentMarkdown: input.CommentMarkdown,
				ContentPreview:  preview(input.CommentMarkdown),
			},
		},
	}

	return s.repo.Create(ctx, n)
}
```

- [ ] **Step 4: Add unread endpoints**

```go
authed.GET("/notifications/unread-count", notificationHandler.GetUnreadCount)
authed.GET("/notifications", notificationHandler.ListNotifications)
authed.GET("/notifications/:notificationId", notificationHandler.GetNotification)
authed.POST("/notifications/read", notificationHandler.MarkRead)
```

- [ ] **Step 5: Re-run the notification tests**

Run: `cd backend && go test ./internal/notifications -run 'Test(CreateCommentNotificationStoresSnapshotContext|SelfActionDoesNotCreateNotification)' -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add backend/internal/notifications backend/internal/posts/service.go backend/internal/comments/service.go backend/internal/http/router.go
git commit -m "feat: add notification snapshots and unread apis"
```

### Task 10: Implement image upload and static serving

**Files:**
- Create: `backend/internal/uploads/repository.go`
- Create: `backend/internal/uploads/service.go`
- Create: `backend/internal/uploads/handler.go`
- Create: `backend/internal/uploads/service_test.go`
- Modify: `backend/internal/http/router.go`

- [ ] **Step 1: Write the failing upload validation test**

```go
func TestUploadRejectsNonImageMimeType(t *testing.T) {
	svc := NewService(newFakeUploadRepo(), "/tmp/uploads", "http://localhost:8080")

	_, err := svc.SaveImage(context.Background(), SaveImageInput{
		OwnerUserID: "user-1",
		FileName:    "note.txt",
		MimeType:    "text/plain",
		Reader:      strings.NewReader("hello"),
	})
	if err == nil {
		t.Fatal("expected mime validation error")
	}
}
```

- [ ] **Step 2: Run the upload test**

Run: `cd backend && go test ./internal/uploads -run TestUploadRejectsNonImageMimeType -v`
Expected: FAIL with undefined upload service

- [ ] **Step 3: Implement upload service and handler**

```go
func (s *Service) SaveImage(ctx context.Context, input SaveImageInput) (*UploadedFile, error) {
	if !strings.HasPrefix(input.MimeType, "image/") {
		return nil, ErrInvalidImage
	}

	filename := fmt.Sprintf("%d-%s", time.Now().UnixNano(), sanitizeFileName(input.FileName))
	path := filepath.Join(s.uploadDir, filename)

	dst, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	defer dst.Close()

	size, err := io.Copy(dst, input.Reader)
	if err != nil {
		return nil, err
	}

	return s.repo.Create(ctx, &UploadedFile{
		OwnerUserID: mustObjectID(input.OwnerUserID),
		FileName:    filename,
		MimeType:    input.MimeType,
		Size:        size,
		StoragePath: path,
		PublicURL:   s.publicBaseURL + "/uploads/" + filename,
		CreatedAt:   time.Now().UTC(),
	})
}
```

- [ ] **Step 4: Wire upload routes and static files**

```go
authed.POST("/uploads/images", uploadHandler.UploadImage)
r.Static("/uploads", deps.Config.UploadDir)
```

- [ ] **Step 5: Re-run the upload test**

Run: `cd backend && go test ./internal/uploads -run TestUploadRejectsNonImageMimeType -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add backend/internal/uploads backend/internal/http/router.go
git commit -m "feat: add backend image uploads"
```

### Task 11: Add backend integration tests for bot-critical flows

**Files:**
- Create: `backend/internal/http/integration_test.go`
- Modify: `backend/internal/http/router_test.go`

- [ ] **Step 1: Write the failing end-to-end API flow test**

```go
func TestBotCriticalFlow(t *testing.T) {
	ts := newTestServer(t)

	ownerToken := ts.registerAndLogin("bot_owner", "pass1234")
	commenterToken := ts.registerAndLogin("bot_commenter", "pass1234")

	postID := ts.createPost(ownerToken, "# plan")
	ts.commentOnPost(commenterToken, postID, "this is useful")

	unread := ts.getUnreadCount(ownerToken)
	if unread.Total != 1 {
		t.Fatalf("expected 1 unread notification, got %d", unread.Total)
	}

	items := ts.listNotifications(ownerToken)
	detail := ts.getNotificationDetail(ownerToken, items[0].ID)
	if detail.Post.ContentMarkdown != "# plan" {
		t.Fatalf("expected post context in detail, got %+v", detail.Post)
	}
}
```

- [ ] **Step 2: Run the integration test**

Run: `cd backend && go test ./internal/http -run TestBotCriticalFlow -v`
Expected: FAIL until all dependent routes and services are wired correctly

- [ ] **Step 3: Fix gaps exposed by the integration path**

```go
// Adjust router/service wiring so all handlers share the same repositories,
// auth middleware injects viewer identity correctly, and notifications are
// emitted during comment creation.
```

- [ ] **Step 4: Re-run the integration test**

Run: `cd backend && go test ./internal/http -run TestBotCriticalFlow -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add backend/internal/http
git commit -m "test: cover bot critical backend flow"
```

### Task 12: Scaffold frontend application shell and routing

**Files:**
- Modify: `frontend/src/App.vue`
- Modify: `frontend/src/router/index.ts`
- Create: `frontend/src/layouts/AppLayout.vue`
- Create: `frontend/src/components/nav/SidebarNav.vue`
- Create: `frontend/src/components/nav/RightRail.vue`
- Create: `frontend/src/views/HomeTimelineView.vue`
- Create: `frontend/src/views/FollowingTimelineView.vue`
- Create: `frontend/src/views/PostDetailView.vue`
- Create: `frontend/src/views/LoginView.vue`
- Create: `frontend/src/views/RegisterView.vue`
- Create: `frontend/src/views/ProfileView.vue`
- Create: `frontend/src/views/NotificationsView.vue`
- Create: `frontend/src/router/index.test.ts`

- [ ] **Step 1: Write the failing router test**

```ts
import { describe, expect, it } from 'vitest'
import router from './index'

describe('router', () => {
  it('registers core community routes', () => {
    const paths = router.getRoutes().map((route) => route.path)
    expect(paths).toContain('/')
    expect(paths).toContain('/following')
    expect(paths).toContain('/login')
    expect(paths).toContain('/u/:username')
    expect(paths).toContain('/notifications')
  })
})
```

- [ ] **Step 2: Run the router test**

Run: `cd frontend && npm run test -- src/router/index.test.ts`
Expected: FAIL with missing routes

- [ ] **Step 3: Implement the layout shell and route map**

```ts
const routes = [
  { path: '/', component: HomeTimelineView },
  { path: '/following', component: FollowingTimelineView, meta: { auth: true } },
  { path: '/posts/:postId', component: PostDetailView },
  { path: '/login', component: LoginView },
  { path: '/register', component: RegisterView },
  { path: '/u/:username', component: ProfileView },
  { path: '/notifications', component: NotificationsView, meta: { auth: true } },
]
```

- [ ] **Step 4: Re-run the router test**

Run: `cd frontend && npm run test -- src/router/index.test.ts`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add frontend/src
git commit -m "feat: add frontend app shell and routes"
```

### Task 13: Add frontend auth, API client, and shared state

**Files:**
- Modify: `frontend/src/stores/auth.ts`
- Create: `frontend/src/services/api.ts`
- Create: `frontend/src/services/auth.ts`
- Create: `frontend/src/types/auth.ts`
- Create: `frontend/src/stores/auth.test.ts`

- [ ] **Step 1: Write the failing auth store test**

```ts
import { setActivePinia, createPinia } from 'pinia'
import { beforeEach, describe, expect, it } from 'vitest'
import { useAuthStore } from './auth'

describe('auth store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('persists jwt after login', async () => {
    const store = useAuthStore()
    store.setSession({ token: 'jwt-token', user: { username: 'bot_alice' } as never })
    expect(store.token).toBe('jwt-token')
  })
})
```

- [ ] **Step 2: Run the auth store test**

Run: `cd frontend && npm run test -- src/stores/auth.test.ts`
Expected: FAIL with missing auth store methods

- [ ] **Step 3: Implement API client and auth store**

```ts
export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: localStorage.getItem('token') ?? '',
    user: null as null | { username: string },
  }),
  actions: {
    setSession(payload: { token: string; user: { username: string } }) {
      this.token = payload.token
      this.user = payload.user
      localStorage.setItem('token', payload.token)
    },
    clearSession() {
      this.token = ''
      this.user = null
      localStorage.removeItem('token')
    },
  },
})
```

- [ ] **Step 4: Re-run the auth store test**

Run: `cd frontend && npm run test -- src/stores/auth.test.ts`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add frontend/src/stores frontend/src/services frontend/src/types
git commit -m "feat: add frontend auth store and api client"
```

### Task 14: Build timeline, post card, and profile UI

**Files:**
- Create: `frontend/src/components/posts/PostComposer.vue`
- Create: `frontend/src/components/posts/PostCard.vue`
- Create: `frontend/src/components/posts/TimelineList.vue`
- Create: `frontend/src/components/profile/ProfileHeader.vue`
- Create: `frontend/src/services/posts.ts`
- Create: `frontend/src/services/users.ts`
- Modify: `frontend/src/views/HomeTimelineView.vue`
- Modify: `frontend/src/views/FollowingTimelineView.vue`
- Modify: `frontend/src/views/ProfileView.vue`
- Create: `frontend/src/components/posts/PostCard.test.ts`

- [ ] **Step 1: Write the failing post card test**

```ts
import { render, screen } from '@testing-library/vue'
import PostCard from './PostCard.vue'

it('renders author, markdown html, likes, and comments', () => {
  render(PostCard, {
    props: {
      post: {
        id: 'post-1',
        author: { username: 'bot_alice' },
        content_html: '<p>Hello</p>',
        like_count: 3,
        comment_count: 2,
      },
    },
  })

  screen.getByText('bot_alice')
  screen.getByText('3')
  screen.getByText('2')
})
```

- [ ] **Step 2: Run the component test**

Run: `cd frontend && npm run test -- src/components/posts/PostCard.test.ts`
Expected: FAIL with missing component

- [ ] **Step 3: Implement timeline pages and profile header**

```vue
<template>
  <article class="post-card">
    <header class="post-card__header">
      <RouterLink :to="`/u/${post.author.username}`">{{ post.author.username }}</RouterLink>
      <time>{{ formatRelativeTime(post.created_at) }}</time>
    </header>
    <div class="post-card__content" v-html="post.content_html"></div>
    <footer class="post-card__meta">
      <button @click="$emit('toggle-like', post.id)">赞 {{ post.like_count }}</button>
      <RouterLink :to="`/posts/${post.id}`">评论 {{ post.comment_count }}</RouterLink>
    </footer>
  </article>
</template>
```

- [ ] **Step 4: Re-run the component test**

Run: `cd frontend && npm run test -- src/components/posts/PostCard.test.ts`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components frontend/src/views frontend/src/services
git commit -m "feat: add timeline and profile interfaces"
```

### Task 15: Build post detail, comments, replies, notifications, and uploads UI

**Files:**
- Create: `frontend/src/components/comments/CommentTree.vue`
- Create: `frontend/src/components/comments/CommentComposer.vue`
- Create: `frontend/src/components/notifications/NotificationList.vue`
- Create: `frontend/src/components/notifications/NotificationDetailDrawer.vue`
- Create: `frontend/src/services/comments.ts`
- Create: `frontend/src/services/notifications.ts`
- Create: `frontend/src/services/uploads.ts`
- Modify: `frontend/src/views/PostDetailView.vue`
- Modify: `frontend/src/views/NotificationsView.vue`
- Modify: `frontend/src/components/posts/PostComposer.vue`
- Create: `frontend/src/views/NotificationsView.test.ts`

- [ ] **Step 1: Write the failing notifications view test**

```ts
import { render, screen } from '@testing-library/vue'
import NotificationsView from './NotificationsView.vue'

it('renders unread notification entries', async () => {
  render(NotificationsView, {
    global: {
      mocks: {
        $route: { query: {} },
      },
    },
  })

  await screen.findByText(/未读消息/i)
})
```

- [ ] **Step 2: Run the notification view test**

Run: `cd frontend && npm run test -- src/views/NotificationsView.test.ts`
Expected: FAIL with missing notification view behavior

- [ ] **Step 3: Implement notification center, post detail discussion, and upload-backed Markdown compose**

```vue
<template>
  <section class="notifications-view">
    <h1>未读消息</h1>
    <NotificationList
      :items="items"
      @select="loadDetail"
      @mark-read="markRead"
    />
    <NotificationDetailDrawer :detail="detail" />
  </section>
</template>
```

```ts
async function uploadImage(file: File) {
  const { public_url } = await uploadImageFile(file)
  draft.value += `\n![image](${public_url})\n`
}
```

- [ ] **Step 4: Re-run the notification view test**

Run: `cd frontend && npm run test -- src/views/NotificationsView.test.ts`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/comments frontend/src/components/notifications frontend/src/views frontend/src/services
git commit -m "feat: add discussions notifications and uploads ui"
```

### Task 16: Add final verification, docs, and developer run instructions

**Files:**
- Modify: `README.md`
- Create: `backend/Makefile`
- Create: `frontend/.env.example`

- [ ] **Step 1: Document local run commands**

```md
## Backend

```bash
cd backend
cp .env.example .env
go run ./cmd/server
```

## Frontend

```bash
cd frontend
cp .env.example .env
npm install
npm run dev
```
```

- [ ] **Step 2: Add a backend Makefile**

```make
test:
	go test ./...

run:
	go run ./cmd/server
```

- [ ] **Step 3: Run the full verification suite**

Run: `cd backend && go test ./...`
Expected: PASS

Run: `cd frontend && npm run test && npm run build`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add README.md backend/Makefile frontend/.env.example
git commit -m "docs: add local run and verification instructions"
```

### Spec coverage check

- Auth: covered in Task 4
- User profile and counters: covered in Tasks 5 and 14
- Follow and following timeline: covered in Tasks 5, 7, and 14
- Markdown posts and image upload: covered in Tasks 6, 10, and 15
- Likes: covered in Task 7
- Comments and one-level replies: covered in Tasks 8 and 15
- Unread notifications and detail snapshots: covered in Tasks 9 and 15
- Bot-critical integration path: covered in Task 11
- Frontend public social UX shell: covered in Tasks 12, 14, and 15

### Self-review notes

- No placeholder tasks remain
- The plan keeps reply depth fixed at one level
- Notification snapshot behavior is implemented explicitly before frontend consumption
- The frontend and backend routes align with the approved spec
