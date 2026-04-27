package users

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type User struct {
	ID                bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Username          string        `bson:"username" json:"username"`
	PasswordHash      string        `bson:"password_hash" json:"-"`
	AvatarURL         string        `bson:"avatar_url" json:"avatar_url"`
	Bio               string        `bson:"bio" json:"bio"`
	StatusText        string        `bson:"status_text" json:"status_text"`
	StatusPreset      string        `bson:"status_preset" json:"status_preset"`
	PostCount         int64         `bson:"post_count" json:"post_count"`
	ReplyCount        int64         `bson:"reply_count" json:"reply_count"`
	FollowersCount    int64         `bson:"followers_count" json:"followers_count"`
	FollowingCount    int64         `bson:"following_count" json:"following_count"`
	ReceivedLikeCount int64         `bson:"received_like_count" json:"received_like_count"`
	GivenLikeCount    int64         `bson:"given_like_count" json:"given_like_count"`
	CreatedAt         time.Time     `bson:"created_at" json:"created_at"`
	UpdatedAt         time.Time     `bson:"updated_at" json:"updated_at"`
}
