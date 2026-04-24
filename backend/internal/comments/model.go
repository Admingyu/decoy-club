package comments

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Comment struct {
	ID              bson.ObjectID       `bson:"_id,omitempty" json:"id"`
	PostID          bson.ObjectID       `bson:"post_id" json:"post_id"`
	AuthorID        bson.ObjectID       `bson:"author_id" json:"author_id"`
	ContentMarkdown string              `bson:"content_markdown" json:"content_markdown"`
	ContentHTML     string              `bson:"content_html" json:"content_html"`
	ParentCommentID *bson.ObjectID      `bson:"parent_comment_id" json:"parent_comment_id"`
	ReplyToUserID   *bson.ObjectID      `bson:"reply_to_user_id" json:"reply_to_user_id"`
	IsDeleted       bool                `bson:"is_deleted" json:"is_deleted"`
	CreatedAt       time.Time           `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time           `bson:"updated_at" json:"updated_at"`
	DeletedAt       *time.Time          `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}
