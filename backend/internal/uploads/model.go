package uploads

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type UploadedFile struct {
	ID          bson.ObjectID `bson:"_id,omitempty" json:"id"`
	OwnerUserID bson.ObjectID `bson:"owner_user_id" json:"owner_user_id"`
	FileName    string        `bson:"file_name" json:"file_name"`
	MimeType    string        `bson:"mime_type" json:"mime_type"`
	Size        int64         `bson:"size" json:"size"`
	StoragePath string        `bson:"storage_path" json:"storage_path"`
	PublicURL   string        `bson:"public_url" json:"public_url"`
	CreatedAt   time.Time     `bson:"created_at" json:"created_at"`
}
