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
	userRepo     users.Repository
	indexesMu    sync.Mutex
	indexesReady bool
}

type likeEdge struct {
	ID        bson.ObjectID `bson:"_id,omitempty"`
	PostID    bson.ObjectID `bson:"post_id"`
	UserID    bson.ObjectID `bson:"user_id"`
	CreatedAt time.Time     `bson:"created_at"`
}

func NewMongoRepository(db *mongo.Database, userRepo users.Repository) *MongoRepository {
	return &MongoRepository{db: db, userRepo: userRepo}
}

func (r *MongoRepository) postsCollection() *mongo.Collection {
	return r.db.Collection("posts")
}

func (r *MongoRepository) usersCollection() *mongo.Collection {
	return r.db.Collection("users")
}

func (r *MongoRepository) likesCollection() *mongo.Collection {
	return r.db.Collection("post_likes")
}

func (r *MongoRepository) topicsCollection() *mongo.Collection {
	return r.db.Collection("topics")
}

func (r *MongoRepository) AdjustCommentCount(ctx context.Context, postID string, delta int64) error {
	if err := r.ensureIndexes(ctx); err != nil {
		return err
	}

	postObjectID, err := bson.ObjectIDFromHex(postID)
	if err != nil {
		return ErrPostNotFound
	}

	now := time.Now().UTC()
	result, err := r.postsCollection().UpdateOne(ctx, bson.M{
		"_id":        postObjectID,
		"is_deleted": false,
	}, bson.M{
		"$inc": bson.M{"comment_count": delta},
		"$set": bson.M{"updated_at": now},
	})
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrPostNotFound
	}
	return nil
}

func (r *MongoRepository) CreatePost(ctx context.Context, post *Post) (*Post, error) {
	if err := r.ensureIndexes(ctx); err != nil {
		return nil, err
	}

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

	if !post.AuthorID.IsZero() {
		result, err := r.usersCollection().UpdateOne(ctx, bson.M{"_id": post.AuthorID}, bson.M{
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

	if _, err := r.postsCollection().InsertOne(ctx, post); err != nil {
		if !post.AuthorID.IsZero() {
			_, _ = r.usersCollection().UpdateOne(ctx, bson.M{"_id": post.AuthorID}, bson.M{
				"$inc": bson.M{"post_count": int64(-1)},
				"$set": bson.M{"updated_at": now},
			})
		}
		return nil, err
	}
	if err := r.adjustTopics(ctx, post.Topics, 1, post.CreatedAt); err != nil {
		_, _ = r.postsCollection().DeleteOne(ctx, bson.M{"_id": post.ID})
		return nil, err
	}

	return post, nil
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

func (r *MongoRepository) ListPublicTimeline(ctx context.Context, limit int, before *time.Time, beforeID string) ([]*Post, error) {
	if err := r.ensureIndexes(ctx); err != nil {
		return nil, err
	}

	if limit <= 0 {
		limit = 20
	}

	filter := bson.M{"is_deleted": false}
	if beforeID != "" && (before == nil || before.IsZero()) {
		return nil, ErrInvalidTimelineCursor
	}
	if before != nil && !before.IsZero() {
		if beforeID != "" {
			beforeObjectID, err := bson.ObjectIDFromHex(beforeID)
			if err != nil {
				return nil, ErrPostNotFound
			}
			filter["$or"] = bson.A{
				bson.M{"created_at": bson.M{"$lt": *before}},
				bson.M{
					"created_at": *before,
					"_id":        bson.M{"$lt": beforeObjectID},
				},
			}
		} else {
			filter["created_at"] = bson.M{"$lt": *before}
		}
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
	var post Post
	if err := r.postsCollection().FindOne(ctx, bson.M{"_id": postObjectID}).Decode(&post); err == nil {
		_ = r.adjustTopics(ctx, post.Topics, -1, now)
	}

	return nil
}

func (r *MongoRepository) ListTrendingTopics(ctx context.Context, limit int) ([]*Topic, error) {
	if err := r.ensureIndexes(ctx); err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 10
	}

	cursor, err := r.topicsCollection().Find(ctx, bson.M{"post_count": bson.M{"$gt": 0}}, options.Find().
		SetSort(bson.D{{Key: "post_count", Value: -1}, {Key: "last_post_at", Value: -1}, {Key: "name", Value: 1}}).
		SetLimit(int64(limit)))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var topics []*Topic
	if err := cursor.All(ctx, &topics); err != nil {
		return nil, err
	}
	return topics, nil
}

func (r *MongoRepository) ListFollowingTimeline(ctx context.Context, viewerID string, page, size int) ([]*Post, error) {
	if err := r.ensureIndexes(ctx); err != nil {
		return nil, err
	}

	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}

	if r.userRepo == nil {
		return []*Post{}, nil
	}

	followeeIDs, err := r.userRepo.ListFollowingIDs(ctx, viewerID)
	if err != nil {
		return nil, err
	}
	if len(followeeIDs) == 0 {
		return []*Post{}, nil
	}

	followeeObjectIDs := make([]bson.ObjectID, 0, len(followeeIDs))
	for _, followeeID := range followeeIDs {
		objectID, err := bson.ObjectIDFromHex(followeeID)
		if err != nil {
			continue
		}
		followeeObjectIDs = append(followeeObjectIDs, objectID)
	}
	if len(followeeObjectIDs) == 0 {
		return []*Post{}, nil
	}

	filter := bson.M{
		"is_deleted": false,
		"author_id":  bson.M{"$in": followeeObjectIDs},
	}
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}, {Key: "_id", Value: -1}}).
		SetSkip(int64((page - 1) * size)).
		SetLimit(int64(size))

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

func (r *MongoRepository) ListPostsByAuthorID(ctx context.Context, authorID string, limit int) ([]*Post, error) {
	if err := r.ensureIndexes(ctx); err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 20
	}

	authorObjectID, err := bson.ObjectIDFromHex(authorID)
	if err != nil {
		return []*Post{}, nil
	}

	cursor, err := r.postsCollection().Find(ctx, bson.M{
		"author_id":  authorObjectID,
		"is_deleted": false,
	}, options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}, {Key: "_id", Value: -1}}).
		SetLimit(int64(limit)))
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

func (r *MongoRepository) HasLike(ctx context.Context, postID, userID string) (bool, error) {
	if err := r.ensureIndexes(ctx); err != nil {
		return false, err
	}
	postObjectID, err := bson.ObjectIDFromHex(postID)
	if err != nil {
		return false, nil
	}
	userObjectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return false, nil
	}
	err = r.likesCollection().FindOne(ctx, bson.M{"post_id": postObjectID, "user_id": userObjectID}).Err()
	if err == nil {
		return true, nil
	}
	if errors.Is(err, mongo.ErrNoDocuments) {
		return false, nil
	}
	return false, err
}

func (r *MongoRepository) ListLikedPostIDs(ctx context.Context, userID string, postIDs []string) (map[string]bool, error) {
	if err := r.ensureIndexes(ctx); err != nil {
		return nil, err
	}
	if len(postIDs) == 0 {
		return map[string]bool{}, nil
	}
	userObjectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return map[string]bool{}, nil
	}
	postObjectIDs := make([]bson.ObjectID, 0, len(postIDs))
	for _, postID := range postIDs {
		objectID, err := bson.ObjectIDFromHex(postID)
		if err != nil {
			continue
		}
		postObjectIDs = append(postObjectIDs, objectID)
	}
	if len(postObjectIDs) == 0 {
		return map[string]bool{}, nil
	}
	cursor, err := r.likesCollection().Find(ctx, bson.M{
		"user_id": userObjectID,
		"post_id": bson.M{"$in": postObjectIDs},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	result := make(map[string]bool)
	for cursor.Next(ctx) {
		var edge likeEdge
		if err := cursor.Decode(&edge); err != nil {
			return nil, err
		}
		result[edge.PostID.Hex()] = true
	}
	return result, cursor.Err()
}

func (r *MongoRepository) CreateLikeIfAbsent(ctx context.Context, postID, userID string) (bool, error) {
	if err := r.ensureIndexes(ctx); err != nil {
		return false, err
	}

	postObjectID, err := bson.ObjectIDFromHex(postID)
	if err != nil {
		return false, ErrPostNotFound
	}
	userObjectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return false, ErrPostNotFound
	}

	err = r.likesCollection().FindOne(ctx, bson.M{
		"post_id": postObjectID,
		"user_id": userObjectID,
	}).Err()
	if err == nil {
		return false, nil
	}
	if !errors.Is(err, mongo.ErrNoDocuments) {
		return false, err
	}

	_, err = r.likesCollection().InsertOne(ctx, likeEdge{
		ID:        bson.NewObjectID(),
		PostID:    postObjectID,
		UserID:    userObjectID,
		CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (r *MongoRepository) RemoveLikeIfPresent(ctx context.Context, postID, userID string) (bool, error) {
	if err := r.ensureIndexes(ctx); err != nil {
		return false, err
	}

	postObjectID, err := bson.ObjectIDFromHex(postID)
	if err != nil {
		return false, ErrPostNotFound
	}
	userObjectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return false, ErrPostNotFound
	}

	result, err := r.likesCollection().DeleteOne(ctx, bson.M{
		"post_id": postObjectID,
		"user_id": userObjectID,
	})
	if err != nil {
		return false, err
	}

	return result.DeletedCount > 0, nil
}

func (r *MongoRepository) LikePost(ctx context.Context, postID, userID string) (bool, error) {
	if err := r.ensureIndexes(ctx); err != nil {
		return false, err
	}

	changed, err := r.CreateLikeIfAbsent(ctx, postID, userID)
	if err != nil || !changed {
		return changed, err
	}

	if err := r.updateLikeCounters(ctx, postID, userID, 1); err != nil {
		_, _ = r.RemoveLikeIfPresent(ctx, postID, userID)
		_ = r.updateLikeCounters(ctx, postID, userID, -1)
		return false, err
	}

	return true, nil
}

func (r *MongoRepository) UnlikePost(ctx context.Context, postID, userID string) (bool, error) {
	if err := r.ensureIndexes(ctx); err != nil {
		return false, err
	}

	postObjectID, err := bson.ObjectIDFromHex(postID)
	if err != nil {
		return false, ErrPostNotFound
	}
	userObjectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return false, ErrPostNotFound
	}

	filter := bson.M{"post_id": postObjectID, "user_id": userObjectID}
	var removed likeEdge
	if err := r.likesCollection().FindOne(ctx, filter).Decode(&removed); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return false, nil
		}
		return false, err
	}

	result, err := r.likesCollection().DeleteOne(ctx, filter)
	if err != nil {
		return false, err
	}
	if result.DeletedCount == 0 {
		return false, nil
	}

	if err := r.updateLikeCounters(ctx, postID, userID, -1); err != nil {
		_, _ = r.likesCollection().InsertOne(ctx, removed)
		_ = r.updateLikeCounters(ctx, postID, userID, 1)
		return false, err
	}

	return true, nil
}

func (r *MongoRepository) ApplyLikeSideEffects(ctx context.Context, postID, userID string) error {
	return r.updateLikeCounters(ctx, postID, userID, 1)
}

func (r *MongoRepository) RevertLikeSideEffects(ctx context.Context, postID, userID string) error {
	return r.updateLikeCounters(ctx, postID, userID, -1)
}

func (r *MongoRepository) updateLikeCounters(ctx context.Context, postID, userID string, delta int64) error {
	postObjectID, err := bson.ObjectIDFromHex(postID)
	if err != nil {
		return ErrPostNotFound
	}
	userObjectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return ErrPostNotFound
	}

	now := time.Now().UTC()
	postResult, err := r.postsCollection().UpdateOne(ctx, bson.M{
		"_id":        postObjectID,
		"is_deleted": false,
	}, bson.M{
		"$inc": bson.M{"like_count": delta},
		"$set": bson.M{"updated_at": now},
	})
	if err != nil {
		return err
	}
	if postResult.MatchedCount == 0 {
		return ErrPostNotFound
	}

	userResult, err := r.usersCollection().UpdateOne(ctx, bson.M{"_id": userObjectID}, bson.M{
		"$inc": bson.M{"given_like_count": delta},
		"$set": bson.M{"updated_at": now},
	})
	if err != nil {
		return err
	}
	if userResult.MatchedCount == 0 {
		return users.ErrUserNotFound
	}

	var post Post
	if err := r.postsCollection().FindOne(ctx, bson.M{
		"_id":        postObjectID,
		"is_deleted": false,
	}).Decode(&post); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return ErrPostNotFound
		}
		return err
	}

	authorResult, err := r.usersCollection().UpdateOne(ctx, bson.M{"_id": post.AuthorID}, bson.M{
		"$inc": bson.M{"received_like_count": delta},
		"$set": bson.M{"updated_at": now},
	})
	if err != nil {
		return err
	}
	if authorResult.MatchedCount == 0 {
		return users.ErrUserNotFound
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

	if _, err := r.likesCollection().Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "post_id", Value: 1},
			{Key: "user_id", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	}); err != nil {
		return err
	}

	if _, err := r.postsCollection().Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "topics", Value: 1}},
	}); err != nil {
		return err
	}

	if _, err := r.topicsCollection().Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "name", Value: 1}},
		Options: options.Index().SetUnique(true),
	}); err != nil {
		return err
	}

	r.indexesReady = true
	return nil
}

func (r *MongoRepository) adjustTopics(ctx context.Context, topics []string, delta int64, lastPostAt time.Time) error {
	if len(topics) == 0 {
		return nil
	}
	now := time.Now().UTC()
	for _, topic := range topics {
		if topic == "" {
			continue
		}
		update := bson.M{
			"$inc": bson.M{"post_count": delta},
			"$set": bson.M{"updated_at": now},
			"$setOnInsert": bson.M{
				"_id":        bson.NewObjectID(),
				"name":       topic,
				"created_at": now,
			},
		}
		if delta > 0 {
			update["$max"] = bson.M{"last_post_at": lastPostAt}
		}
		opts := options.UpdateOne()
		if delta > 0 {
			opts.SetUpsert(true)
		}
		if _, err := r.topicsCollection().UpdateOne(ctx, bson.M{"name": topic}, update, opts); err != nil {
			return err
		}
	}
	return nil
}

type MemoryRepository struct {
	mu       sync.RWMutex
	posts    map[string]*Post
	topics   map[string]*Topic
	likes    map[string]map[string]struct{}
	userRepo users.Repository
}

func NewMemoryRepository(userRepo users.Repository) *MemoryRepository {
	return &MemoryRepository{
		posts:    make(map[string]*Post),
		topics:   make(map[string]*Topic),
		likes:    make(map[string]map[string]struct{}),
		userRepo: userRepo,
	}
}

func (r *MemoryRepository) AdjustCommentCount(_ context.Context, postID string, delta int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	post, ok := r.posts[postID]
	if !ok || post.IsDeleted {
		return ErrPostNotFound
	}
	post.CommentCount += delta
	if post.CommentCount < 0 {
		post.CommentCount = 0
	}
	post.UpdatedAt = time.Now().UTC()
	return nil
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
	cp.Topics = append([]string{}, post.Topics...)
	r.posts[post.ID.Hex()] = &cp
	r.adjustTopicsLocked(cp.Topics, 1, cp.CreatedAt)
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

func (r *MemoryRepository) ListPublicTimeline(_ context.Context, limit int, before *time.Time, beforeID string) ([]*Post, error) {
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
		if before != nil && !before.IsZero() {
			if post.CreatedAt.After(*before) {
				continue
			}
			if post.CreatedAt.Equal(*before) {
				if beforeID == "" || post.ID.Hex() >= beforeID {
					continue
				}
			}
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

func (r *MemoryRepository) ListFollowingTimeline(ctx context.Context, viewerID string, page, size int) ([]*Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}

	if r.userRepo == nil {
		return []*Post{}, nil
	}

	followeeIDs, err := r.userRepo.ListFollowingIDs(ctx, viewerID)
	if err != nil {
		return nil, err
	}
	followees := make(map[string]struct{}, len(followeeIDs))
	for _, followeeID := range followeeIDs {
		followees[followeeID] = struct{}{}
	}

	posts := make([]*Post, 0, len(r.posts))
	for _, post := range r.posts {
		if post.IsDeleted {
			continue
		}
		if _, ok := followees[post.AuthorID.Hex()]; !ok {
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

	start := (page - 1) * size
	if start >= len(posts) {
		return []*Post{}, nil
	}

	end := start + size
	if end > len(posts) {
		end = len(posts)
	}

	return posts[start:end], nil
}

func (r *MemoryRepository) ListPostsByAuthorID(_ context.Context, authorID string, limit int) ([]*Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if limit <= 0 {
		limit = 20
	}

	posts := make([]*Post, 0, len(r.posts))
	for _, post := range r.posts {
		if post.IsDeleted || post.AuthorID.Hex() != authorID {
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

func (r *MemoryRepository) HasLike(_ context.Context, postID, userID string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	likesByPost, ok := r.likes[postID]
	if !ok {
		return false, nil
	}
	_, exists := likesByPost[userID]
	return exists, nil
}

func (r *MemoryRepository) ListLikedPostIDs(_ context.Context, userID string, postIDs []string) (map[string]bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make(map[string]bool)
	for _, postID := range postIDs {
		if likesByPost, ok := r.likes[postID]; ok {
			if _, exists := likesByPost[userID]; exists {
				result[postID] = true
			}
		}
	}
	return result, nil
}

func (r *MemoryRepository) SoftDeletePost(_ context.Context, postID, authorID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	post, ok := r.posts[postID]
	if !ok || post.IsDeleted {
		return ErrPostNotFound
	}
	if post.AuthorID.Hex() != authorID {
		return ErrPostNotFound
	}

	now := time.Now().UTC()
	post.IsDeleted = true
	post.DeletedAt = &now
	post.UpdatedAt = now
	r.adjustTopicsLocked(post.Topics, -1, now)
	return nil
}

func (r *MemoryRepository) ListTrendingTopics(_ context.Context, limit int) ([]*Topic, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if limit <= 0 {
		limit = 10
	}
	topics := make([]*Topic, 0, len(r.topics))
	for _, topic := range r.topics {
		if topic.PostCount <= 0 {
			continue
		}
		cp := *topic
		topics = append(topics, &cp)
	}
	sort.Slice(topics, func(i, j int) bool {
		if topics[i].PostCount == topics[j].PostCount {
			if topics[i].LastPostAt.Equal(topics[j].LastPostAt) {
				return topics[i].Name < topics[j].Name
			}
			return topics[i].LastPostAt.After(topics[j].LastPostAt)
		}
		return topics[i].PostCount > topics[j].PostCount
	})
	if len(topics) > limit {
		topics = topics[:limit]
	}
	return topics, nil
}

func (r *MemoryRepository) CreateLikeIfAbsent(_ context.Context, postID, userID string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	post, ok := r.posts[postID]
	if !ok || post.IsDeleted {
		return false, ErrPostNotFound
	}
	if _, ok := r.likes[postID]; !ok {
		r.likes[postID] = make(map[string]struct{})
	}
	if _, exists := r.likes[postID][userID]; exists {
		return false, nil
	}
	r.likes[postID][userID] = struct{}{}
	return true, nil
}

func (r *MemoryRepository) LikePost(_ context.Context, postID, userID string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	post, ok := r.posts[postID]
	if !ok || post.IsDeleted {
		return false, ErrPostNotFound
	}
	if _, ok := r.likes[postID]; !ok {
		r.likes[postID] = make(map[string]struct{})
	}
	if _, exists := r.likes[postID][userID]; exists {
		return false, nil
	}

	r.likes[postID][userID] = struct{}{}
	post.LikeCount++
	post.UpdatedAt = time.Now().UTC()
	return true, nil
}

func (r *MemoryRepository) RemoveLikeIfPresent(_ context.Context, postID, userID string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	likesByPost, ok := r.likes[postID]
	if !ok {
		return false, nil
	}
	if _, exists := likesByPost[userID]; !exists {
		return false, nil
	}
	delete(likesByPost, userID)
	if len(likesByPost) == 0 {
		delete(r.likes, postID)
	}
	return true, nil
}

func (r *MemoryRepository) UnlikePost(_ context.Context, postID, userID string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	post, ok := r.posts[postID]
	if !ok || post.IsDeleted {
		return false, ErrPostNotFound
	}
	likesByPost, ok := r.likes[postID]
	if !ok {
		return false, nil
	}
	if _, exists := likesByPost[userID]; !exists {
		return false, nil
	}

	delete(likesByPost, userID)
	if len(likesByPost) == 0 {
		delete(r.likes, postID)
	}
	if post.LikeCount > 0 {
		post.LikeCount--
	}
	post.UpdatedAt = time.Now().UTC()
	return true, nil
}

func (r *MemoryRepository) ApplyLikeSideEffects(_ context.Context, postID, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	post, ok := r.posts[postID]
	if !ok || post.IsDeleted {
		return ErrPostNotFound
	}
	if _, ok := r.likes[postID]; !ok {
		return nil
	}
	if _, exists := r.likes[postID][userID]; !exists {
		return nil
	}

	post.LikeCount++
	now := time.Now().UTC()
	post.UpdatedAt = now
	return nil
}

func (r *MemoryRepository) RevertLikeSideEffects(_ context.Context, postID, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	post, ok := r.posts[postID]
	if !ok || post.IsDeleted {
		return ErrPostNotFound
	}
	if _, ok := r.likes[postID]; !ok {
		return nil
	}
	if _, exists := r.likes[postID][userID]; !exists {
		return nil
	}

	if post.LikeCount > 0 {
		post.LikeCount--
	}
	now := time.Now().UTC()
	post.UpdatedAt = now
	return nil
}

func (r *MemoryRepository) adjustTopicsLocked(topics []string, delta int64, lastPostAt time.Time) {
	if len(topics) == 0 {
		return
	}
	now := time.Now().UTC()
	for _, name := range topics {
		if name == "" {
			continue
		}
		topic, ok := r.topics[name]
		if !ok {
			topic = &Topic{
				ID:        bson.NewObjectID(),
				Name:      name,
				CreatedAt: now,
			}
			r.topics[name] = topic
		}
		topic.PostCount += delta
		if topic.PostCount < 0 {
			topic.PostCount = 0
		}
		topic.UpdatedAt = now
		if delta > 0 && lastPostAt.After(topic.LastPostAt) {
			topic.LastPostAt = lastPostAt
		}
	}
}

var _ Repository = (*MongoRepository)(nil)
var _ Repository = (*MemoryRepository)(nil)
