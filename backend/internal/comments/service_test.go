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
	recorder := &fakeActivityRecorder{}
	svc := NewService(repo, nil, recorder)

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
	if recorder.comments != 1 {
		t.Fatalf("expected comment activity record, got %d", recorder.comments)
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

func TestLikeCommentIsIdempotent(t *testing.T) {
	repo := NewMemoryRepository(nil)
	recorder := &fakeActivityRecorder{}
	commentID := bson.NewObjectID()
	postID := bson.NewObjectID()
	userID := bson.NewObjectID().Hex()
	now := time.Now().UTC()
	repo.comments[commentID.Hex()] = &Comment{
		ID:              commentID,
		PostID:          postID,
		AuthorID:        bson.NewObjectID(),
		ContentMarkdown: "comment",
		ContentHTML:     "<p>comment</p>",
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	svc := NewService(repo, nil, recorder)
	if err := svc.LikeComment(context.Background(), commentID.Hex(), userID); err != nil {
		t.Fatalf("first like failed: %v", err)
	}
	if err := svc.LikeComment(context.Background(), commentID.Hex(), userID); err != nil {
		t.Fatalf("second like should be idempotent, got %v", err)
	}

	comment, err := repo.FindCommentByID(context.Background(), commentID.Hex())
	if err != nil {
		t.Fatalf("expected comment, got %v", err)
	}
	if comment.LikeCount != 1 {
		t.Fatalf("expected one like, got %d", comment.LikeCount)
	}
	liked, err := repo.HasLike(context.Background(), commentID.Hex(), userID)
	if err != nil {
		t.Fatalf("has like failed: %v", err)
	}
	if !liked {
		t.Fatal("expected viewer like to be stored")
	}
	if recorder.commentLikes != 1 {
		t.Fatalf("expected one comment like activity record, got %d", recorder.commentLikes)
	}
	if recorder.lastPostID != postID.Hex() || recorder.lastCommentID != commentID.Hex() {
		t.Fatalf("expected activity to reference post/comment, got post=%s comment=%s", recorder.lastPostID, recorder.lastCommentID)
	}
}

func TestUnlikeCommentIsIdempotent(t *testing.T) {
	repo := NewMemoryRepository(nil)
	commentID := bson.NewObjectID()
	userID := bson.NewObjectID().Hex()
	now := time.Now().UTC()
	repo.comments[commentID.Hex()] = &Comment{
		ID:              commentID,
		PostID:          bson.NewObjectID(),
		AuthorID:        bson.NewObjectID(),
		ContentMarkdown: "comment",
		ContentHTML:     "<p>comment</p>",
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	svc := NewService(repo, nil)
	if err := svc.LikeComment(context.Background(), commentID.Hex(), userID); err != nil {
		t.Fatalf("like failed: %v", err)
	}
	if err := svc.UnlikeComment(context.Background(), commentID.Hex(), userID); err != nil {
		t.Fatalf("first unlike failed: %v", err)
	}
	if err := svc.UnlikeComment(context.Background(), commentID.Hex(), userID); err != nil {
		t.Fatalf("second unlike should be idempotent, got %v", err)
	}

	comment, err := repo.FindCommentByID(context.Background(), commentID.Hex())
	if err != nil {
		t.Fatalf("expected comment, got %v", err)
	}
	if comment.LikeCount != 0 {
		t.Fatalf("expected zero likes, got %d", comment.LikeCount)
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

type fakeActivityRecorder struct {
	comments      int
	commentLikes  int
	lastPostID    string
	lastCommentID string
}

func (r *fakeActivityRecorder) RecordComment(_ context.Context, userID, postID, commentID string) error {
	r.comments++
	return nil
}

func (r *fakeActivityRecorder) RecordCommentLike(_ context.Context, userID, postID, commentID string) error {
	r.commentLikes++
	r.lastPostID = postID
	r.lastCommentID = commentID
	return nil
}
