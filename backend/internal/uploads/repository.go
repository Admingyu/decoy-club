package uploads

import (
	"context"
	"errors"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Repository interface {
	CreateUploadedFile(ctx context.Context, file *UploadedFile) (*UploadedFile, error)
}

type MongoRepository struct {
	db           *mongo.Database
	indexesMu    sync.Mutex
	indexesReady bool
}

func NewMongoRepository(db *mongo.Database) *MongoRepository {
	return &MongoRepository{db: db}
}

func (r *MongoRepository) uploadsCollection() *mongo.Collection {
	return r.db.Collection("uploaded_files")
}

func (r *MongoRepository) CreateUploadedFile(ctx context.Context, file *UploadedFile) (*UploadedFile, error) {
	if err := r.ensureIndexes(ctx); err != nil {
		return nil, err
	}

	if file.ID.IsZero() {
		file.ID = bson.NewObjectID()
	}
	if file.CreatedAt.IsZero() {
		file.CreatedAt = time.Now().UTC()
	}

	if _, err := r.uploadsCollection().InsertOne(ctx, file); err != nil {
		return nil, err
	}
	created := *file
	return &created, nil
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
	if _, err := r.uploadsCollection().Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "owner_user_id", Value: 1}, {Key: "created_at", Value: -1}},
		Options: options.Index(),
	}); err != nil {
		return err
	}
	r.indexesReady = true
	return nil
}

type MemoryRepository struct {
	mu    sync.Mutex
	files map[string]*UploadedFile
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{files: make(map[string]*UploadedFile)}
}

func (r *MemoryRepository) CreateUploadedFile(_ context.Context, file *UploadedFile) (*UploadedFile, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if file.ID.IsZero() {
		file.ID = bson.NewObjectID()
	}
	if file.CreatedAt.IsZero() {
		file.CreatedAt = time.Now().UTC()
	}
	cp := *file
	r.files[cp.ID.Hex()] = &cp
	created := cp
	return &created, nil
}

var _ Repository = (*MongoRepository)(nil)
var _ Repository = (*MemoryRepository)(nil)
