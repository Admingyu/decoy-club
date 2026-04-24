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

type fakePostRepo struct {
	posts          map[string]*Post
	postCountDelta int64
}

func newFakePostRepo() *fakePostRepo {
	return &fakePostRepo{
		posts: make(map[string]*Post),
	}
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

func (r *fakePostRepo) ListPublicTimeline(_ context.Context, limit int, before *time.Time) ([]*Post, error) {
	_ = limit
	_ = before
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

var _ Repository = (*fakePostRepo)(nil)

func mustObjectID(hex string) bson.ObjectID {
	id, err := bson.ObjectIDFromHex(hex)
	if err != nil {
		panic(err)
	}
	return id
}
