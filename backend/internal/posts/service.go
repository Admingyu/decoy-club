package posts

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

var ErrPostNotFound = errors.New("post not found")

type Repository interface {
	CreatePost(ctx context.Context, post *Post) (*Post, error)
	FindPostByID(ctx context.Context, id string) (*Post, error)
	ListPublicTimeline(ctx context.Context, limit int, before *time.Time) ([]*Post, error)
	SoftDeletePost(ctx context.Context, postID, authorID string) error
}

type Service struct {
	repo Repository
}

type CreatePostInput struct {
	AuthorID        string
	ContentMarkdown string
	EmbeddedImages  []string
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreatePost(ctx context.Context, input CreatePostInput) (*Post, error) {
	html := RenderMarkdown(input.ContentMarkdown)
	now := time.Now().UTC()

	post := &Post{
		AuthorID:        objectIDOrZero(input.AuthorID),
		ContentMarkdown: input.ContentMarkdown,
		ContentHTML:     html,
		EmbeddedImages:  append([]string(nil), input.EmbeddedImages...),
		LikeCount:       0,
		CommentCount:    0,
		IsDeleted:       false,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	return s.repo.CreatePost(ctx, post)
}

func (s *Service) ListPublicTimeline(ctx context.Context, limit int, before *time.Time) ([]*Post, error) {
	return s.repo.ListPublicTimeline(ctx, limit, before)
}

func (s *Service) GetPost(ctx context.Context, postID string) (*Post, error) {
	return s.repo.FindPostByID(ctx, postID)
}

func (s *Service) DeletePost(ctx context.Context, postID, authorID string) error {
	return s.repo.SoftDeletePost(ctx, postID, authorID)
}

func objectIDOrZero(value string) bson.ObjectID {
	id, err := bson.ObjectIDFromHex(value)
	if err != nil {
		return bson.NilObjectID
	}
	return id
}
