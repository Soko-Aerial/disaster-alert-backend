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

type SOSRepository struct {
	collection *mongo.Collection
}

func NewSOSRepository(db *mongo.Database) *SOSRepository {
	return &SOSRepository{
		collection: db.Collection("sos_requests"),
	}
}

func (r *SOSRepository) Create(sos models.SOSRequest) (*models.SOSRequest, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := r.collection.InsertOne(ctx, sos)
	if err != nil {
		return nil, err
	}

	insertedID, ok := result.InsertedID.(primitive.ObjectID)
	if ok {
		sos.ID = insertedID
	}

	return &sos, nil
}

func (r *SOSRepository) FindAll() ([]models.SOSRequest, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "createdAt", Value: -1}})

	cursor, err := r.collection.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		return nil, err
	}

	defer cursor.Close(ctx)

	requests := make([]models.SOSRequest, 0)

	for cursor.Next(ctx) {
		var sos models.SOSRequest

		if err := cursor.Decode(&sos); err != nil {
			return nil, err
		}

		requests = append(requests, sos)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return requests, nil
}

func (r *SOSRepository) FindByID(sosID primitive.ObjectID) (*models.SOSRequest, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var sos models.SOSRequest

	err := r.collection.FindOne(ctx, bson.M{
		"_id": sosID,
	}).Decode(&sos)

	if err != nil {
		return nil, err
	}

	return &sos, nil
}

func (r *SOSRepository) UpdateStatus(
	sosID primitive.ObjectID,
	status string,
) (*models.SOSRequest, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	now := time.Now()

	updateFields := bson.M{
		"status":    status,
		"updatedAt": now,
	}

	if status == "resolved" || status == "cancelled" {
		updateFields["resolvedAt"] = now
		updateFields["isLiveTracking"] = false
	}

	opts := options.FindOneAndUpdate().
		SetReturnDocument(options.After)

	var updatedSOS models.SOSRequest

	err := r.collection.FindOneAndUpdate(
		ctx,
		bson.M{"_id": sosID},
		bson.M{"$set": updateFields},
		opts,
	).Decode(&updatedSOS)

	if err != nil {
		return nil, err
	}

	return &updatedSOS, nil
}