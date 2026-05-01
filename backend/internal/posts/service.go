package posts

import (
	"context"
	"errors"
	"time"

	"decoy-club/backend/internal/common/textparse"
	"decoy-club/backend/internal/users"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var ErrPostNotFound = errors.New("post not found")
var ErrInvalidAuthorID = errors.New("invalid author id")
var ErrInvalidTimelineCursor = errors.New("before_id requires before")

type Repository interface {
	CreatePost(ctx context.Context, post *Post) (*Post, error)
	FindPostByID(ctx context.Context, id string) (*Post, error)
	ListPublicTimeline(ctx context.Context, limit int, before *time.Time, beforeID string) ([]*Post, error)
	ListTrendingTopics(ctx context.Context, limit int) ([]*Topic, error)
	ListFollowingTimeline(ctx context.Context, viewerID string, page, size int) ([]*Post, error)
	ListPostsByAuthorID(ctx context.Context, authorID string, limit int) ([]*Post, error)
	HasLike(ctx context.Context, postID, userID string) (bool, error)
	ListLikedPostIDs(ctx context.Context, userID string, postIDs []string) (map[string]bool, error)
	ListPostLikers(ctx context.Context, postID string, limit int) ([]*users.User, error)
	SoftDeletePost(ctx context.Context, postID, authorID string) error
	LikePost(ctx context.Context, postID, userID string) (bool, error)
	UnlikePost(ctx context.Context, postID, userID string) (bool, error)
}

type LikeNotifier interface {
	NotifyPostLiked(ctx context.Context, postID, actorUserID string) error
	NotifyPostMentioned(ctx context.Context, postID, actorUserID string, mentionedUsernames []string) error
}

type ActivityRecorder interface {
	RecordPostView(ctx context.Context, userID, postID string) error
	RecordPostLike(ctx context.Context, userID, postID string) error
}

type Service struct {
	repo             Repository
	notifier         LikeNotifier
	activityRecorder ActivityRecorder
}

type CreatePostInput struct {
	AuthorID        string
	ContentMarkdown string
	EmbeddedImages  []string
}

func NewService(repo Repository, notifier LikeNotifier, recorders ...ActivityRecorder) *Service {
	var recorder ActivityRecorder
	if len(recorders) > 0 {
		recorder = recorders[0]
	}
	return &Service{repo: repo, notifier: notifier, activityRecorder: recorder}
}

func (s *Service) CreatePost(ctx context.Context, input CreatePostInput) (*Post, error) {
	authorID, err := objectIDFromString(input.AuthorID)
	if err != nil {
		return nil, ErrInvalidAuthorID
	}

	html := RenderMarkdown(input.ContentMarkdown)
	topics := textparse.ExtractTopics(input.ContentMarkdown)
	now := time.Now().UTC()

	post := &Post{
		AuthorID:        authorID,
		ContentMarkdown: input.ContentMarkdown,
		ContentHTML:     html,
		EmbeddedImages:  append([]string(nil), input.EmbeddedImages...),
		Topics:          topics,
		LikeCount:       0,
		CommentCount:    0,
		IsDeleted:       false,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	created, err := s.repo.CreatePost(ctx, post)
	if err != nil {
		return nil, err
	}
	if s.notifier != nil {
		if err := s.notifier.NotifyPostMentioned(ctx, created.ID.Hex(), input.AuthorID, textparse.ExtractMentions(input.ContentMarkdown)); err != nil {
			return nil, err
		}
	}
	return created, nil
}

func (s *Service) ListPublicTimeline(ctx context.Context, limit int, before *time.Time, beforeID string) ([]*Post, error) {
	if beforeID != "" && (before == nil || before.IsZero()) {
		return nil, ErrInvalidTimelineCursor
	}
	return s.repo.ListPublicTimeline(ctx, limit, before, beforeID)
}

func (s *Service) ListTrendingTopics(ctx context.Context, limit int) ([]*Topic, error) {
	return s.repo.ListTrendingTopics(ctx, limit)
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

func (s *Service) ListPostLikers(ctx context.Context, postID string, limit int) ([]*users.User, error) {
	return s.repo.ListPostLikers(ctx, postID, limit)
}

func (s *Service) RecordPostView(ctx context.Context, userID, postID string) error {
	if s.activityRecorder == nil || userID == "" || postID == "" {
		return nil
	}
	return s.activityRecorder.RecordPostView(ctx, userID, postID)
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
		if err := s.notifier.NotifyPostLiked(ctx, postID, userID); err != nil {
			return err
		}
	}
	if changed && s.activityRecorder != nil {
		return s.activityRecorder.RecordPostLike(ctx, userID, postID)
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
