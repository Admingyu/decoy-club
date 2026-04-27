package notifications

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

var ErrNotificationNotFound = errors.New("notification not found")

type Repository interface {
	CreateNotification(ctx context.Context, notification *Notification) (*Notification, error)
	ListNotifications(ctx context.Context, recipientUserID string, unreadOnly bool, types []NotificationType, page, pageSize int) ([]*Notification, error)
	FindNotificationByID(ctx context.Context, notificationID, recipientUserID string) (*Notification, error)
	CountUnread(ctx context.Context, recipientUserID string) (map[NotificationType]int64, int64, error)
	MarkRead(ctx context.Context, recipientUserID string, notificationIDs []string) error
}

type MongoRepository struct {
	db           *mongo.Database
	indexesMu    sync.Mutex
	indexesReady bool
}

func NewMongoRepository(db *mongo.Database) *MongoRepository {
	return &MongoRepository{db: db}
}

func (r *MongoRepository) notificationsCollection() *mongo.Collection {
	return r.db.Collection("notifications")
}

func (r *MongoRepository) CreateNotification(ctx context.Context, notification *Notification) (*Notification, error) {
	if err := r.ensureIndexes(ctx); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	if notification.ID.IsZero() {
		notification.ID = bson.NewObjectID()
	}
	if notification.CreatedAt.IsZero() {
		notification.CreatedAt = now
	}

	if _, err := r.notificationsCollection().InsertOne(ctx, notification); err != nil {
		return nil, err
	}
	created := *notification
	return &created, nil
}

func (r *MongoRepository) ListNotifications(ctx context.Context, recipientUserID string, unreadOnly bool, types []NotificationType, page, pageSize int) ([]*Notification, error) {
	if err := r.ensureIndexes(ctx); err != nil {
		return nil, err
	}
	recipientObjectID, err := bson.ObjectIDFromHex(recipientUserID)
	if err != nil {
		return []*Notification{}, nil
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	filter := bson.M{"recipient_user_id": recipientObjectID}
	if unreadOnly {
		filter["is_read"] = false
	}
	if len(types) > 0 {
		filter["type"] = bson.M{"$in": types}
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}, {Key: "_id", Value: -1}}).
		SetSkip(int64((page - 1) * pageSize)).
		SetLimit(int64(pageSize))

	cursor, err := r.notificationsCollection().Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var notifications []*Notification
	if err := cursor.All(ctx, &notifications); err != nil {
		return nil, err
	}
	return notifications, nil
}

func (r *MongoRepository) FindNotificationByID(ctx context.Context, notificationID, recipientUserID string) (*Notification, error) {
	if err := r.ensureIndexes(ctx); err != nil {
		return nil, err
	}
	notificationObjectID, err := bson.ObjectIDFromHex(notificationID)
	if err != nil {
		return nil, ErrNotificationNotFound
	}
	recipientObjectID, err := bson.ObjectIDFromHex(recipientUserID)
	if err != nil {
		return nil, ErrNotificationNotFound
	}

	var notification Notification
	if err := r.notificationsCollection().FindOne(ctx, bson.M{
		"_id":               notificationObjectID,
		"recipient_user_id": recipientObjectID,
	}).Decode(&notification); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrNotificationNotFound
		}
		return nil, err
	}
	return &notification, nil
}

func (r *MongoRepository) CountUnread(ctx context.Context, recipientUserID string) (map[NotificationType]int64, int64, error) {
	if err := r.ensureIndexes(ctx); err != nil {
		return nil, 0, err
	}
	recipientObjectID, err := bson.ObjectIDFromHex(recipientUserID)
	if err != nil {
		return map[NotificationType]int64{}, 0, nil
	}

	cursor, err := r.notificationsCollection().Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"recipient_user_id": recipientObjectID, "is_read": false}}},
		{{Key: "$group", Value: bson.M{"_id": "$type", "count": bson.M{"$sum": 1}}}},
	})
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	counts := map[NotificationType]int64{}
	var total int64
	for cursor.Next(ctx) {
		var row struct {
			ID    NotificationType `bson:"_id"`
			Count int64            `bson:"count"`
		}
		if err := cursor.Decode(&row); err != nil {
			return nil, 0, err
		}
		counts[row.ID] = row.Count
		total += row.Count
	}
	if err := cursor.Err(); err != nil {
		return nil, 0, err
	}
	return counts, total, nil
}

func (r *MongoRepository) MarkRead(ctx context.Context, recipientUserID string, notificationIDs []string) error {
	if err := r.ensureIndexes(ctx); err != nil {
		return err
	}
	recipientObjectID, err := bson.ObjectIDFromHex(recipientUserID)
	if err != nil {
		return nil
	}
	if len(notificationIDs) == 0 {
		return nil
	}
	objectIDs := make([]bson.ObjectID, 0, len(notificationIDs))
	for _, id := range notificationIDs {
		objectID, err := bson.ObjectIDFromHex(id)
		if err != nil {
			continue
		}
		objectIDs = append(objectIDs, objectID)
	}
	if len(objectIDs) == 0 {
		return nil
	}

	now := time.Now().UTC()
	_, err = r.notificationsCollection().UpdateMany(ctx, bson.M{
		"_id":               bson.M{"$in": objectIDs},
		"recipient_user_id": recipientObjectID,
		"is_read":           false,
	}, bson.M{
		"$set": bson.M{"is_read": true, "read_at": now},
	})
	return err
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
	if _, err := r.notificationsCollection().Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "recipient_user_id", Value: 1}, {Key: "is_read", Value: 1}, {Key: "created_at", Value: -1}},
	}); err != nil {
		return err
	}
	r.indexesReady = true
	return nil
}

type MemoryRepository struct {
	mu            sync.RWMutex
	notifications map[string]*Notification
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{notifications: make(map[string]*Notification)}
}

func (r *MemoryRepository) CreateNotification(_ context.Context, notification *Notification) (*Notification, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if notification.ID.IsZero() {
		notification.ID = bson.NewObjectID()
	}
	if notification.CreatedAt.IsZero() {
		notification.CreatedAt = time.Now().UTC()
	}
	cp := *notification
	r.notifications[cp.ID.Hex()] = &cp
	created := cp
	return &created, nil
}

func (r *MemoryRepository) ListNotifications(_ context.Context, recipientUserID string, unreadOnly bool, types []NotificationType, page, pageSize int) ([]*Notification, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	result := make([]*Notification, 0)
	typeSet := make(map[NotificationType]struct{}, len(types))
	for _, notificationType := range types {
		typeSet[notificationType] = struct{}{}
	}
	for _, notification := range r.notifications {
		if notification.RecipientUserID.Hex() != recipientUserID {
			continue
		}
		if unreadOnly && notification.IsRead {
			continue
		}
		if len(typeSet) > 0 {
			if _, ok := typeSet[notification.Type]; !ok {
				continue
			}
		}
		cp := *notification
		result = append(result, &cp)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].CreatedAt.Equal(result[j].CreatedAt) {
			return result[i].ID.Hex() > result[j].ID.Hex()
		}
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})
	start := (page - 1) * pageSize
	if start >= len(result) {
		return []*Notification{}, nil
	}
	end := start + pageSize
	if end > len(result) {
		end = len(result)
	}
	return result[start:end], nil
}

func (r *MemoryRepository) FindNotificationByID(_ context.Context, notificationID, recipientUserID string) (*Notification, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	notification, ok := r.notifications[notificationID]
	if !ok || notification.RecipientUserID.Hex() != recipientUserID {
		return nil, ErrNotificationNotFound
	}
	cp := *notification
	return &cp, nil
}

func (r *MemoryRepository) CountUnread(_ context.Context, recipientUserID string) (map[NotificationType]int64, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	counts := map[NotificationType]int64{}
	var total int64
	for _, notification := range r.notifications {
		if notification.RecipientUserID.Hex() != recipientUserID || notification.IsRead {
			continue
		}
		counts[notification.Type]++
		total++
	}
	return counts, total, nil
}

func (r *MemoryRepository) MarkRead(_ context.Context, recipientUserID string, notificationIDs []string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now().UTC()
	for _, id := range notificationIDs {
		notification, ok := r.notifications[id]
		if !ok || notification.RecipientUserID.Hex() != recipientUserID || notification.IsRead {
			continue
		}
		notification.IsRead = true
		notification.ReadAt = &now
	}
	return nil
}

var _ Repository = (*MongoRepository)(nil)
var _ Repository = (*MemoryRepository)(nil)
