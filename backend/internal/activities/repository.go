package activities

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"

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

func (r *MongoRepository) activitiesCollection() *mongo.Collection {
	return r.db.Collection("user_activities")
}

func (r *MongoRepository) RecordPostView(ctx context.Context, userID, postID string) error {
	return r.record(ctx, userID, postID, "", ActivityTypeView, true)
}

func (r *MongoRepository) RecordPostLike(ctx context.Context, userID, postID string) error {
	return r.record(ctx, userID, postID, "", ActivityTypeLike, false)
}

func (r *MongoRepository) RecordCommentLike(ctx context.Context, userID, postID, commentID string) error {
	return r.record(ctx, userID, postID, commentID, ActivityTypeLike, false)
}

func (r *MongoRepository) RecordComment(ctx context.Context, userID, postID, commentID string) error {
	return r.record(ctx, userID, postID, commentID, ActivityTypeComment, false)
}

func (r *MongoRepository) record(ctx context.Context, userID, postID, commentID string, activityType ActivityType, increment bool) error {
	if err := r.ensureIndexes(ctx); err != nil {
		return err
	}

	userObjectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil
	}
	postObjectID, err := bson.ObjectIDFromHex(postID)
	if err != nil {
		return nil
	}

	filter := bson.M{
		"user_id": userObjectID,
		"type":    activityType,
		"post_id": postObjectID,
	}
	setOnInsert := bson.M{
		"_id":        bson.NewObjectID(),
		"user_id":    userObjectID,
		"type":       activityType,
		"post_id":    postObjectID,
		"count":      int64(1),
		"created_at": time.Now().UTC(),
	}
	if commentID != "" {
		commentObjectID, err := bson.ObjectIDFromHex(commentID)
		if err != nil {
			return nil
		}
		filter["comment_id"] = commentObjectID
		setOnInsert["comment_id"] = commentObjectID
	}

	update := bson.M{
		"$set":         bson.M{"updated_at": time.Now().UTC()},
		"$setOnInsert": setOnInsert,
	}
	if increment {
		update["$inc"] = bson.M{"count": int64(1)}
		delete(setOnInsert, "count")
	}

	_, err = r.activitiesCollection().UpdateOne(ctx, filter, update, options.UpdateOne().SetUpsert(true))
	return err
}

func (r *MongoRepository) Counts(ctx context.Context, userID string) (Counts, error) {
	if err := r.ensureIndexes(ctx); err != nil {
		return Counts{}, err
	}

	userObjectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return Counts{}, nil
	}

	countFor := func(activityType ActivityType) (int64, error) {
		return r.activitiesCollection().CountDocuments(ctx, bson.M{"user_id": userObjectID, "type": activityType})
	}

	views, err := countFor(ActivityTypeView)
	if err != nil {
		return Counts{}, err
	}
	likes, err := countFor(ActivityTypeLike)
	if err != nil {
		return Counts{}, err
	}
	comments, err := countFor(ActivityTypeComment)
	if err != nil {
		return Counts{}, err
	}

	return Counts{Views: views, Likes: likes, Comments: comments}, nil
}

func (r *MongoRepository) List(ctx context.Context, userID string, activityType ActivityType, limit int) ([]*Activity, error) {
	if err := r.ensureIndexes(ctx); err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 20
	}

	userObjectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return []*Activity{}, nil
	}

	cursor, err := r.activitiesCollection().Find(ctx, bson.M{
		"user_id": userObjectID,
		"type":    activityType,
	}, options.Find().
		SetSort(bson.D{{Key: "updated_at", Value: -1}, {Key: "_id", Value: -1}}).
		SetLimit(int64(limit)))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var activities []*Activity
	if err := cursor.All(ctx, &activities); err != nil {
		return nil, err
	}
	if len(activities) == 0 {
		return []*Activity{}, nil
	}
	return activities, nil
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

	if _, err := r.activitiesCollection().Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "user_id", Value: 1},
			{Key: "type", Value: 1},
			{Key: "updated_at", Value: -1},
		},
	}); err != nil {
		return err
	}

	if _, err := r.activitiesCollection().Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "user_id", Value: 1},
			{Key: "type", Value: 1},
			{Key: "post_id", Value: 1},
			{Key: "comment_id", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	}); err != nil {
		return err
	}

	r.indexesReady = true
	return nil
}

type MemoryRepository struct {
	mu         sync.RWMutex
	activities map[string]*Activity
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{activities: make(map[string]*Activity)}
}

func (r *MemoryRepository) RecordPostView(_ context.Context, userID, postID string) error {
	return r.record(userID, postID, "", ActivityTypeView, true)
}

func (r *MemoryRepository) RecordPostLike(_ context.Context, userID, postID string) error {
	return r.record(userID, postID, "", ActivityTypeLike, false)
}

func (r *MemoryRepository) RecordCommentLike(_ context.Context, userID, postID, commentID string) error {
	return r.record(userID, postID, commentID, ActivityTypeLike, false)
}

func (r *MemoryRepository) RecordComment(_ context.Context, userID, postID, commentID string) error {
	return r.record(userID, postID, commentID, ActivityTypeComment, false)
}

func (r *MemoryRepository) record(userID, postID, commentID string, activityType ActivityType, increment bool) error {
	userObjectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil
	}
	postObjectID, err := bson.ObjectIDFromHex(postID)
	if err != nil {
		return nil
	}

	var commentObjectID *bson.ObjectID
	if commentID != "" {
		parsed, err := bson.ObjectIDFromHex(commentID)
		if err != nil {
			return nil
		}
		commentObjectID = &parsed
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	key := userID + ":" + string(activityType) + ":" + postID + ":" + commentID
	now := time.Now().UTC()
	activity, ok := r.activities[key]
	if !ok {
		activity = &Activity{
			ID:        bson.NewObjectID(),
			UserID:    userObjectID,
			Type:      activityType,
			PostID:    postObjectID,
			CommentID: commentObjectID,
			Count:     1,
			CreatedAt: now,
		}
		r.activities[key] = activity
	}
	if increment && ok {
		activity.Count++
	}
	activity.UpdatedAt = now
	return nil
}

func (r *MemoryRepository) Counts(_ context.Context, userID string) (Counts, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	counts := Counts{}
	for _, activity := range r.activities {
		if activity.UserID.Hex() != userID {
			continue
		}
		switch activity.Type {
		case ActivityTypeView:
			counts.Views++
		case ActivityTypeLike:
			counts.Likes++
		case ActivityTypeComment:
			counts.Comments++
		}
	}
	return counts, nil
}

func (r *MemoryRepository) List(_ context.Context, userID string, activityType ActivityType, limit int) ([]*Activity, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if limit <= 0 {
		limit = 20
	}
	activities := make([]*Activity, 0)
	for _, activity := range r.activities {
		if activity.UserID.Hex() != userID || activity.Type != activityType {
			continue
		}
		cp := *activity
		activities = append(activities, &cp)
	}
	sort.Slice(activities, func(i, j int) bool {
		if activities[i].UpdatedAt.Equal(activities[j].UpdatedAt) {
			return activities[i].ID.Hex() > activities[j].ID.Hex()
		}
		return activities[i].UpdatedAt.After(activities[j].UpdatedAt)
	})
	if len(activities) > limit {
		activities = activities[:limit]
	}
	if len(activities) == 0 {
		return []*Activity{}, nil
	}
	return activities, nil
}

var _ Repository = (*MongoRepository)(nil)
var _ Repository = (*MemoryRepository)(nil)
