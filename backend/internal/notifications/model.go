package notifications

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type NotificationType string

type Notification struct {
	ID              bson.ObjectID       `bson:"_id,omitempty" json:"id"`
	RecipientUserID bson.ObjectID       `bson:"recipient_user_id" json:"recipient_user_id"`
	ActorUserID     bson.ObjectID       `bson:"actor_user_id" json:"actor_user_id"`
	Type            NotificationType    `bson:"type" json:"type"`
	IsRead          bool                `bson:"is_read" json:"is_read"`
	ReadAt          *time.Time          `bson:"read_at,omitempty" json:"read_at,omitempty"`
	CreatedAt       time.Time           `bson:"created_at" json:"created_at"`
	Context         NotificationContext `bson:"context" json:"context"`
}

const (
	TypePostLiked      NotificationType = "post_liked"
	TypePostCommented  NotificationType = "post_commented"
	TypeCommentReplied NotificationType = "comment_replied"
)

type NotificationContext struct {
	Post          *NotificationPostSnapshot          `bson:"post,omitempty" json:"post,omitempty"`
	Comment       *NotificationCommentSnapshot       `bson:"comment,omitempty" json:"comment,omitempty"`
	ParentComment *NotificationParentCommentSnapshot `bson:"parent_comment,omitempty" json:"parent_comment,omitempty"`
	Actor         *NotificationActorSnapshot         `bson:"actor,omitempty" json:"actor,omitempty"`
	SnapshotFlags NotificationSnapshotFlags          `bson:"snapshot_flags" json:"snapshot_flags"`
}

type NotificationPostSnapshot struct {
	ID              bson.ObjectID `bson:"id" json:"id"`
	AuthorID        bson.ObjectID `bson:"author_id" json:"author_id"`
	AuthorUsername  string             `bson:"author_username" json:"author_username"`
	ContentMarkdown string             `bson:"content_markdown" json:"content_markdown"`
	ContentPreview  string             `bson:"content_preview" json:"content_preview"`
}

type NotificationCommentSnapshot struct {
	ID              bson.ObjectID `bson:"id" json:"id"`
	AuthorID        bson.ObjectID `bson:"author_id" json:"author_id"`
	AuthorUsername  string             `bson:"author_username" json:"author_username"`
	ContentMarkdown string             `bson:"content_markdown" json:"content_markdown"`
	ContentPreview  string             `bson:"content_preview" json:"content_preview"`
	CreatedAt       time.Time          `bson:"created_at" json:"created_at"`
}

type NotificationParentCommentSnapshot struct {
	ID              bson.ObjectID `bson:"id" json:"id"`
	AuthorID        bson.ObjectID `bson:"author_id" json:"author_id"`
	AuthorUsername  string             `bson:"author_username" json:"author_username"`
	ContentMarkdown string             `bson:"content_markdown" json:"content_markdown"`
	ContentPreview  string             `bson:"content_preview" json:"content_preview"`
	CreatedAt       time.Time          `bson:"created_at" json:"created_at"`
}

type NotificationActorSnapshot struct {
	ID       bson.ObjectID `bson:"id" json:"id"`
	Username string             `bson:"username" json:"username"`
}

type NotificationSnapshotFlags struct {
	PostDeleted          bool `bson:"post_deleted" json:"post_deleted"`
	CommentDeleted       bool `bson:"comment_deleted" json:"comment_deleted"`
	ParentCommentDeleted bool `bson:"parent_comment_deleted" json:"parent_comment_deleted"`
}
