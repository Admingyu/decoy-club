package posts

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Post struct {
	ID              bson.ObjectID `bson:"_id,omitempty" json:"id"`
	AuthorID        bson.ObjectID `bson:"author_id" json:"author_id"`
	ContentMarkdown string        `bson:"content_markdown" json:"content_markdown"`
	ContentHTML     string        `bson:"content_html" json:"content_html"`
	EmbeddedImages  []string      `bson:"embedded_images" json:"embedded_images"`
	Topics          []string      `bson:"topics" json:"topics"`
	LikeCount       int64         `bson:"like_count" json:"like_count"`
	CommentCount    int64         `bson:"comment_count" json:"comment_count"`
	IsDeleted       bool          `bson:"is_deleted" json:"is_deleted"`
	CreatedAt       time.Time     `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time     `bson:"updated_at" json:"updated_at"`
	DeletedAt       *time.Time    `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}

type Topic struct {
	ID         bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Name       string        `bson:"name" json:"name"`
	PostCount  int64         `bson:"post_count" json:"post_count"`
	CreatedAt  time.Time     `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time     `bson:"updated_at" json:"updated_at"`
	LastPostAt time.Time     `bson:"last_post_at" json:"last_post_at"`
}
