package activities

import (
	"context"
	"errors"
)

var ErrInvalidActivityType = errors.New("invalid activity type")

type Counts struct {
	Views    int64 `json:"views"`
	Likes    int64 `json:"likes"`
	Comments int64 `json:"comments"`
}

type Repository interface {
	RecordPostView(ctx context.Context, userID, postID string) error
	RecordPostLike(ctx context.Context, userID, postID string) error
	RecordComment(ctx context.Context, userID, postID, commentID string) error
	Counts(ctx context.Context, userID string) (Counts, error)
	List(ctx context.Context, userID string, activityType ActivityType, limit int) ([]*Activity, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) RecordPostView(ctx context.Context, userID, postID string) error {
	if s == nil || s.repo == nil || userID == "" || postID == "" {
		return nil
	}
	return s.repo.RecordPostView(ctx, userID, postID)
}

func (s *Service) RecordPostLike(ctx context.Context, userID, postID string) error {
	if s == nil || s.repo == nil || userID == "" || postID == "" {
		return nil
	}
	return s.repo.RecordPostLike(ctx, userID, postID)
}

func (s *Service) RecordComment(ctx context.Context, userID, postID, commentID string) error {
	if s == nil || s.repo == nil || userID == "" || postID == "" || commentID == "" {
		return nil
	}
	return s.repo.RecordComment(ctx, userID, postID, commentID)
}

func (s *Service) Counts(ctx context.Context, userID string) (Counts, error) {
	if s == nil || s.repo == nil || userID == "" {
		return Counts{}, nil
	}
	return s.repo.Counts(ctx, userID)
}

func (s *Service) List(ctx context.Context, userID string, activityType ActivityType, limit int) ([]*Activity, error) {
	if s == nil || s.repo == nil || userID == "" {
		return []*Activity{}, nil
	}
	if !isValidActivityType(activityType) {
		return nil, ErrInvalidActivityType
	}
	return s.repo.List(ctx, userID, activityType, limit)
}

func ParseActivityType(value string) (ActivityType, error) {
	switch value {
	case "views", "view":
		return ActivityTypeView, nil
	case "likes", "like":
		return ActivityTypeLike, nil
	case "comments", "comment":
		return ActivityTypeComment, nil
	default:
		return "", ErrInvalidActivityType
	}
}

func isValidActivityType(activityType ActivityType) bool {
	return activityType == ActivityTypeView || activityType == ActivityTypeLike || activityType == ActivityTypeComment
}
