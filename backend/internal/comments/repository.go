package comments

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

var ErrCommentNotFound = errors.New("comment not found")
var ErrInvalidCommentAuthorID = errors.New("invalid comment author id")
var ErrInvalidPostID = errors.New("invalid post id")
var ErrReplyDepthExceeded = errors.New("only one level of replies is supported")

type PostCounter interface {
	AdjustCommentCount(ctx context.Context, postID string, delta int64) error
}

type Repository interface {
	CreateComment(ctx context.Context, comment *Comment) (*Comment, error)
	FindCommentByID(ctx context.Context, id string) (*Comment, error)
	ListCommentsByPostID(ctx context.Context, postID string) ([]*Comment, error)
	HasLike(ctx context.Context, commentID, userID string) (bool, error)
	ListLikedCommentIDs(ctx context.Context, userID string, commentIDs []string) (map[string]bool, error)
	LikeComment(ctx context.Context, commentID, userID string) (bool, error)
	UnlikeComment(ctx context.Context, commentID, userID string) (bool, error)
	SoftDeleteComment(ctx context.Context, commentID, authorID string) error
}

type MongoRepository struct {
	db           *mongo.Database
	postCounter  PostCounter
	indexesMu    sync.Mutex
	indexesReady bool
}

type commentLikeEdge struct {
	ID        bson.ObjectID `bson:"_id,omitempty"`
	CommentID bson.ObjectID `bson:"comment_id"`
	UserID    bson.ObjectID `bson:"user_id"`
	CreatedAt time.Time     `bson:"created_at"`
}

func NewMongoRepository(db *mongo.Database, postCounter PostCounter) *MongoRepository {
	return &MongoRepository{db: db, postCounter: postCounter}
}

func (r *MongoRepository) commentsCollection() *mongo.Collection {
	return r.db.Collection("comments")
}

func (r *MongoRepository) commentLikesCollection() *mongo.Collection {
	return r.db.Collection("comment_likes")
}

func (r *MongoRepository) CreateComment(ctx context.Context, comment *Comment) (*Comment, error) {
	if err := r.ensureIndexes(ctx); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	if comment.ID.IsZero() {
		comment.ID = bson.NewObjectID()
	}
	if comment.CreatedAt.IsZero() {
		comment.CreatedAt = now
	}
	if comment.UpdatedAt.IsZero() {
		comment.UpdatedAt = now
	}

	if comment.ParentCommentID != nil {
		parent, err := r.findCommentByObjectID(ctx, *comment.ParentCommentID)
		if err != nil {
			return nil, err
		}
		if parent.ParentCommentID != nil {
			return nil, ErrReplyDepthExceeded
		}
	}

	if _, err := r.commentsCollection().InsertOne(ctx, comment); err != nil {
		return nil, err
	}
	if r.postCounter != nil {
		if err := r.postCounter.AdjustCommentCount(ctx, comment.PostID.Hex(), 1); err != nil {
			_, _ = r.commentsCollection().DeleteOne(ctx, bson.M{"_id": comment.ID})
			return nil, err
		}
	}

	created := *comment
	return &created, nil
}

func (r *MongoRepository) FindCommentByID(ctx context.Context, id string) (*Comment, error) {
	if err := r.ensureIndexes(ctx); err != nil {
		return nil, err
	}

	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrCommentNotFound
	}

	return r.findCommentByObjectID(ctx, objectID)
}

func (r *MongoRepository) findCommentByObjectID(ctx context.Context, id bson.ObjectID) (*Comment, error) {
	var comment Comment
	if err := r.commentsCollection().FindOne(ctx, bson.M{"_id": id}).Decode(&comment); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrCommentNotFound
		}
		return nil, err
	}
	return &comment, nil
}

func (r *MongoRepository) ListCommentsByPostID(ctx context.Context, postID string) ([]*Comment, error) {
	if err := r.ensureIndexes(ctx); err != nil {
		return nil, err
	}

	postObjectID, err := bson.ObjectIDFromHex(postID)
	if err != nil {
		return nil, ErrInvalidPostID
	}

	opts := options.Find().SetSort(bson.D{
		{Key: "created_at", Value: 1},
		{Key: "_id", Value: 1},
	})
	cursor, err := r.commentsCollection().Find(ctx, bson.M{"post_id": postObjectID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var comments []*Comment
	if err := cursor.All(ctx, &comments); err != nil {
		return nil, err
	}
	return comments, nil
}

func (r *MongoRepository) HasLike(ctx context.Context, commentID, userID string) (bool, error) {
	if err := r.ensureIndexes(ctx); err != nil {
		return false, err
	}
	commentObjectID, err := bson.ObjectIDFromHex(commentID)
	if err != nil {
		return false, nil
	}
	userObjectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return false, nil
	}

	err = r.commentLikesCollection().FindOne(ctx, bson.M{"comment_id": commentObjectID, "user_id": userObjectID}).Err()
	if err == nil {
		return true, nil
	}
	if errors.Is(err, mongo.ErrNoDocuments) {
		return false, nil
	}
	return false, err
}

func (r *MongoRepository) ListLikedCommentIDs(ctx context.Context, userID string, commentIDs []string) (map[string]bool, error) {
	if err := r.ensureIndexes(ctx); err != nil {
		return nil, err
	}
	if len(commentIDs) == 0 {
		return map[string]bool{}, nil
	}
	userObjectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return map[string]bool{}, nil
	}
	commentObjectIDs := make([]bson.ObjectID, 0, len(commentIDs))
	for _, commentID := range commentIDs {
		objectID, err := bson.ObjectIDFromHex(commentID)
		if err != nil {
			continue
		}
		commentObjectIDs = append(commentObjectIDs, objectID)
	}
	if len(commentObjectIDs) == 0 {
		return map[string]bool{}, nil
	}

	cursor, err := r.commentLikesCollection().Find(ctx, bson.M{
		"user_id":    userObjectID,
		"comment_id": bson.M{"$in": commentObjectIDs},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	result := make(map[string]bool)
	for cursor.Next(ctx) {
		var edge commentLikeEdge
		if err := cursor.Decode(&edge); err != nil {
			return nil, err
		}
		result[edge.CommentID.Hex()] = true
	}
	return result, cursor.Err()
}

func (r *MongoRepository) LikeComment(ctx context.Context, commentID, userID string) (bool, error) {
	if err := r.ensureIndexes(ctx); err != nil {
		return false, err
	}
	commentObjectID, err := bson.ObjectIDFromHex(commentID)
	if err != nil {
		return false, ErrCommentNotFound
	}
	userObjectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return false, ErrCommentNotFound
	}

	if err := r.commentsCollection().FindOne(ctx, bson.M{"_id": commentObjectID, "is_deleted": false}).Err(); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return false, ErrCommentNotFound
		}
		return false, err
	}

	_, err = r.commentLikesCollection().InsertOne(ctx, commentLikeEdge{
		ID:        bson.NewObjectID(),
		CommentID: commentObjectID,
		UserID:    userObjectID,
		CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return false, nil
		}
		return false, err
	}

	now := time.Now().UTC()
	result, err := r.commentsCollection().UpdateOne(ctx, bson.M{
		"_id":        commentObjectID,
		"is_deleted": false,
	}, bson.M{
		"$inc": bson.M{"like_count": 1},
		"$set": bson.M{"updated_at": now},
	})
	if err != nil || result.MatchedCount == 0 {
		_, _ = r.commentLikesCollection().DeleteOne(ctx, bson.M{"comment_id": commentObjectID, "user_id": userObjectID})
		if err != nil {
			return false, err
		}
		return false, ErrCommentNotFound
	}

	return true, nil
}

func (r *MongoRepository) UnlikeComment(ctx context.Context, commentID, userID string) (bool, error) {
	if err := r.ensureIndexes(ctx); err != nil {
		return false, err
	}
	commentObjectID, err := bson.ObjectIDFromHex(commentID)
	if err != nil {
		return false, ErrCommentNotFound
	}
	userObjectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return false, ErrCommentNotFound
	}

	filter := bson.M{"comment_id": commentObjectID, "user_id": userObjectID}
	var removed commentLikeEdge
	if err := r.commentLikesCollection().FindOne(ctx, filter).Decode(&removed); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return false, nil
		}
		return false, err
	}

	result, err := r.commentLikesCollection().DeleteOne(ctx, filter)
	if err != nil {
		return false, err
	}
	if result.DeletedCount == 0 {
		return false, nil
	}

	now := time.Now().UTC()
	updateResult, err := r.commentsCollection().UpdateOne(ctx, bson.M{
		"_id":        commentObjectID,
		"is_deleted": false,
	}, bson.M{
		"$inc": bson.M{"like_count": -1},
		"$set": bson.M{"updated_at": now},
	})
	if err != nil || updateResult.MatchedCount == 0 {
		_, _ = r.commentLikesCollection().InsertOne(ctx, removed)
		if err != nil {
			return false, err
		}
		return false, ErrCommentNotFound
	}

	return true, nil
}

func (r *MongoRepository) SoftDeleteComment(ctx context.Context, commentID, authorID string) error {
	if err := r.ensureIndexes(ctx); err != nil {
		return err
	}

	commentObjectID, err := bson.ObjectIDFromHex(commentID)
	if err != nil {
		return ErrCommentNotFound
	}
	authorObjectID, err := bson.ObjectIDFromHex(authorID)
	if err != nil {
		return ErrCommentNotFound
	}

	comment, err := r.findCommentByObjectID(ctx, commentObjectID)
	if err != nil {
		return err
	}
	if comment.AuthorID != authorObjectID || comment.IsDeleted {
		return ErrCommentNotFound
	}

	now := time.Now().UTC()
	result, err := r.commentsCollection().UpdateOne(ctx, bson.M{
		"_id":        commentObjectID,
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
		return ErrCommentNotFound
	}
	if r.postCounter != nil {
		if err := r.postCounter.AdjustCommentCount(ctx, comment.PostID.Hex(), -1); err != nil {
			_, _ = r.commentsCollection().UpdateOne(ctx, bson.M{"_id": commentObjectID}, bson.M{
				"$set": bson.M{
					"is_deleted": false,
					"updated_at": comment.UpdatedAt,
				},
				"$unset": bson.M{"deleted_at": ""},
			})
			return err
		}
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

	if _, err := r.commentsCollection().Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "post_id", Value: 1}, {Key: "created_at", Value: 1}},
	}); err != nil {
		return err
	}
	if _, err := r.commentsCollection().Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "parent_comment_id", Value: 1}, {Key: "created_at", Value: 1}},
	}); err != nil {
		return err
	}
	if _, err := r.commentLikesCollection().Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "comment_id", Value: 1},
			{Key: "user_id", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	}); err != nil {
		return err
	}

	r.indexesReady = true
	return nil
}

type MemoryRepository struct {
	mu          sync.RWMutex
	comments    map[string]*Comment
	likes       map[string]map[string]struct{}
	postCounter PostCounter
}

func NewMemoryRepository(postCounter PostCounter) *MemoryRepository {
	return &MemoryRepository{
		comments:    make(map[string]*Comment),
		likes:       make(map[string]map[string]struct{}),
		postCounter: postCounter,
	}
}

func (r *MemoryRepository) CreateComment(ctx context.Context, comment *Comment) (*Comment, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if comment.ParentCommentID != nil {
		parent, ok := r.comments[comment.ParentCommentID.Hex()]
		if !ok {
			return nil, ErrCommentNotFound
		}
		if parent.ParentCommentID != nil {
			return nil, ErrReplyDepthExceeded
		}
	}

	now := time.Now().UTC()
	if comment.ID.IsZero() {
		comment.ID = bson.NewObjectID()
	}
	if comment.CreatedAt.IsZero() {
		comment.CreatedAt = now
	}
	if comment.UpdatedAt.IsZero() {
		comment.UpdatedAt = now
	}

	cp := *comment
	r.comments[cp.ID.Hex()] = &cp
	if r.postCounter != nil {
		if err := r.postCounter.AdjustCommentCount(ctx, cp.PostID.Hex(), 1); err != nil {
			delete(r.comments, cp.ID.Hex())
			return nil, err
		}
	}

	created := cp
	return &created, nil
}

func (r *MemoryRepository) FindCommentByID(_ context.Context, id string) (*Comment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	comment, ok := r.comments[id]
	if !ok {
		return nil, ErrCommentNotFound
	}
	cp := *comment
	return &cp, nil
}

func (r *MemoryRepository) ListCommentsByPostID(_ context.Context, postID string) ([]*Comment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if _, err := bson.ObjectIDFromHex(postID); err != nil {
		return nil, ErrInvalidPostID
	}

	result := make([]*Comment, 0)
	for _, comment := range r.comments {
		if comment.PostID.Hex() != postID {
			continue
		}
		cp := *comment
		result = append(result, &cp)
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].CreatedAt.Equal(result[j].CreatedAt) {
			return result[i].ID.Hex() < result[j].ID.Hex()
		}
		return result[i].CreatedAt.Before(result[j].CreatedAt)
	})

	return result, nil
}

func (r *MemoryRepository) HasLike(_ context.Context, commentID, userID string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	likesByComment, ok := r.likes[commentID]
	if !ok {
		return false, nil
	}
	_, exists := likesByComment[userID]
	return exists, nil
}

func (r *MemoryRepository) ListLikedCommentIDs(_ context.Context, userID string, commentIDs []string) (map[string]bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make(map[string]bool)
	for _, commentID := range commentIDs {
		if likesByComment, ok := r.likes[commentID]; ok {
			if _, exists := likesByComment[userID]; exists {
				result[commentID] = true
			}
		}
	}
	return result, nil
}

func (r *MemoryRepository) LikeComment(_ context.Context, commentID, userID string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	comment, ok := r.comments[commentID]
	if !ok || comment.IsDeleted {
		return false, ErrCommentNotFound
	}
	if _, ok := r.likes[commentID]; !ok {
		r.likes[commentID] = make(map[string]struct{})
	}
	if _, exists := r.likes[commentID][userID]; exists {
		return false, nil
	}

	r.likes[commentID][userID] = struct{}{}
	comment.LikeCount++
	comment.UpdatedAt = time.Now().UTC()
	return true, nil
}

func (r *MemoryRepository) UnlikeComment(_ context.Context, commentID, userID string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	comment, ok := r.comments[commentID]
	if !ok || comment.IsDeleted {
		return false, ErrCommentNotFound
	}
	likesByComment, ok := r.likes[commentID]
	if !ok {
		return false, nil
	}
	if _, exists := likesByComment[userID]; !exists {
		return false, nil
	}

	delete(likesByComment, userID)
	if len(likesByComment) == 0 {
		delete(r.likes, commentID)
	}
	if comment.LikeCount > 0 {
		comment.LikeCount--
	}
	comment.UpdatedAt = time.Now().UTC()
	return true, nil
}

func (r *MemoryRepository) SoftDeleteComment(ctx context.Context, commentID, authorID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	comment, ok := r.comments[commentID]
	if !ok || comment.AuthorID.Hex() != authorID || comment.IsDeleted {
		return ErrCommentNotFound
	}

	now := time.Now().UTC()
	comment.IsDeleted = true
	comment.DeletedAt = &now
	comment.UpdatedAt = now
	if r.postCounter != nil {
		return r.postCounter.AdjustCommentCount(ctx, comment.PostID.Hex(), -1)
	}
	return nil
}

var _ Repository = (*MongoRepository)(nil)
var _ Repository = (*MemoryRepository)(nil)
