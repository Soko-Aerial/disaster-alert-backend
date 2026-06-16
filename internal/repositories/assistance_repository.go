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

type AssistanceRepository struct {
	collection *mongo.Collection
}

func NewAssistanceRepository(db *mongo.Database) *AssistanceRepository {
	return &AssistanceRepository{
		collection: db.Collection("assistance_requests"),
	}
}

func (r *AssistanceRepository) Create(
	request models.AssistanceRequest,
) (*models.AssistanceRequest, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := r.collection.InsertOne(ctx, request)
	if err != nil {
		return nil, err
	}

	insertedID, ok := result.InsertedID.(primitive.ObjectID)
	if ok {
		request.ID = insertedID
	}

	return &request, nil
}

func (r *AssistanceRepository) FindAll() ([]models.AssistanceRequest, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "createdAt", Value: -1}})

	cursor, err := r.collection.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		return nil, err
	}

	defer cursor.Close(ctx)

	requests := make([]models.AssistanceRequest, 0)

	for cursor.Next(ctx) {
		var request models.AssistanceRequest

		if err := cursor.Decode(&request); err != nil {
			return nil, err
		}

		requests = append(requests, request)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return requests, nil
}

func (r *AssistanceRepository) FindByID(
	requestID primitive.ObjectID,
) (*models.AssistanceRequest, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var request models.AssistanceRequest

	err := r.collection.FindOne(ctx, bson.M{
		"_id": requestID,
	}).Decode(&request)

	if err != nil {
		return nil, err
	}

	return &request, nil
}

func (r *AssistanceRepository) UpdateStatus(
	requestID primitive.ObjectID,
	status string,
) (*models.AssistanceRequest, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	now := time.Now().UTC()

	filter := bson.M{
		"_id": requestID,
	}

	update := bson.M{
		"$set": bson.M{
			"status":    status,
			"updatedAt": now,
		},
	}

	opts := options.FindOneAndUpdate().
		SetReturnDocument(options.After)

	var request models.AssistanceRequest

	err := r.collection.FindOneAndUpdate(
		ctx,
		filter,
		update,
		opts,
	).Decode(&request)

	if err != nil {
		return nil, err
	}

	return &request, nil
}