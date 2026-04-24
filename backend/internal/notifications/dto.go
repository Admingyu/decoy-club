package notifications

import "time"

type UnreadCountResponse struct {
	Total  int64            `json:"total"`
	ByType map[string]int64 `json:"by_type"`
}

type NotificationActorView struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

type NotificationPostView struct {
	ID              string `json:"id"`
	AuthorID        string `json:"author_id"`
	AuthorUsername  string `json:"author_username"`
	ContentMarkdown string `json:"content_markdown,omitempty"`
	ContentPreview  string `json:"content_preview"`
}

type NotificationCommentView struct {
	ID              string    `json:"id"`
	AuthorID        string    `json:"author_id"`
	AuthorUsername  string    `json:"author_username"`
	ContentMarkdown string    `json:"content_markdown,omitempty"`
	ContentPreview  string    `json:"content_preview"`
	CreatedAt       time.Time `json:"created_at"`
}

type NotificationSnapshotFlagsView struct {
	PostDeleted          bool `json:"post_deleted"`
	CommentDeleted       bool `json:"comment_deleted"`
	ParentCommentDeleted bool `json:"parent_comment_deleted"`
}

type NotificationListItem struct {
	ID            string                        `json:"id"`
	Type          NotificationType              `json:"type"`
	IsRead        bool                          `json:"is_read"`
	CreatedAt     time.Time                     `json:"created_at"`
	Actor         *NotificationActorView        `json:"actor,omitempty"`
	Post          *NotificationPostView         `json:"post,omitempty"`
	Comment       *NotificationCommentView      `json:"comment,omitempty"`
	ParentComment *NotificationCommentView      `json:"parent_comment,omitempty"`
	SnapshotFlags NotificationSnapshotFlagsView `json:"snapshot_flags"`
}

type NotificationListResponse struct {
	Notifications []*NotificationListItem `json:"notifications"`
}

type NotificationDetailResponse struct {
	Notification *NotificationListItem `json:"notification"`
}

type MarkReadRequest struct {
	NotificationIDs []string `json:"notification_ids"`
}

func newNotificationListItem(notification *Notification, includeFullMarkdown bool) *NotificationListItem {
	if notification == nil {
		return nil
	}
	item := &NotificationListItem{
		ID:        notification.ID.Hex(),
		Type:      notification.Type,
		IsRead:    notification.IsRead,
		CreatedAt: notification.CreatedAt,
		SnapshotFlags: NotificationSnapshotFlagsView{
			PostDeleted:          notification.Context.SnapshotFlags.PostDeleted,
			CommentDeleted:       notification.Context.SnapshotFlags.CommentDeleted,
			ParentCommentDeleted: notification.Context.SnapshotFlags.ParentCommentDeleted,
		},
	}
	if actor := notification.Context.Actor; actor != nil {
		item.Actor = &NotificationActorView{ID: actor.ID.Hex(), Username: actor.Username}
	}
	if post := notification.Context.Post; post != nil {
		item.Post = &NotificationPostView{
			ID:             post.ID.Hex(),
			AuthorID:       post.AuthorID.Hex(),
			AuthorUsername: post.AuthorUsername,
			ContentPreview: post.ContentPreview,
		}
		if includeFullMarkdown {
			item.Post.ContentMarkdown = post.ContentMarkdown
		}
	}
	if comment := notification.Context.Comment; comment != nil {
		item.Comment = &NotificationCommentView{
			ID:             comment.ID.Hex(),
			AuthorID:       comment.AuthorID.Hex(),
			AuthorUsername: comment.AuthorUsername,
			ContentPreview: comment.ContentPreview,
			CreatedAt:      comment.CreatedAt,
		}
		if includeFullMarkdown {
			item.Comment.ContentMarkdown = comment.ContentMarkdown
		}
	}
	if parent := notification.Context.ParentComment; parent != nil {
		item.ParentComment = &NotificationCommentView{
			ID:             parent.ID.Hex(),
			AuthorID:       parent.AuthorID.Hex(),
			AuthorUsername: parent.AuthorUsername,
			ContentPreview: parent.ContentPreview,
			CreatedAt:      parent.CreatedAt,
		}
		if includeFullMarkdown {
			item.ParentComment.ContentMarkdown = parent.ContentMarkdown
		}
	}
	return item
}

func newNotificationListItems(notifications []*Notification, includeFullMarkdown bool) []*NotificationListItem {
	if len(notifications) == 0 {
		return []*NotificationListItem{}
	}
	items := make([]*NotificationListItem, 0, len(notifications))
	for _, notification := range notifications {
		items = append(items, newNotificationListItem(notification, includeFullMarkdown))
	}
	return items
}
