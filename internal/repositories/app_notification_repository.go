package repositories

import (
	"context"
	"time"

	"disaster_alert_backend/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const readNotificationMainInboxHours = 48

type AppNotificationRepository struct {
	collection *mongo.Collection
}

func NewAppNotificationRepository(db *mongo.Database) *AppNotificationRepository {
	return &AppNotificationRepository{
		collection: db.Collection("app_notifications"),
	}
}

func (r *AppNotificationRepository) Create(
	notification models.AppNotification,
) (*models.AppNotification, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	now := time.Now().UTC()

	if notification.ID.IsZero() {
		notification.ID = primitive.NewObjectID()
	}

	notification.CreatedAt = now
	notification.UpdatedAt = now

	if notification.RecipientRole == "" {
		notification.RecipientRole = "user"
	}

	notification.IsArchived = false

	_, err := r.collection.InsertOne(ctx, notification)
	if err != nil {
		return nil, err
	}

	return &notification, nil
}

func (r *AppNotificationRepository) CreateMany(
	notifications []models.AppNotification,
) error {
	if len(notifications) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	now := time.Now().UTC()

	documents := make([]interface{}, 0, len(notifications))

	for index := range notifications {
		if notifications[index].ID.IsZero() {
			notifications[index].ID = primitive.NewObjectID()
		}

		if notifications[index].RecipientRole == "" {
			notifications[index].RecipientRole = "user"
		}

		notifications[index].IsArchived = false
		notifications[index].CreatedAt = now
		notifications[index].UpdatedAt = now

		documents = append(documents, notifications[index])
	}

	_, err := r.collection.InsertMany(ctx, documents)

	return err
}

// FindByRecipientID returns the main notification inbox.
//
// Main inbox behavior:
// - unread notifications stay visible
// - recently read notifications stay visible for a short time
// - old read notifications are hidden from the main inbox
func (r *AppNotificationRepository) FindByRecipientID(
	recipientID primitive.ObjectID,
	limit int,
) ([]models.AppNotification, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if limit <= 0 {
		limit = 50
	}

	recentReadCutoff := time.Now().UTC().Add(
		-readNotificationMainInboxHours * time.Hour,
	)

	filter := bson.M{
		"recipientId": recipientID,
		"isArchived": bson.M{
			"$ne": true,
		},
		"$or": []bson.M{
			{
				"isRead": false,
			},
			{
				"isRead": true,
				"readAt": bson.M{
					"$gte": recentReadCutoff,
				},
			},
			{
				"isRead": true,
				"readAt": bson.M{
					"$exists": false,
				},
				"updatedAt": bson.M{
					"$gte": recentReadCutoff,
				},
			},
		},
	}

	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "createdAt", Value: -1}})
	findOptions.SetLimit(int64(limit))

	cursor, err := r.collection.Find(
		ctx,
		filter,
		findOptions,
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	notifications := make([]models.AppNotification, 0)

	for cursor.Next(ctx) {
		var notification models.AppNotification

		if err := cursor.Decode(&notification); err != nil {
			return nil, err
		}

		notifications = append(notifications, notification)
	}

	return notifications, cursor.Err()
}

// FindHistoryByRecipientID returns the longer notification history.
// Use this for a separate "History" tab/page.
func (r *AppNotificationRepository) FindHistoryByRecipientID(
	recipientID primitive.ObjectID,
	limit int,
) ([]models.AppNotification, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if limit <= 0 {
		limit = 100
	}

	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "createdAt", Value: -1}})
	findOptions.SetLimit(int64(limit))

	cursor, err := r.collection.Find(
		ctx,
		bson.M{
			"recipientId": recipientID,
		},
		findOptions,
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	notifications := make([]models.AppNotification, 0)

	for cursor.Next(ctx) {
		var notification models.AppNotification

		if err := cursor.Decode(&notification); err != nil {
			return nil, err
		}

		notifications = append(notifications, notification)
	}

	return notifications, cursor.Err()
}

func (r *AppNotificationRepository) CountUnreadByRecipientID(
	recipientID primitive.ObjectID,
) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return r.collection.CountDocuments(
		ctx,
		bson.M{
			"recipientId": recipientID,
			"isRead":      false,
			"isArchived": bson.M{
				"$ne": true,
			},
		},
	)
}

func (r *AppNotificationRepository) MarkRead(
	recipientID primitive.ObjectID,
	notificationIDs []primitive.ObjectID,
) error {
	if len(notificationIDs) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	now := time.Now().UTC()

	_, err := r.collection.UpdateMany(
		ctx,
		bson.M{
			"recipientId": recipientID,
			"_id": bson.M{
				"$in": notificationIDs,
			},
		},
		bson.M{
			"$set": bson.M{
				"isRead":    true,
				"readAt":    now,
				"updatedAt": now,
			},
		},
	)

	return err
}

func (r *AppNotificationRepository) MarkAllRead(
	recipientID primitive.ObjectID,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	now := time.Now().UTC()

	_, err := r.collection.UpdateMany(
		ctx,
		bson.M{
			"recipientId": recipientID,
			"isRead":      false,
		},
		bson.M{
			"$set": bson.M{
				"isRead":    true,
				"readAt":    now,
				"updatedAt": now,
			},
		},
	)

	return err
}

// ArchiveReadNotificationsOlderThan hides old read notifications from the main inbox.
// It does not delete them.
func (r *AppNotificationRepository) ArchiveReadNotificationsOlderThan(
	recipientID primitive.ObjectID,
	hours int,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if hours <= 0 {
		hours = readNotificationMainInboxHours
	}

	now := time.Now().UTC()
	cutoff := now.Add(-time.Duration(hours) * time.Hour)

	_, err := r.collection.UpdateMany(
		ctx,
		bson.M{
			"recipientId": recipientID,
			"isRead":      true,
			"isArchived": bson.M{
				"$ne": true,
			},
			"$or": []bson.M{
				{
					"readAt": bson.M{
						"$lt": cutoff,
					},
				},
				{
					"readAt": bson.M{
						"$exists": false,
					},
					"updatedAt": bson.M{
						"$lt": cutoff,
					},
				},
			},
		},
		bson.M{
			"$set": bson.M{
				"isArchived": true,
				"archivedAt": now,
				"updatedAt":  now,
			},
		},
	)

	return err
}

func (r *AppNotificationRepository) CleanupReadNotificationsOlderThan(
	days int,
) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	if days <= 0 {
		days = 30
	}

	cutoff := time.Now().UTC().Add(-time.Duration(days) * 24 * time.Hour)

	result, err := r.collection.DeleteMany(
		ctx,
		bson.M{
			"isRead": true,
			"$or": []bson.M{
				{
					"readAt": bson.M{
						"$lt": cutoff,
					},
				},
				{
					"readAt": bson.M{
						"$exists": false,
					},
					"updatedAt": bson.M{
						"$lt": cutoff,
					},
				},
			},
		},
	)
	if err != nil {
		return 0, err
	}

	return result.DeletedCount, nil
}

func (r *AppNotificationRepository) EnsureIndexes() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "recipientId", Value: 1},
				{Key: "createdAt", Value: -1},
			},
		},
		{
			Keys: bson.D{
				{Key: "recipientId", Value: 1},
				{Key: "isRead", Value: 1},
				{Key: "isArchived", Value: 1},
				{Key: "readAt", Value: -1},
			},
		},
		{
			Keys: bson.D{
				{Key: "type", Value: 1},
				{Key: "referenceId", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "expiresAt", Value: 1},
			},
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexes)

	return err
}
