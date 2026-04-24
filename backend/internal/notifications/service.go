package notifications

import (
	"context"
	"strings"

	"decoy-club/backend/internal/comments"
	"decoy-club/backend/internal/posts"
	"decoy-club/backend/internal/users"
)

type Service struct {
	repo         Repository
	usersRepo    users.Repository
	postsRepo    posts.Repository
	commentsRepo comments.Repository
}

func NewService(repo Repository, usersRepo users.Repository, postsRepo posts.Repository, commentsRepo comments.Repository) *Service {
	return &Service{
		repo:         repo,
		usersRepo:    usersRepo,
		postsRepo:    postsRepo,
		commentsRepo: commentsRepo,
	}
}

func (s *Service) NotifyPostLiked(ctx context.Context, postID, actorUserID string) error {
	post, err := s.postsRepo.FindPostByID(ctx, postID)
	if err != nil {
		return err
	}
	if post.AuthorID.Hex() == actorUserID {
		return nil
	}
	actor, err := s.usersRepo.FindByID(ctx, actorUserID)
	if err != nil {
		return err
	}
	author, err := s.usersRepo.FindByID(ctx, post.AuthorID.Hex())
	if err != nil {
		return err
	}

	return s.createNotification(ctx, &Notification{
		RecipientUserID: author.ID,
		ActorUserID:     actor.ID,
		Type:            TypePostLiked,
		IsRead:          false,
		Context: NotificationContext{
			Post:          buildPostSnapshot(post, author),
			Actor:         buildActorSnapshot(actor),
			SnapshotFlags: NotificationSnapshotFlags{PostDeleted: post.IsDeleted},
		},
	})
}

func (s *Service) NotifyPostCommented(ctx context.Context, commentID string) error {
	comment, err := s.commentsRepo.FindCommentByID(ctx, commentID)
	if err != nil {
		return err
	}
	post, err := s.postsRepo.FindPostByID(ctx, comment.PostID.Hex())
	if err != nil {
		return err
	}
	if post.AuthorID == comment.AuthorID {
		return nil
	}
	actor, err := s.usersRepo.FindByID(ctx, comment.AuthorID.Hex())
	if err != nil {
		return err
	}
	author, err := s.usersRepo.FindByID(ctx, post.AuthorID.Hex())
	if err != nil {
		return err
	}

	return s.createNotification(ctx, &Notification{
		RecipientUserID: author.ID,
		ActorUserID:     actor.ID,
		Type:            TypePostCommented,
		IsRead:          false,
		Context: NotificationContext{
			Post:          buildPostSnapshot(post, author),
			Comment:       buildCommentSnapshot(comment, actor),
			Actor:         buildActorSnapshot(actor),
			SnapshotFlags: NotificationSnapshotFlags{PostDeleted: post.IsDeleted, CommentDeleted: comment.IsDeleted},
		},
	})
}

func (s *Service) NotifyCommentReplied(ctx context.Context, commentID string) error {
	comment, err := s.commentsRepo.FindCommentByID(ctx, commentID)
	if err != nil {
		return err
	}
	if comment.ParentCommentID == nil {
		return nil
	}
	parent, err := s.commentsRepo.FindCommentByID(ctx, comment.ParentCommentID.Hex())
	if err != nil {
		return err
	}
	if parent.AuthorID == comment.AuthorID {
		return nil
	}
	post, err := s.postsRepo.FindPostByID(ctx, comment.PostID.Hex())
	if err != nil {
		return err
	}
	actor, err := s.usersRepo.FindByID(ctx, comment.AuthorID.Hex())
	if err != nil {
		return err
	}
	postAuthor, err := s.usersRepo.FindByID(ctx, post.AuthorID.Hex())
	if err != nil {
		return err
	}
	parentAuthor, err := s.usersRepo.FindByID(ctx, parent.AuthorID.Hex())
	if err != nil {
		return err
	}

	return s.createNotification(ctx, &Notification{
		RecipientUserID: parentAuthor.ID,
		ActorUserID:     actor.ID,
		Type:            TypeCommentReplied,
		IsRead:          false,
		Context: NotificationContext{
			Post:          buildPostSnapshot(post, postAuthor),
			Comment:       buildCommentSnapshot(comment, actor),
			ParentComment: buildParentCommentSnapshot(parent, parentAuthor),
			Actor:         buildActorSnapshot(actor),
			SnapshotFlags: NotificationSnapshotFlags{PostDeleted: post.IsDeleted, CommentDeleted: comment.IsDeleted, ParentCommentDeleted: parent.IsDeleted},
		},
	})
}

func (s *Service) GetUnreadCount(ctx context.Context, viewerID string) (map[NotificationType]int64, int64, error) {
	return s.repo.CountUnread(ctx, viewerID)
}

func (s *Service) ListNotifications(ctx context.Context, viewerID string, unreadOnly bool, page, pageSize int) ([]*Notification, error) {
	return s.repo.ListNotifications(ctx, viewerID, unreadOnly, page, pageSize)
}

func (s *Service) GetNotification(ctx context.Context, viewerID, notificationID string) (*Notification, error) {
	return s.repo.FindNotificationByID(ctx, notificationID, viewerID)
}

func (s *Service) MarkRead(ctx context.Context, viewerID string, notificationIDs []string) error {
	return s.repo.MarkRead(ctx, viewerID, notificationIDs)
}

func (s *Service) createNotification(ctx context.Context, notification *Notification) error {
	_, err := s.repo.CreateNotification(ctx, notification)
	return err
}

func buildPostSnapshot(post *posts.Post, author *users.User) *NotificationPostSnapshot {
	if post == nil || author == nil {
		return nil
	}
	return &NotificationPostSnapshot{
		ID:              post.ID,
		AuthorID:        author.ID,
		AuthorUsername:  author.Username,
		ContentMarkdown: post.ContentMarkdown,
		ContentPreview:  preview(post.ContentMarkdown),
	}
}

func buildCommentSnapshot(comment *comments.Comment, author *users.User) *NotificationCommentSnapshot {
	if comment == nil || author == nil {
		return nil
	}
	return &NotificationCommentSnapshot{
		ID:              comment.ID,
		AuthorID:        author.ID,
		AuthorUsername:  author.Username,
		ContentMarkdown: comment.ContentMarkdown,
		ContentPreview:  preview(comment.ContentMarkdown),
		CreatedAt:       comment.CreatedAt,
	}
}

func buildParentCommentSnapshot(comment *comments.Comment, author *users.User) *NotificationParentCommentSnapshot {
	if comment == nil || author == nil {
		return nil
	}
	return &NotificationParentCommentSnapshot{
		ID:              comment.ID,
		AuthorID:        author.ID,
		AuthorUsername:  author.Username,
		ContentMarkdown: comment.ContentMarkdown,
		ContentPreview:  preview(comment.ContentMarkdown),
		CreatedAt:       comment.CreatedAt,
	}
}

func buildActorSnapshot(actor *users.User) *NotificationActorSnapshot {
	if actor == nil {
		return nil
	}
	return &NotificationActorSnapshot{ID: actor.ID, Username: actor.Username}
}

func preview(input string) string {
	normalized := strings.TrimSpace(strings.ReplaceAll(input, "\n", " "))
	if len(normalized) <= 140 {
		return normalized
	}
	return normalized[:140]
}
