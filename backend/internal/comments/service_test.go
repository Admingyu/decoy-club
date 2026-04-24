package comments

import (
	"context"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestCreateCommentRendersMarkdownAndIncrementsPostCount(t *testing.T) {
	counter := &fakePostCounter{}
	repo := NewMemoryRepository(counter)
	svc := NewService(repo, nil)

	comment, err := svc.CreateComment(context.Background(), CreateCommentInput{
		PostID:          "507f1f77bcf86cd799439011",
		AuthorID:        "507f1f77bcf86cd799439012",
		ContentMarkdown: "# hello",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if comment.ContentHTML == "" {
		t.Fatal("expected rendered html")
	}
	if counter.deltas["507f1f77bcf86cd799439011"] != 1 {
		t.Fatalf("expected post comment count increment, got %d", counter.deltas["507f1f77bcf86cd799439011"])
	}
}

func TestReplyToCommentRejectsNestedReplies(t *testing.T) {
	counter := &fakePostCounter{}
	repo := NewMemoryRepository(counter)
	now := time.Now().UTC()
	rootID := bson.NewObjectID()
	replyID := bson.NewObjectID()
	postID := bson.NewObjectID()
	authorID := bson.NewObjectID()
	repo.comments[rootID.Hex()] = &Comment{
		ID:              rootID,
		PostID:          postID,
		AuthorID:        authorID,
		ContentMarkdown: "root",
		ContentHTML:     "<p>root</p>",
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	repo.comments[replyID.Hex()] = &Comment{
		ID:              replyID,
		PostID:          postID,
		AuthorID:        bson.NewObjectID(),
		ContentMarkdown: "reply",
		ContentHTML:     "<p>reply</p>",
		ParentCommentID: &rootID,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	svc := NewService(repo, nil)
	_, err := svc.ReplyToComment(context.Background(), replyID.Hex(), bson.NewObjectID().Hex(), "nested")
	if err != ErrReplyDepthExceeded {
		t.Fatalf("expected ErrReplyDepthExceeded, got %v", err)
	}
}

func TestBuildCommentTreeGroupsRepliesUnderParent(t *testing.T) {
	postID := bson.NewObjectID()
	rootID := bson.NewObjectID()
	replyID := bson.NewObjectID()
	now := time.Now().UTC()
	comments := []*Comment{
		{
			ID:              rootID,
			PostID:          postID,
			AuthorID:        bson.NewObjectID(),
			ContentMarkdown: "root",
			ContentHTML:     "<p>root</p>",
			CreatedAt:       now,
			UpdatedAt:       now,
		},
		{
			ID:              replyID,
			PostID:          postID,
			AuthorID:        bson.NewObjectID(),
			ContentMarkdown: "reply",
			ContentHTML:     "<p>reply</p>",
			ParentCommentID: &rootID,
			CreatedAt:       now.Add(time.Second),
			UpdatedAt:       now.Add(time.Second),
		},
	}

	tree := BuildCommentTree(comments)
	if len(tree) != 1 {
		t.Fatalf("expected 1 root, got %d", len(tree))
	}
	if len(tree[0].Replies) != 1 {
		t.Fatalf("expected 1 reply, got %d", len(tree[0].Replies))
	}
}

type fakePostCounter struct {
	deltas map[string]int64
}

func (f *fakePostCounter) AdjustCommentCount(_ context.Context, postID string, delta int64) error {
	if f.deltas == nil {
		f.deltas = make(map[string]int64)
	}
	f.deltas[postID] += delta
	return nil
}
