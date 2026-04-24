package notifications

import (
	"context"
	"testing"
	"time"

	"decoy-club/backend/internal/comments"
	"decoy-club/backend/internal/posts"
	"decoy-club/backend/internal/users"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestNotifyPostLikedCreatesUnreadNotification(t *testing.T) {
	usersRepo := users.NewMemoryRepository()
	postsRepo := posts.NewMemoryRepository(nil)
	commentsRepo := comments.NewMemoryRepository(nil)
	repo := NewMemoryRepository()

	postAuthor := seedUser(t, usersRepo, "author")
	actor := seedUser(t, usersRepo, "liker")
	post := seedPost(t, postsRepo, postAuthor.ID.Hex(), "hello")

	svc := NewService(repo, usersRepo, postsRepo, commentsRepo)
	if err := svc.NotifyPostLiked(context.Background(), post.ID.Hex(), actor.ID.Hex()); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	notifications, err := repo.ListNotifications(context.Background(), postAuthor.ID.Hex(), true, 1, 20)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(notifications) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(notifications))
	}
	if notifications[0].Type != TypePostLiked {
		t.Fatalf("expected post liked notification, got %s", notifications[0].Type)
	}
}

func TestNotifyCommentRepliedIncludesParentContext(t *testing.T) {
	usersRepo := users.NewMemoryRepository()
	postsRepo := posts.NewMemoryRepository(nil)
	commentsRepo := comments.NewMemoryRepository(nil)
	repo := NewMemoryRepository()

	postAuthor := seedUser(t, usersRepo, "poster")
	parentAuthor := seedUser(t, usersRepo, "parent")
	replyAuthor := seedUser(t, usersRepo, "replier")
	post := seedPost(t, postsRepo, postAuthor.ID.Hex(), "post")
	parent := seedComment(t, commentsRepo, post.ID, parentAuthor.ID, "parent", nil)
	reply := seedComment(t, commentsRepo, post.ID, replyAuthor.ID, "reply", &parent.ID)

	svc := NewService(repo, usersRepo, postsRepo, commentsRepo)
	if err := svc.NotifyCommentReplied(context.Background(), reply.ID.Hex()); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	notifications, _ := repo.ListNotifications(context.Background(), parentAuthor.ID.Hex(), true, 1, 20)
	if len(notifications) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(notifications))
	}
	if notifications[0].Context.ParentComment == nil {
		t.Fatal("expected parent comment context")
	}
}

func seedUser(t *testing.T, repo users.Repository, username string) *users.User {
	t.Helper()
	memoryRepo, ok := repo.(*users.MemoryRepository)
	if !ok {
		t.Fatal("expected memory users repo")
	}
	user := &users.User{
		ID:           bson.NewObjectID(),
		Username:     username,
		PasswordHash: "hash",
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
	memoryRepo.Seed(user)
	return user
}

func seedPost(t *testing.T, repo posts.Repository, authorID, markdown string) *posts.Post {
	t.Helper()
	post, err := repo.CreatePost(context.Background(), &posts.Post{
		ID:              bson.NewObjectID(),
		AuthorID:        mustObjectID(t, authorID),
		ContentMarkdown: markdown,
		ContentHTML:     "<p>" + markdown + "</p>",
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	return post
}

func seedComment(t *testing.T, repo comments.Repository, postID, authorID bson.ObjectID, markdown string, parentID *bson.ObjectID) *comments.Comment {
	t.Helper()
	comment, err := repo.CreateComment(context.Background(), &comments.Comment{
		ID:              bson.NewObjectID(),
		PostID:          postID,
		AuthorID:        authorID,
		ContentMarkdown: markdown,
		ContentHTML:     "<p>" + markdown + "</p>",
		ParentCommentID: parentID,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	return comment
}

func mustObjectID(t *testing.T, value string) bson.ObjectID {
	t.Helper()
	id, err := bson.ObjectIDFromHex(value)
	if err != nil {
		t.Fatalf("expected valid object id, got %v", err)
	}
	return id
}
