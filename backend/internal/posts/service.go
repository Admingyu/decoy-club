package posts

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

var ErrPostNotFound = errors.New("post not found")
var ErrInvalidAuthorID = errors.New("invalid author id")
var ErrInvalidTimelineCursor = errors.New("before_id requires before")

type Repository interface {
	CreatePost(ctx context.Context, post *Post) (*Post, error)
	FindPostByID(ctx context.Context, id string) (*Post, error)
	ListPublicTimeline(ctx context.Context, limit int, before *time.Time, beforeID string) ([]*Post, error)
	ListFollowingTimeline(ctx context.Context, viewerID string, page, size int) ([]*Post, error)
	ListPostsByAuthorID(ctx context.Context, authorID string, limit int) ([]*Post, error)
	HasLike(ctx context.Context, postID, userID string) (bool, error)
	ListLikedPostIDs(ctx context.Context, userID string, postIDs []string) (map[string]bool, error)
	SoftDeletePost(ctx context.Context, postID, authorID string) error
	LikePost(ctx context.Context, postID, userID string) (bool, error)
	UnlikePost(ctx context.Context, postID, userID string) (bool, error)
}

type LikeNotifier interface {
	NotifyPostLiked(ctx context.Context, postID, actorUserID string) error
}

type Service struct {
	repo     Repository
	notifier LikeNotifier
}

type CreatePostInput struct {
	AuthorID        string
	ContentMarkdown string
	EmbeddedImages  []string
}

func NewService(repo Repository, notifier LikeNotifier) *Service {
	return &Service{repo: repo, notifier: notifier}
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
	if beforeID != "" && (before == nil || before.IsZero()) {
		return nil, ErrInvalidTimelineCursor
	}
	return s.repo.ListPublicTimeline(ctx, limit, before, beforeID)
}

func (s *Service) ListFollowingTimeline(ctx context.Context, viewerID string, page, size int) ([]*Post, error) {
	return s.repo.ListFollowingTimeline(ctx, viewerID, page, size)
}

func (s *Service) ListPostsByAuthorID(ctx context.Context, authorID string, limit int) ([]*Post, error) {
	return s.repo.ListPostsByAuthorID(ctx, authorID, limit)
}

func (s *Service) GetPost(ctx context.Context, postID string) (*Post, error) {
	return s.repo.FindPostByID(ctx, postID)
}

func (s *Service) DeletePost(ctx context.Context, postID, authorID string) error {
	return s.repo.SoftDeletePost(ctx, postID, authorID)
}

func (s *Service) LikePost(ctx context.Context, postID, userID string) error {
	changed, err := s.repo.LikePost(ctx, postID, userID)
	if err != nil {
		return err
	}
	if changed && s.notifier != nil {
		return s.notifier.NotifyPostLiked(ctx, postID, userID)
	}
	return nil
}

func (s *Service) UnlikePost(ctx context.Context, postID, userID string) error {
	_, err := s.repo.UnlikePost(ctx, postID, userID)
	return err
}

func objectIDFromString(value string) (bson.ObjectID, error) {
	id, err := bson.ObjectIDFromHex(value)
	if err != nil {
		return bson.NilObjectID, err
	}
	return id, nil
}
