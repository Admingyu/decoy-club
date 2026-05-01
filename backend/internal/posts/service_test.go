package posts

import (
	"context"
	"testing"
	"time"

	"decoy-club/backend/internal/users"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestCreatePostRendersMarkdownAndIncrementsCounter(t *testing.T) {
	repo := newFakePostRepo()
	svc := NewService(repo, nil)

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
	svc := NewService(repo, nil)

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
	recorder := &fakeActivityRecorder{}
	svc := NewService(repo, nil, recorder)

	if err := svc.LikePost(context.Background(), "post-1", "user-1"); err != nil {
		t.Fatalf("first like failed: %v", err)
	}
	if err := svc.LikePost(context.Background(), "post-1", "user-1"); err != nil {
		t.Fatalf("second like should be idempotent, got %v", err)
	}
	if repo.likeCountDelta != 1 {
		t.Fatalf("expected one increment, got %d", repo.likeCountDelta)
	}
	if recorder.likes != 1 {
		t.Fatalf("expected one like activity record, got %d", recorder.likes)
	}
}

func TestListPostLikersReturnsRepositoryUsers(t *testing.T) {
	repo := newFakePostRepo()
	repo.likers["post-1"] = []*users.User{
		{ID: mustObjectID("507f1f77bcf86cd799439031"), Username: "ada", AvatarURL: "/uploads/ada.png"},
	}
	svc := NewService(repo, nil)

	likers, err := svc.ListPostLikers(context.Background(), "post-1", 6)
	if err != nil {
		t.Fatalf("list likers failed: %v", err)
	}
	if len(likers) != 1 {
		t.Fatalf("expected one liker, got %d", len(likers))
	}
	if likers[0].Username != "ada" || likers[0].AvatarURL != "/uploads/ada.png" {
		t.Fatalf("unexpected liker: %+v", likers[0])
	}
}

func TestUnlikePostIsIdempotent(t *testing.T) {
	repo := newFakePostRepo()
	svc := NewService(repo, nil)

	if err := svc.LikePost(context.Background(), "post-1", "user-1"); err != nil {
		t.Fatalf("like failed: %v", err)
	}
	if err := svc.UnlikePost(context.Background(), "post-1", "user-1"); err != nil {
		t.Fatalf("first unlike failed: %v", err)
	}
	if err := svc.UnlikePost(context.Background(), "post-1", "user-1"); err != nil {
		t.Fatalf("second unlike should be idempotent, got %v", err)
	}
	if repo.likeCountDelta != 0 {
		t.Fatalf("expected like counter to return to zero, got %d", repo.likeCountDelta)
	}
}

func TestListFollowingTimelineUsesFollowedAuthorsOnly(t *testing.T) {
	repo := newFakePostRepo()
	svc := NewService(repo, nil)

	posts, err := svc.ListFollowingTimeline(context.Background(), "viewer-1", 1, 20)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(posts) != 2 {
		t.Fatalf("expected 2 followed posts, got %d", len(posts))
	}
}

func TestListPublicTimelineRejectsBeforeIDWithoutBefore(t *testing.T) {
	repo := newFakePostRepo()
	svc := NewService(repo, nil)

	if _, err := svc.ListPublicTimeline(context.Background(), 20, nil, "post-2"); err != ErrInvalidTimelineCursor {
		t.Fatalf("expected ErrInvalidTimelineCursor, got %v", err)
	}
}

type fakePostRepo struct {
	posts              map[string]*Post
	follows            map[string]map[string]struct{}
	likes              map[string]map[string]struct{}
	likers             map[string][]*users.User
	postCountDelta     int64
	likeCountDelta     int64
	lastPublicBefore   *time.Time
	lastPublicBeforeID string
}

func newFakePostRepo() *fakePostRepo {
	repo := &fakePostRepo{
		posts:   make(map[string]*Post),
		follows: make(map[string]map[string]struct{}),
		likes:   make(map[string]map[string]struct{}),
		likers:  make(map[string][]*users.User),
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
	r.lastPublicBefore = before
	r.lastPublicBeforeID = beforeID
	return nil, nil
}

func (r *fakePostRepo) ListTrendingTopics(_ context.Context, limit int) ([]*Topic, error) {
	_ = limit
	return []*Topic{}, nil
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

func (r *fakePostRepo) LikePost(_ context.Context, postID, userID string) (bool, error) {
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
	post.LikeCount++
	r.likeCountDelta++
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

func (r *fakePostRepo) UnlikePost(_ context.Context, postID, userID string) (bool, error) {
	post, ok := r.posts[postID]
	if !ok || post.IsDeleted {
		return false, ErrPostNotFound
	}
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
	if post.LikeCount > 0 {
		post.LikeCount--
	}
	r.likeCountDelta--
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

func (r *fakePostRepo) ListPostsByAuthorID(_ context.Context, authorID string, limit int) ([]*Post, error) {
	posts := make([]*Post, 0, len(r.posts))
	for _, post := range r.posts {
		if post.IsDeleted || post.AuthorID.Hex() != authorID {
			continue
		}
		cp := *post
		posts = append(posts, &cp)
	}
	if limit > 0 && len(posts) > limit {
		posts = posts[:limit]
	}
	return posts, nil
}

func (r *fakePostRepo) HasLike(_ context.Context, postID, userID string) (bool, error) {
	likesByPost, ok := r.likes[postID]
	if !ok {
		return false, nil
	}
	_, exists := likesByPost[userID]
	return exists, nil
}

func (r *fakePostRepo) ListLikedPostIDs(_ context.Context, userID string, postIDs []string) (map[string]bool, error) {
	result := make(map[string]bool)
	for _, postID := range postIDs {
		if likesByPost, ok := r.likes[postID]; ok {
			if _, exists := likesByPost[userID]; exists {
				result[postID] = true
			}
		}
	}
	return result, nil
}

func (r *fakePostRepo) ListPostLikers(_ context.Context, postID string, limit int) ([]*users.User, error) {
	likers := append([]*users.User{}, r.likers[postID]...)
	if limit > 0 && len(likers) > limit {
		likers = likers[:limit]
	}
	return likers, nil
}

var _ Repository = (*fakePostRepo)(nil)

type fakeActivityRecorder struct {
	views    int
	likes    int
	comments int
}

func (r *fakeActivityRecorder) RecordPostView(_ context.Context, userID, postID string) error {
	r.views++
	return nil
}

func (r *fakeActivityRecorder) RecordPostLike(_ context.Context, userID, postID string) error {
	r.likes++
	return nil
}

func (r *fakeActivityRecorder) RecordComment(_ context.Context, userID, postID, commentID string) error {
	r.comments++
	return nil
}

func mustObjectID(hex string) bson.ObjectID {
	id, err := bson.ObjectIDFromHex(hex)
	if err != nil {
		panic(err)
	}
	return id
}
