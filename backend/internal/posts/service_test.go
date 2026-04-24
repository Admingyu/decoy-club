package posts

import (
	"context"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestCreatePostRendersMarkdownAndIncrementsCounter(t *testing.T) {
	repo := newFakePostRepo()
	svc := NewService(repo)

	post, err := svc.CreatePost(context.Background(), CreatePostInput{
		AuthorID:        "507f1f77bcf86cd799439012",
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

func TestCreatePostRejectsInvalidAuthorID(t *testing.T) {
	repo := newFakePostRepo()
	svc := NewService(repo)

	_, err := svc.CreatePost(context.Background(), CreatePostInput{
		AuthorID:        "user-1",
		ContentMarkdown: "# hello",
	})
	if err != ErrInvalidAuthorID {
		t.Fatalf("expected ErrInvalidAuthorID, got %v", err)
	}
}

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

type fakePostRepo struct {
	posts          map[string]*Post
	follows        map[string]map[string]struct{}
	likes          map[string]map[string]struct{}
	postCountDelta int64
	likeCountDelta  int64
}

func newFakePostRepo() *fakePostRepo {
	repo := &fakePostRepo{
		posts:   make(map[string]*Post),
		follows: make(map[string]map[string]struct{}),
		likes:   make(map[string]map[string]struct{}),
	}

	now := time.Now().UTC()
	repo.posts["post-1"] = &Post{
		ID:              mustObjectID("507f1f77bcf86cd799439011"),
		AuthorID:        mustObjectID("507f1f77bcf86cd799439021"),
		ContentMarkdown: "post one",
		ContentHTML:     "<p>post one</p>",
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	repo.posts["post-2"] = &Post{
		ID:              mustObjectID("507f1f77bcf86cd799439012"),
		AuthorID:        mustObjectID("507f1f77bcf86cd799439022"),
		ContentMarkdown: "post two",
		ContentHTML:     "<p>post two</p>",
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	repo.posts["post-3"] = &Post{
		ID:              mustObjectID("507f1f77bcf86cd799439013"),
		AuthorID:        mustObjectID("507f1f77bcf86cd799439023"),
		ContentMarkdown: "post three",
		ContentHTML:     "<p>post three</p>",
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	repo.follows["viewer-1"] = map[string]struct{}{
		repo.posts["post-1"].AuthorID.Hex(): {},
		repo.posts["post-2"].AuthorID.Hex(): {},
	}

	return repo
}

func (r *fakePostRepo) CreatePost(_ context.Context, post *Post) (*Post, error) {
	if post.ID.IsZero() {
		post.ID = mustObjectID("507f1f77bcf86cd799439011")
	}
	now := time.Now().UTC()
	if post.CreatedAt.IsZero() {
		post.CreatedAt = now
	}
	if post.UpdatedAt.IsZero() {
		post.UpdatedAt = now
	}

	cp := *post
	r.posts[post.ID.Hex()] = &cp
	r.postCountDelta++
	return &cp, nil
}

func (r *fakePostRepo) FindPostByID(_ context.Context, id string) (*Post, error) {
	post, ok := r.posts[id]
	if !ok || post.IsDeleted {
		return nil, ErrPostNotFound
	}
	cp := *post
	return &cp, nil
}

func (r *fakePostRepo) ListPublicTimeline(_ context.Context, limit int, before *time.Time, beforeID string) ([]*Post, error) {
	_ = limit
	_ = before
	_ = beforeID
	return nil, nil
}

func (r *fakePostRepo) SoftDeletePost(_ context.Context, postID, authorID string) error {
	post, ok := r.posts[postID]
	if !ok {
		return ErrPostNotFound
	}
	if post.AuthorID.Hex() != authorID {
		return ErrPostNotFound
	}
	now := time.Now().UTC()
	post.IsDeleted = true
	post.DeletedAt = &now
	post.UpdatedAt = now
	return nil
}

func (r *fakePostRepo) CreateLikeIfAbsent(_ context.Context, postID, userID string) (bool, error) {
	post, ok := r.posts[postID]
	if !ok || post.IsDeleted {
		return false, ErrPostNotFound
	}
	if _, ok := r.likes[postID]; !ok {
		r.likes[postID] = make(map[string]struct{})
	}
	if _, exists := r.likes[postID][userID]; exists {
		return false, nil
	}
	r.likes[postID][userID] = struct{}{}
	return true, nil
}

func (r *fakePostRepo) ApplyLikeSideEffects(_ context.Context, postID, userID string) error {
	post, ok := r.posts[postID]
	if !ok || post.IsDeleted {
		return ErrPostNotFound
	}
	if _, ok := r.likes[postID]; !ok {
		return nil
	}
	if _, exists := r.likes[postID][userID]; !exists {
		return nil
	}
	post.LikeCount++
	r.likeCountDelta++
	return nil
}

func (r *fakePostRepo) RemoveLikeIfPresent(_ context.Context, postID, userID string) (bool, error) {
	likesByPost, ok := r.likes[postID]
	if !ok {
		return false, nil
	}
	if _, exists := likesByPost[userID]; !exists {
		return false, nil
	}
	delete(likesByPost, userID)
	if len(likesByPost) == 0 {
		delete(r.likes, postID)
	}
	return true, nil
}

func (r *fakePostRepo) RevertLikeSideEffects(_ context.Context, postID, userID string) error {
	post, ok := r.posts[postID]
	if !ok || post.IsDeleted {
		return ErrPostNotFound
	}
	if _, ok := r.likes[postID]; !ok {
		return nil
	}
	if _, exists := r.likes[postID][userID]; !exists {
		return nil
	}
	if post.LikeCount > 0 {
		post.LikeCount--
	}
	return nil
}

func (r *fakePostRepo) ListFollowingTimeline(_ context.Context, viewerID string, page, size int) ([]*Post, error) {
	following := r.follows[viewerID]
	posts := make([]*Post, 0, len(r.posts))
	for _, post := range r.posts {
		if post.IsDeleted {
			continue
		}
		if _, ok := following[post.AuthorID.Hex()]; !ok {
			continue
		}
		cp := *post
		posts = append(posts, &cp)
	}
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	start := (page - 1) * size
	if start >= len(posts) {
		return []*Post{}, nil
	}
	end := start + size
	if end > len(posts) {
		end = len(posts)
	}
	return posts[start:end], nil
}

var _ Repository = (*fakePostRepo)(nil)

func mustObjectID(hex string) bson.ObjectID {
	id, err := bson.ObjectIDFromHex(hex)
	if err != nil {
		panic(err)
	}
	return id
}
