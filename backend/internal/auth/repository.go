package auth

import (
	"context"
	"errors"
	"sync"
	"time"

	"decoy-club/backend/internal/users"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var ErrUserNotFound = errors.New("user not found")
var ErrUsernameTaken = errors.New("username already taken")

type Repository interface {
	CreateUser(ctx context.Context, user *users.User) (*users.User, error)
	FindByUsername(ctx context.Context, username string) (*users.User, error)
}

type MongoRepository struct {
	collection *mongo.Collection
}

func NewMongoRepository(db *mongo.Database) *MongoRepository {
	return &MongoRepository{
		collection: db.Collection("users"),
	}
}

func (r *MongoRepository) CreateUser(ctx context.Context, user *users.User) (*users.User, error) {
	if user.ID.IsZero() {
		user.ID = bson.NewObjectID()
	}
	now := time.Now().UTC()
	if user.CreatedAt.IsZero() {
		user.CreatedAt = now
	}
	if user.UpdatedAt.IsZero() {
		user.UpdatedAt = now
	}

	existing, err := r.FindByUsername(ctx, user.Username)
	switch {
	case err == nil && existing != nil:
		return nil, ErrUsernameTaken
	case err != nil && !errors.Is(err, ErrUserNotFound):
		return nil, err
	}

	if _, err := r.collection.InsertOne(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (r *MongoRepository) FindByUsername(ctx context.Context, username string) (*users.User, error) {
	var user users.User
	if err := r.collection.FindOne(ctx, bson.M{"username": username}).Decode(&user); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

type MemoryRepository struct {
	mu    sync.RWMutex
	users map[string]*users.User
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		users: make(map[string]*users.User),
	}
}

func (r *MemoryRepository) CreateUser(_ context.Context, user *users.User) (*users.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[user.Username]; exists {
		return nil, ErrUsernameTaken
	}
	if user.ID.IsZero() {
		user.ID = bson.NewObjectID()
	}

	r.users[user.Username] = user
	return user, nil
}

func (r *MemoryRepository) FindByUsername(_ context.Context, username string) (*users.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.users[username]
	if !ok {
		return nil, ErrUserNotFound
	}

	return user, nil
}
