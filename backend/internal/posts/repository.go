package posts

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"

	"decoy-club/backend/internal/users"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MongoRepository struct {
	db           *mongo.Database
	indexesMu    sync.Mutex
	indexesReady bool
}

func NewMongoRepository(db *mongo.Database) *MongoRepository {
	return &MongoRepository{db: db}
}

func (r *MongoRepository) postsCollection() *mongo.Collection {
	return r.db.Collection("posts")
}

func (r *MongoRepository) usersCollection() *mongo.Collection {
	return r.db.Collection("users")
}

func (r *MongoRepository) CreatePost(ctx context.Context, post *Post) (*Post, error) {
	if err := r.ensureIndexes(ctx); err != nil {
		return nil, err
	}

	return r.runTransaction(ctx, func(sessionCtx context.Context) (*Post, error) {
		now := time.Now().UTC()
		if post.ID.IsZero() {
			post.ID = bson.NewObjectID()
		}
		if post.CreatedAt.IsZero() {
			post.CreatedAt = now
		}
		if post.UpdatedAt.IsZero() {
			post.UpdatedAt = now
		}

		if _, err := r.postsCollection().InsertOne(sessionCtx, post); err != nil {
			return nil, err
		}

		if !post.AuthorID.IsZero() {
			result, err := r.usersCollection().UpdateOne(sessionCtx, bson.M{"_id": post.AuthorID}, bson.M{
				"$inc": bson.M{"post_count": int64(1)},
				"$set": bson.M{"updated_at": now},
			})
			if err != nil {
				return nil, err
			}
			if result.MatchedCount == 0 {
				return nil, users.ErrUserNotFound
			}
		}

		return post, nil
	})
}

func (r *MongoRepository) FindPostByID(ctx context.Context, id string) (*Post, error) {
	if err := r.ensureIndexes(ctx); err != nil {
		return nil, err
	}

	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrPostNotFound
	}

	var post Post
	if err := r.postsCollection().FindOne(ctx, bson.M{
		"_id":        objectID,
		"is_deleted": false,
	}).Decode(&post); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrPostNotFound
		}
		return nil, err
	}

	return &post, nil
}

func (r *MongoRepository) ListPublicTimeline(ctx context.Context, limit int, before *time.Time) ([]*Post, error) {
	if err := r.ensureIndexes(ctx); err != nil {
		return nil, err
	}

	if limit <= 0 {
		limit = 20
	}

	filter := bson.M{"is_deleted": false}
	if before != nil && !before.IsZero() {
		filter["created_at"] = bson.M{"$lt": *before}
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}, {Key: "_id", Value: -1}}).
		SetLimit(int64(limit))

	cursor, err := r.postsCollection().Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var posts []*Post
	if err := cursor.All(ctx, &posts); err != nil {
		return nil, err
	}

	return posts, nil
}

func (r *MongoRepository) SoftDeletePost(ctx context.Context, postID, authorID string) error {
	if err := r.ensureIndexes(ctx); err != nil {
		return err
	}

	postObjectID, err := bson.ObjectIDFromHex(postID)
	if err != nil {
		return ErrPostNotFound
	}
	authorObjectID, err := bson.ObjectIDFromHex(authorID)
	if err != nil {
		return ErrPostNotFound
	}

	now := time.Now().UTC()
	result, err := r.postsCollection().UpdateOne(ctx, bson.M{
		"_id":        postObjectID,
		"author_id":  authorObjectID,
		"is_deleted": false,
	}, bson.M{
		"$set": bson.M{
			"is_deleted": true,
			"deleted_at": now,
			"updated_at": now,
		},
	})
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrPostNotFound
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

	if _, err := r.postsCollection().Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "is_deleted", Value: 1},
			{Key: "created_at", Value: -1},
		},
	}); err != nil {
		return err
	}

	if _, err := r.postsCollection().Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "author_id", Value: 1},
			{Key: "created_at", Value: -1},
		},
	}); err != nil {
		return err
	}

	r.indexesReady = true
	return nil
}

func (r *MongoRepository) runTransaction(ctx context.Context, fn func(sessionCtx context.Context) (*Post, error)) (*Post, error) {
	if r.db == nil {
		return nil, errors.New("mongo database is nil")
	}

	client := r.db.Client()
	if client == nil {
		return nil, errors.New("mongo client is nil")
	}

	session, err := client.StartSession()
	if err != nil {
		return nil, err
	}
	defer session.EndSession(ctx)

	var created *Post
	_, err = session.WithTransaction(ctx, func(sessionCtx context.Context) (any, error) {
		var innerErr error
		created, innerErr = fn(sessionCtx)
		if innerErr != nil {
			return nil, innerErr
		}
		return nil, nil
	})
	if err != nil {
		return nil, err
	}

	return created, nil
}

type MemoryRepository struct {
	mu    sync.RWMutex
	posts map[string]*Post
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		posts: make(map[string]*Post),
	}
}

func (r *MemoryRepository) CreatePost(_ context.Context, post *Post) (*Post, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now().UTC()
	if post.ID.IsZero() {
		post.ID = bson.NewObjectID()
	}
	if post.CreatedAt.IsZero() {
		post.CreatedAt = now
	}
	if post.UpdatedAt.IsZero() {
		post.UpdatedAt = now
	}

	cp := *post
	r.posts[post.ID.Hex()] = &cp
	return &cp, nil
}

func (r *MemoryRepository) FindPostByID(_ context.Context, id string) (*Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	post, ok := r.posts[id]
	if !ok || post.IsDeleted {
		return nil, ErrPostNotFound
	}

	cp := *post
	return &cp, nil
}

func (r *MemoryRepository) ListPublicTimeline(_ context.Context, limit int, before *time.Time) ([]*Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if limit <= 0 {
		limit = 20
	}

	posts := make([]*Post, 0, len(r.posts))
	for _, post := range r.posts {
		if post.IsDeleted {
			continue
		}
		if before != nil && !before.IsZero() && !post.CreatedAt.Before(*before) {
			continue
		}
		cp := *post
		posts = append(posts, &cp)
	}

	sort.Slice(posts, func(i, j int) bool {
		if posts[i].CreatedAt.Equal(posts[j].CreatedAt) {
			return posts[i].ID.Hex() > posts[j].ID.Hex()
		}
		return posts[i].CreatedAt.After(posts[j].CreatedAt)
	})

	if len(posts) > limit {
		posts = posts[:limit]
	}

	return posts, nil
}

func (r *MemoryRepository) SoftDeletePost(_ context.Context, postID, authorID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	post, ok := r.posts[postID]
	if !ok {
		return ErrPostNotFound
	}
	if post.AuthorID.Hex() != authorID {
		return ErrPostNotFound
	}

	now := time.Now().UTC()
	post.IsDeleted = true
	post.DeletedAt = &now
	post.UpdatedAt = now
	return nil
}

var _ Repository = (*MongoRepository)(nil)
var _ Repository = (*MemoryRepository)(nil)
