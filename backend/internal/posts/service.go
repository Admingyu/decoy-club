package posts

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

var ErrPostNotFound = errors.New("post not found")
var ErrInvalidAuthorID = errors.New("invalid author id")

type Repository interface {
	CreatePost(ctx context.Context, post *Post) (*Post, error)
	FindPostByID(ctx context.Context, id string) (*Post, error)
	ListPublicTimeline(ctx context.Context, limit int, before *time.Time, beforeID string) ([]*Post, error)
	ListFollowingTimeline(ctx context.Context, viewerID string, page, size int) ([]*Post, error)
	SoftDeletePost(ctx context.Context, postID, authorID string) error
	CreateLikeIfAbsent(ctx context.Context, postID, userID string) (bool, error)
	RemoveLikeIfPresent(ctx context.Context, postID, userID string) (bool, error)
	ApplyLikeSideEffects(ctx context.Context, postID, userID string) error
	RevertLikeSideEffects(ctx context.Context, postID, userID string) error
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
	authorID, err := objectIDFromString(input.AuthorID)
	if err != nil {
		return nil, ErrInvalidAuthorID
	}

	html := RenderMarkdown(input.ContentMarkdown)
	now := time.Now().UTC()

	post := &Post{
		AuthorID:        authorID,
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

func (s *Service) ListPublicTimeline(ctx context.Context, limit int, before *time.Time, beforeID string) ([]*Post, error) {
	return s.repo.ListPublicTimeline(ctx, limit, before, beforeID)
}

func (s *Service) ListFollowingTimeline(ctx context.Context, viewerID string, page, size int) ([]*Post, error) {
	return s.repo.ListFollowingTimeline(ctx, viewerID, page, size)
}

func (s *Service) GetPost(ctx context.Context, postID string) (*Post, error) {
	return s.repo.FindPostByID(ctx, postID)
}

func (s *Service) DeletePost(ctx context.Context, postID, authorID string) error {
	return s.repo.SoftDeletePost(ctx, postID, authorID)
}

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

func (s *Service) UnlikePost(ctx context.Context, postID, userID string) error {
	removed, err := s.repo.RemoveLikeIfPresent(ctx, postID, userID)
	if err != nil {
		return err
	}
	if !removed {
		return nil
	}
	return s.repo.RevertLikeSideEffects(ctx, postID, userID)
}

func objectIDFromString(value string) (bson.ObjectID, error) {
	id, err := bson.ObjectIDFromHex(value)
	if err != nil {
		return bson.NilObjectID, err
	}
	return id, nil
}
