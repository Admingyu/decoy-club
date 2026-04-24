package auth

import (
	"context"
	"testing"
	"time"

	"decoy-club/backend/internal/users"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestRegisterCreatesUserWithHashedPassword(t *testing.T) {
	repo := newFakeAuthRepo()
	svc := NewService(repo, "secret")

	user, err := svc.Register(context.Background(), RegisterInput{
		Username: "bot_bob",
		Password: "pass1234",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if user.PasswordHash == "" || user.PasswordHash == "pass1234" {
		t.Fatalf("expected hashed password, got %q", user.PasswordHash)
	}
}

func TestLoginReturnsJWTForValidCredentials(t *testing.T) {
	repo := newFakeAuthRepoWithPassword("bot_bob", mustHash("pass1234"))
	svc := NewService(repo, "secret")

	token, _, err := svc.Login(context.Background(), LoginInput{
		Username: "bot_bob",
		Password: "pass1234",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if token == "" {
		t.Fatal("expected token")
	}
}

func TestLoginReturnsInvalidCredentialsForMissingUser(t *testing.T) {
	repo := newFakeAuthRepo()
	svc := NewService(repo, "secret")

	_, _, err := svc.Login(context.Background(), LoginInput{
		Username: "missing",
		Password: "pass1234",
	})
	if err != ErrUserNotFound {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

type fakeAuthRepo struct {
	users map[string]*users.User
}

func newFakeAuthRepo() *fakeAuthRepo {
	return &fakeAuthRepo{users: map[string]*users.User{}}
}

func newFakeAuthRepoWithPassword(username, passwordHash string) *fakeAuthRepo {
	repo := newFakeAuthRepo()
	repo.users[username] = &users.User{
		ID:           mustObjectID("507f1f77bcf86cd799439011"),
		Username:     username,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
	return repo
}

func (r *fakeAuthRepo) CreateUser(_ context.Context, user *users.User) (*users.User, error) {
	if user.ID.IsZero() {
		user.ID = mustObjectID("507f1f77bcf86cd799439011")
	}
	user.CreatedAt = user.CreatedAt.UTC()
	user.UpdatedAt = user.UpdatedAt.UTC()
	r.users[user.Username] = user
	return user, nil
}

func (r *fakeAuthRepo) FindByUsername(_ context.Context, username string) (*users.User, error) {
	user, ok := r.users[username]
	if !ok {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func mustHash(password string) string {
	hash, err := HashPassword(password)
	if err != nil {
		panic(err)
	}
	return hash
}

func mustObjectID(hex string) bson.ObjectID {
	id, err := bson.ObjectIDFromHex(hex)
	if err != nil {
		panic(err)
	}
	return id
}
