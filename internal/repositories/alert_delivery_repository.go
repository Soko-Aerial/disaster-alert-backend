package repositories

import (
	"context"
	"errors"
	"time"

	"disaster_alert_backend/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AlertDeliveryRepository struct {
	batchCollection     *mongo.Collection
	recipientCollection *mongo.Collection
}

func NewAlertDeliveryRepository(db *mongo.Database) *AlertDeliveryRepository {
	return &AlertDeliveryRepository{
		batchCollection:     db.Collection("alert_delivery_batches"),
		recipientCollection: db.Collection("alert_recipient_deliveries"),
	}
}

func (r *AlertDeliveryRepository) EnsureIndexes() error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	_, err := r.batchCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "alertId", Value: 1},
				{Key: "createdAt", Value: -1},
			},
		},
		{
			Keys: bson.D{
				{Key: "deliveryType", Value: 1},
				{Key: "targetingMode", Value: 1},
			},
		},
	})
	if err != nil {
		return err
	}

	_, err = r.recipientCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "alertId", Value: 1},
				{Key: "userId", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "deliveryBatchId", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "dedupKey", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{
				{Key: "status", Value: 1},
				{Key: "createdAt", Value: -1},
			},
		},
	})

	return err
}

func (r *AlertDeliveryRepository) CreateBatch(
	batch models.AlertDeliveryBatch,
) (*models.AlertDeliveryBatch, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	now := time.Now().UTC()

	if batch.ID.IsZero() {
		batch.ID = primitive.NewObjectID()
	}

	if batch.StartedAt.IsZero() {
		batch.StartedAt = now
	}

	if batch.CreatedAt.IsZero() {
		batch.CreatedAt = now
	}

	batch.UpdatedAt = now

	_, err := r.batchCollection.InsertOne(ctx, batch)
	if err != nil {
		return nil, err
	}

	return &batch, nil
}

func (r *AlertDeliveryRepository) CompleteBatch(
	batchID primitive.ObjectID,
	successCount int,
	failureCount int,
	skippedCount int,
) error {
	if batchID.IsZero() {
		return errors.New("batch id is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	now := time.Now().UTC()

	_, err := r.batchCollection.UpdateOne(
		ctx,
		bson.M{"_id": batchID},
		bson.M{
			"$set": bson.M{
				"successCount": successCount,
				"failureCount": failureCount,
				"skippedCount": skippedCount,
				"completedAt":  now,
				"updatedAt":    now,
			},
		},
	)

	return err
}

func (r *AlertDeliveryRepository) HasDedupKey(
	dedupKey string,
) (bool, error) {
	if dedupKey == "" {
		return false, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	count, err := r.recipientCollection.CountDocuments(
		ctx,
		bson.M{"dedupKey": dedupKey},
	)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *AlertDeliveryRepository) CreateRecipientDelivery(
	delivery models.AlertRecipientDelivery,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	now := time.Now().UTC()

	if delivery.ID.IsZero() {
		delivery.ID = primitive.NewObjectID()
	}

	if delivery.CreatedAt.IsZero() {
		delivery.CreatedAt = now
	}

	delivery.UpdatedAt = now

	_, err := r.recipientCollection.InsertOne(ctx, delivery)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil
		}

		return err
	}

	return nil
}

func (r *AlertDeliveryRepository) FindBatchesByAlertID(
	alertID primitive.ObjectID,
) ([]models.AlertDeliveryBatch, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	findOptions := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}})

	cursor, err := r.batchCollection.Find(
		ctx,
		bson.M{"alertId": alertID},
		findOptions,
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	batches := []models.AlertDeliveryBatch{}

	for cursor.Next(ctx) {
		var batch models.AlertDeliveryBatch
		if err := cursor.Decode(&batch); err != nil {
			return nil, err
		}

		batches = append(batches, batch)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return batches, nil
}

func (r *AlertDeliveryRepository) FindRecipientsByAlertID(
	alertID primitive.ObjectID,
	limit int64,
) ([]models.AlertRecipientDelivery, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if limit <= 0 {
		limit = 500
	}

	findOptions := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}}).
		SetLimit(limit)

	cursor, err := r.recipientCollection.Find(
		ctx,
		bson.M{"alertId": alertID},
		findOptions,
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	recipients := []models.AlertRecipientDelivery{}

	for cursor.Next(ctx) {
		var recipient models.AlertRecipientDelivery
		if err := cursor.Decode(&recipient); err != nil {
			return nil, err
		}

		recipients = append(recipients, recipient)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return recipients, nil
}

func (r *AlertDeliveryRepository) HasUserReceivedAlert(
	alertID primitive.ObjectID,
	userID primitive.ObjectID,
) (bool, error) {
	if alertID.IsZero() || userID.IsZero() {
		return false, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	count, err := r.recipientCollection.CountDocuments(
		ctx,
		bson.M{
			"alertId": alertID,
			"userId":  userID,
			"status": bson.M{
				"$in": []string{
					models.AlertDeliveryStatusSent,
					models.AlertDeliveryStatusPending,
				},
			},
		},
	)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
