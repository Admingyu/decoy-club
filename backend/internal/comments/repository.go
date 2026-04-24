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
	SoftDeleteComment(ctx context.Context, commentID, authorID string) error
}

type MongoRepository struct {
	db           *mongo.Database
	postCounter  PostCounter
	indexesMu    sync.Mutex
	indexesReady bool
}

func NewMongoRepository(db *mongo.Database, postCounter PostCounter) *MongoRepository {
	return &MongoRepository{db: db, postCounter: postCounter}
}

func (r *MongoRepository) commentsCollection() *mongo.Collection {
	return r.db.Collection("comments")
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

	r.indexesReady = true
	return nil
}

type MemoryRepository struct {
	mu          sync.RWMutex
	comments    map[string]*Comment
	postCounter PostCounter
}

func NewMemoryRepository(postCounter PostCounter) *MemoryRepository {
	return &MemoryRepository{
		comments:    make(map[string]*Comment),
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
