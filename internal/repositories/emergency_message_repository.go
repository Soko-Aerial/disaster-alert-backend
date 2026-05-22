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

type EmergencyMessageRepository struct {
	collection *mongo.Collection
}

func NewEmergencyMessageRepository(db *mongo.Database) *EmergencyMessageRepository {
	return &EmergencyMessageRepository{
		collection: db.Collection("emergency_messages"),
	}
}

func (r *EmergencyMessageRepository) Create(
	message models.EmergencyMessage,
) (*models.EmergencyMessage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	message.ID = primitive.NewObjectID()
	message.CreatedAt = time.Now().UTC()
	message.UpdatedAt = time.Now().UTC()

	_, err := r.collection.InsertOne(ctx, message)
	if err != nil {
		return nil, err
	}

	return &message, nil
}

func (r *EmergencyMessageRepository) FindByUserID(
	userID primitive.ObjectID,
) ([]models.EmergencyMessage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"userId": userID,
	}

	opts := options.Find().
		SetSort(bson.D{
			{Key: "createdAt", Value: -1},
		})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	messages := make([]models.EmergencyMessage, 0)
	if err := cursor.All(ctx, &messages); err != nil {
		return nil, err
	}

	return messages, nil
}

func (r *EmergencyMessageRepository) FindByID(
	id primitive.ObjectID,
	userID primitive.ObjectID,
) (*models.EmergencyMessage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"_id":    id,
		"userId": userID,
	}

	var message models.EmergencyMessage
	err := r.collection.FindOne(ctx, filter).Decode(&message)
	if err != nil {
		return nil, err
	}

	return &message, nil
}

func (r *EmergencyMessageRepository) Delete(
	id primitive.ObjectID,
	userID primitive.ObjectID,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"_id":    id,
		"userId": userID,
	}

	_, err := r.collection.DeleteOne(ctx, filter)
	return err
}