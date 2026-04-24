package auth

import (
	"context"
	"testing"

	"decoy-club/backend/internal/users"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestRepositoryCreateAndFindUser(t *testing.T) {
	repo := NewMemoryRepository()

	created, err := repo.CreateUser(context.Background(), &users.User{
		Username:     "bot_bob",
		PasswordHash: "hashed",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if created.ID.IsZero() {
		t.Fatal("expected user ID to be assigned")
	}

	found, err := repo.FindByUsername(context.Background(), "bot_bob")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if found.Username != "bot_bob" || found.PasswordHash != "hashed" {
		t.Fatalf("unexpected user: %+v", found)
	}
}

func TestRepositoryRejectsDuplicateUsername(t *testing.T) {
	repo := NewMemoryRepository()

	_, err := repo.CreateUser(context.Background(), &users.User{
		Username:     "bot_bob",
		PasswordHash: "hashed",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	_, err = repo.CreateUser(context.Background(), &users.User{
		Username:     "bot_bob",
		PasswordHash: "other",
	})
	if err != ErrUsernameTaken {
		t.Fatalf("expected ErrUsernameTaken, got %v", err)
	}
}

func TestRepositoryAssignsIDWhenMissing(t *testing.T) {
	repo := NewMemoryRepository()

	created, err := repo.CreateUser(context.Background(), &users.User{
		Username:     "bot_alice",
		PasswordHash: "hashed",
		ID:           bson.NilObjectID,
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if created.ID.IsZero() {
		t.Fatal("expected assigned object ID")
	}
}
