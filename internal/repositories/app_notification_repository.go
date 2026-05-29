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

		notifications[index].CreatedAt = now
		notifications[index].UpdatedAt = now

		documents = append(documents, notifications[index])
	}

	_, err := r.collection.InsertMany(ctx, documents)
	return err
}

func (r *AppNotificationRepository) FindByRecipientID(
	recipientID primitive.ObjectID,
	limit int,
) ([]models.AppNotification, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if limit <= 0 {
		limit = 50
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
			"isRead":     false,
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
				"updatedAt": time.Now().UTC(),
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

	_, err := r.collection.UpdateMany(
		ctx,
		bson.M{
			"recipientId": recipientID,
			"isRead":     false,
		},
		bson.M{
			"$set": bson.M{
				"isRead":    true,
				"updatedAt": time.Now().UTC(),
			},
		},
	)

	return err
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
			},
		},
		{
			Keys: bson.D{
				{Key: "type", Value: 1},
				{Key: "referenceId", Value: 1},
			},
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	return err
}