package comments

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Service struct {
	repo     Repository
	notifier Notifier
}

type Notifier interface {
	NotifyPostCommented(ctx context.Context, commentID string) error
	NotifyCommentReplied(ctx context.Context, commentID string) error
}

type CreateCommentInput struct {
	PostID          string
	AuthorID        string
	ContentMarkdown string
	ParentCommentID string
	ReplyToUserID   string
}

func NewService(repo Repository, notifier Notifier) *Service {
	return &Service{repo: repo, notifier: notifier}
}

func (s *Service) CreateComment(ctx context.Context, input CreateCommentInput) (*Comment, error) {
	postID, err := bson.ObjectIDFromHex(input.PostID)
	if err != nil {
		return nil, ErrInvalidPostID
	}
	authorID, err := bson.ObjectIDFromHex(input.AuthorID)
	if err != nil {
		return nil, ErrInvalidCommentAuthorID
	}

	comment := &Comment{
		PostID:          postID,
		AuthorID:        authorID,
		ContentMarkdown: input.ContentMarkdown,
		ContentHTML:     RenderMarkdown(input.ContentMarkdown),
		IsDeleted:       false,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}

	if input.ParentCommentID != "" {
		parentID, err := bson.ObjectIDFromHex(input.ParentCommentID)
		if err != nil {
			return nil, ErrCommentNotFound
		}
		comment.ParentCommentID = &parentID
	}
	if input.ReplyToUserID != "" {
		replyToUserID, err := bson.ObjectIDFromHex(input.ReplyToUserID)
		if err != nil {
			return nil, ErrInvalidCommentAuthorID
		}
		comment.ReplyToUserID = &replyToUserID
	}

	created, err := s.repo.CreateComment(ctx, comment)
	if err != nil {
		return nil, err
	}

	if s.notifier != nil {
		if created.ParentCommentID != nil {
			if err := s.notifier.NotifyCommentReplied(ctx, created.ID.Hex()); err != nil {
				return nil, err
			}
		} else {
			if err := s.notifier.NotifyPostCommented(ctx, created.ID.Hex()); err != nil {
				return nil, err
			}
		}
	}

	return created, nil
}

func (s *Service) ReplyToComment(ctx context.Context, parentCommentID, authorID, contentMarkdown string) (*Comment, error) {
	parent, err := s.repo.FindCommentByID(ctx, parentCommentID)
	if err != nil {
		return nil, err
	}
	if parent.ParentCommentID != nil {
		return nil, ErrReplyDepthExceeded
	}

	return s.CreateComment(ctx, CreateCommentInput{
		PostID:          parent.PostID.Hex(),
		AuthorID:        authorID,
		ContentMarkdown: contentMarkdown,
		ParentCommentID: parentCommentID,
		ReplyToUserID:   parent.AuthorID.Hex(),
	})
}

func (s *Service) ListCommentsByPostID(ctx context.Context, postID string) ([]*Comment, error) {
	return s.repo.ListCommentsByPostID(ctx, postID)
}

func (s *Service) DeleteComment(ctx context.Context, commentID, authorID string) error {
	return s.repo.SoftDeleteComment(ctx, commentID, authorID)
}
