package activities

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type ActivityType string

const (
	ActivityTypeView    ActivityType = "view"
	ActivityTypeLike    ActivityType = "like"
	ActivityTypeComment ActivityType = "comment"
)

type Activity struct {
	ID        bson.ObjectID  `bson:"_id,omitempty" json:"id"`
	UserID    bson.ObjectID  `bson:"user_id" json:"user_id"`
	Type      ActivityType   `bson:"type" json:"type"`
	PostID    bson.ObjectID  `bson:"post_id" json:"post_id"`
	CommentID *bson.ObjectID `bson:"comment_id,omitempty" json:"comment_id,omitempty"`
	Count     int64          `bson:"count" json:"count"`
	CreatedAt time.Time      `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time      `bson:"updated_at" json:"updated_at"`
}
