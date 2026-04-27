package users

import (
	"context"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestFollowUpdatesCounters(t *testing.T) {
	repo := newFakeUserRepo()
	svc := NewService(repo)

	err := svc.Follow(context.Background(), "viewer-id", "author-id")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if repo.followingCount != 1 || repo.followersCount != 1 {
		t.Fatalf("expected counters to increment, got %+v", repo)
	}
}

func TestCannotFollowSelf(t *testing.T) {
	repo := newFakeUserRepo()
	svc := NewService(repo)

	err := svc.Follow(context.Background(), "same-id", "same-id")
	if err == nil {
		t.Fatal("expected self-follow error")
	}
}

func TestFollowIsIdempotentForDuplicateRequests(t *testing.T) {
	repo := newFakeUserRepo()
	svc := NewService(repo)

	if err := svc.Follow(context.Background(), "viewer-id", "author-id"); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if err := svc.Follow(context.Background(), "viewer-id", "author-id"); err != nil {
		t.Fatalf("expected nil error on duplicate follow, got %v", err)
	}

	if repo.followingCount != 1 || repo.followersCount != 1 {
		t.Fatalf("expected duplicate follow to be idempotent, got %+v", repo)
	}
}

func TestUpdateStatusTrimsAndPersistsStatus(t *testing.T) {
	repo := newFakeUserRepo()
	svc := NewService(repo)

	profile, err := svc.UpdateStatus(context.Background(), "507f1f77bcf86cd799439011", "  building tonight  ", "  coding  ")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if profile.StatusText != "building tonight" || profile.StatusPreset != "coding" {
		t.Fatalf("expected trimmed status to be saved, got %+v", profile)
	}
}

func TestUpdateStatusRejectsTooLongText(t *testing.T) {
	repo := newFakeUserRepo()
	svc := NewService(repo)

	_, err := svc.UpdateStatus(context.Background(), "507f1f77bcf86cd799439011", string(make([]byte, 81)), "")
	if err != ErrStatusTooLong {
		t.Fatalf("expected ErrStatusTooLong, got %v", err)
	}
}

type fakeUserRepo struct {
	users          map[string]*User
	usersByID      map[string]*User
	follows        map[string]map[string]struct{}
	followingCount int64
	followersCount int64
}

func newFakeUserRepo() *fakeUserRepo {
	viewerID := mustObjectID("507f1f77bcf86cd799439011")
	authorID := mustObjectID("507f1f77bcf86cd799439012")

	viewer := &User{ID: viewerID, Username: "viewer"}
	author := &User{ID: authorID, Username: "author"}

	return &fakeUserRepo{
		users: map[string]*User{
			viewer.Username: viewer,
			author.Username: author,
		},
		usersByID: map[string]*User{
			viewer.ID.Hex(): viewer,
			author.ID.Hex(): author,
		},
		follows: make(map[string]map[string]struct{}),
	}
}

func (r *fakeUserRepo) FindByUsername(_ context.Context, username string) (*User, error) {
	user, ok := r.users[username]
	if !ok {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (r *fakeUserRepo) FindByID(_ context.Context, id string) (*User, error) {
	user, ok := r.usersByID[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (r *fakeUserRepo) IsFollowing(_ context.Context, followerID, followeeID string) (bool, error) {
	edges, ok := r.follows[followerID]
	if !ok {
		return false, nil
	}
	_, ok = edges[followeeID]
	return ok, nil
}

func (r *fakeUserRepo) ListFollowingIDs(_ context.Context, followerID string) ([]string, error) {
	edges, ok := r.follows[followerID]
	if !ok {
		return []string{}, nil
	}

	followingIDs := make([]string, 0, len(edges))
	for followeeID := range edges {
		followingIDs = append(followingIDs, followeeID)
	}

	return followingIDs, nil
}

func (r *fakeUserRepo) RunFollowTransaction(_ context.Context, followerID, followeeID string) error {
	if followerID == followeeID {
		return ErrCannotFollowSelf
	}

	if _, ok := r.follows[followerID]; !ok {
		r.follows[followerID] = make(map[string]struct{})
	}
	if _, exists := r.follows[followerID][followeeID]; exists {
		return nil
	}

	r.follows[followerID][followeeID] = struct{}{}
	r.followingCount++
	r.followersCount++
	return nil
}

func (r *fakeUserRepo) RunUnfollowTransaction(_ context.Context, followerID, followeeID string) error {
	edges, ok := r.follows[followerID]
	if !ok {
		return nil
	}
	if _, exists := edges[followeeID]; !exists {
		return nil
	}

	delete(edges, followeeID)
	if r.followingCount > 0 {
		r.followingCount--
	}
	if r.followersCount > 0 {
		r.followersCount--
	}
	return nil
}

func (r *fakeUserRepo) UpdateStatus(_ context.Context, userID, statusText, statusPreset string) (*User, error) {
	user, ok := r.usersByID[userID]
	if !ok {
		return nil, ErrUserNotFound
	}
	user.StatusText = statusText
	user.StatusPreset = statusPreset
	return user, nil
}

var _ Repository = (*fakeUserRepo)(nil)

func mustObjectID(hex string) bson.ObjectID {
	id, err := bson.ObjectIDFromHex(hex)
	if err != nil {
		panic(err)
	}
	return id
}
