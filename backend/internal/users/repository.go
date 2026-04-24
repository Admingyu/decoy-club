package users

import (
	"context"
	"errors"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var ErrUserNotFound = errors.New("user not found")

type Repository interface {
	FindByUsername(ctx context.Context, username string) (*User, error)
	FindByID(ctx context.Context, id string) (*User, error)
	IsFollowing(ctx context.Context, followerID, followeeID string) (bool, error)
	ListFollowingIDs(ctx context.Context, followerID string) ([]string, error)
	RunFollowTransaction(ctx context.Context, followerID, followeeID string) error
	RunUnfollowTransaction(ctx context.Context, followerID, followeeID string) error
}

type MongoRepository struct {
	db           *mongo.Database
	indexesMu    sync.Mutex
	indexesReady bool
}

type followEdge struct {
	ID         bson.ObjectID `bson:"_id,omitempty"`
	FollowerID bson.ObjectID `bson:"follower_id"`
	FolloweeID bson.ObjectID `bson:"followee_id"`
	CreatedAt  time.Time     `bson:"created_at"`
}

func NewMongoRepository(db *mongo.Database) *MongoRepository {
	return &MongoRepository{db: db}
}

func (r *MongoRepository) usersCollection() *mongo.Collection {
	return r.db.Collection("users")
}

func (r *MongoRepository) followsCollection() *mongo.Collection {
	return r.db.Collection("user_follows")
}

func (r *MongoRepository) FindByUsername(ctx context.Context, username string) (*User, error) {
	if err := r.ensureIndexes(ctx); err != nil {
		return nil, err
	}

	var user User
	if err := r.usersCollection().FindOne(ctx, bson.M{"username": username}).Decode(&user); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *MongoRepository) FindByID(ctx context.Context, id string) (*User, error) {
	if err := r.ensureIndexes(ctx); err != nil {
		return nil, err
	}

	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrUserNotFound
	}

	var user User
	if err := r.usersCollection().FindOne(ctx, bson.M{"_id": objectID}).Decode(&user); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *MongoRepository) IsFollowing(ctx context.Context, followerID, followeeID string) (bool, error) {
	if err := r.ensureIndexes(ctx); err != nil {
		return false, err
	}

	followerObjectID, err := bson.ObjectIDFromHex(followerID)
	if err != nil {
		return false, ErrUserNotFound
	}
	followeeObjectID, err := bson.ObjectIDFromHex(followeeID)
	if err != nil {
		return false, ErrUserNotFound
	}

	err = r.followsCollection().FindOne(ctx, bson.M{
		"follower_id": followerObjectID,
		"followee_id": followeeObjectID,
	}).Err()
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (r *MongoRepository) ListFollowingIDs(ctx context.Context, followerID string) ([]string, error) {
	if err := r.ensureIndexes(ctx); err != nil {
		return nil, err
	}

	followerObjectID, err := bson.ObjectIDFromHex(followerID)
	if err != nil {
		return []string{}, nil
	}

	cursor, err := r.followsCollection().Find(ctx, bson.M{"follower_id": followerObjectID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	followingIDs := make([]string, 0)
	for cursor.Next(ctx) {
		var edge followEdge
		if err := cursor.Decode(&edge); err != nil {
			return nil, err
		}
		followingIDs = append(followingIDs, edge.FolloweeID.Hex())
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return followingIDs, nil
}

func (r *MongoRepository) RunFollowTransaction(ctx context.Context, followerID, followeeID string) error {
	if err := r.ensureIndexes(ctx); err != nil {
		return err
	}

	followerObjectID, err := bson.ObjectIDFromHex(followerID)
	if err != nil {
		return ErrUserNotFound
	}
	followeeObjectID, err := bson.ObjectIDFromHex(followeeID)
	if err != nil {
		return ErrUserNotFound
	}

	now := time.Now().UTC()
	edge := followEdge{
		ID:         bson.NewObjectID(),
		FollowerID: followerObjectID,
		FolloweeID: followeeObjectID,
		CreatedAt:  now,
	}
	if _, err := r.followsCollection().InsertOne(ctx, edge); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil
		}
		return err
	}

	if err := r.adjustFollowCounters(ctx, followerObjectID, followeeObjectID, 1, now); err != nil {
		_, _ = r.followsCollection().DeleteOne(ctx, bson.M{
			"follower_id": followerObjectID,
			"followee_id": followeeObjectID,
		})
		_ = r.adjustFollowCounters(ctx, followerObjectID, followeeObjectID, -1, now)
		return err
	}

	return nil
}

func (r *MongoRepository) RunUnfollowTransaction(ctx context.Context, followerID, followeeID string) error {
	if err := r.ensureIndexes(ctx); err != nil {
		return err
	}

	followerObjectID, err := bson.ObjectIDFromHex(followerID)
	if err != nil {
		return ErrUserNotFound
	}
	followeeObjectID, err := bson.ObjectIDFromHex(followeeID)
	if err != nil {
		return ErrUserNotFound
	}

	filter := bson.M{
		"follower_id": followerObjectID,
		"followee_id": followeeObjectID,
	}
	var removed followEdge
	if err := r.followsCollection().FindOne(ctx, filter).Decode(&removed); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil
		}
		return err
	}

	result, err := r.followsCollection().DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return nil
	}

	now := time.Now().UTC()
	if err := r.adjustFollowCounters(ctx, followerObjectID, followeeObjectID, -1, now); err != nil {
		_, _ = r.followsCollection().InsertOne(ctx, removed)
		_ = r.adjustFollowCounters(ctx, followerObjectID, followeeObjectID, 1, now)
		return err
	}

	return nil
}

func (r *MongoRepository) adjustFollowCounters(ctx context.Context, followerObjectID, followeeObjectID bson.ObjectID, delta int64, now time.Time) error {
	if _, err := r.usersCollection().UpdateByID(ctx, followerObjectID, bson.M{
		"$inc": bson.M{"following_count": delta},
		"$set": bson.M{"updated_at": now},
	}); err != nil {
		return err
	}

	if _, err := r.usersCollection().UpdateByID(ctx, followeeObjectID, bson.M{
		"$inc": bson.M{"followers_count": delta},
		"$set": bson.M{"updated_at": now},
	}); err != nil {
		return err
	}

	return nil
}

func (r *MongoRepository) ensureIndexes(ctx context.Context) error {
	r.indexesMu.Lock()
	defer r.indexesMu.Unlock()

	if r.indexesReady {
		return nil
	}
	if r.db == nil {
		return errors.New("mongo database is nil")
	}

	if _, err := r.usersCollection().Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "username", Value: 1}},
		Options: options.Index().SetUnique(true),
	}); err != nil {
		return err
	}

	if _, err := r.followsCollection().Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "follower_id", Value: 1},
			{Key: "followee_id", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	}); err != nil {
		return err
	}

	r.indexesReady = true
	return nil
}

type MemoryRepository struct {
	mu        sync.RWMutex
	users     map[string]*User
	usersByID map[string]*User
	follows   map[string]map[string]struct{}
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		users:     make(map[string]*User),
		usersByID: make(map[string]*User),
		follows:   make(map[string]map[string]struct{}),
	}
}

func (r *MemoryRepository) Seed(user *User) {
	if user == nil {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	cp := *user
	if cp.ID.IsZero() {
		cp.ID = bson.NewObjectID()
	}
	r.users[cp.Username] = &cp
	r.usersByID[cp.ID.Hex()] = &cp
}

func (r *MemoryRepository) ensureUser(username string) *User {
	user, ok := r.users[username]
	if !ok {
		user = &User{
			Username: username,
			ID:       bson.NewObjectID(),
		}
		r.users[username] = user
		r.usersByID[user.ID.Hex()] = user
	}
	return user
}

func (r *MemoryRepository) FindByUsername(_ context.Context, username string) (*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.users[username]
	if !ok {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (r *MemoryRepository) FindByID(_ context.Context, id string) (*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.usersByID[id]
	if ok {
		return user, nil
	}

	return nil, ErrUserNotFound
}

func (r *MemoryRepository) IsFollowing(_ context.Context, followerID, followeeID string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	edges, ok := r.follows[followerID]
	if !ok {
		return false, nil
	}
	_, ok = edges[followeeID]
	return ok, nil
}

func (r *MemoryRepository) ListFollowingIDs(_ context.Context, followerID string) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

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

func (r *MemoryRepository) RunFollowTransaction(_ context.Context, followerID, followeeID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if followerID == followeeID {
		return nil
	}
	if _, ok := r.follows[followerID]; !ok {
		r.follows[followerID] = make(map[string]struct{})
	}
	if _, exists := r.follows[followerID][followeeID]; exists {
		return nil
	}

	follower := r.ensureUserByID(followerID)
	followee := r.ensureUserByID(followeeID)
	r.follows[followerID][followeeID] = struct{}{}
	follower.FollowingCount++
	followee.FollowersCount++
	return nil
}

func (r *MemoryRepository) RunUnfollowTransaction(_ context.Context, followerID, followeeID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	edges, ok := r.follows[followerID]
	if !ok {
		return nil
	}
	if _, exists := edges[followeeID]; !exists {
		return nil
	}

	delete(edges, followeeID)
	follower := r.ensureUserByID(followerID)
	followee := r.ensureUserByID(followeeID)
	if follower.FollowingCount > 0 {
		follower.FollowingCount--
	}
	if followee.FollowersCount > 0 {
		followee.FollowersCount--
	}
	return nil
}

func (r *MemoryRepository) ensureUserByID(id string) *User {
	if user, ok := r.usersByID[id]; ok {
		return user
	}

	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		objectID = bson.NewObjectID()
	}
	user := &User{ID: objectID}
	r.usersByID[user.ID.Hex()] = user
	return user
}
