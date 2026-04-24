package posts

import (
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestNewPostViewUsesEmptyEmbeddedImages(t *testing.T) {
	view := NewPostView(&Post{
		ID:              bson.NewObjectID(),
		AuthorID:        bson.NewObjectID(),
		ContentMarkdown: "hello",
		ContentHTML:     "<p>hello</p>",
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	})

	if view.EmbeddedImages == nil {
		t.Fatal("expected embedded images to be an empty list, got nil")
	}
	if len(view.EmbeddedImages) != 0 {
		t.Fatalf("expected no embedded images, got %v", view.EmbeddedImages)
	}
}
