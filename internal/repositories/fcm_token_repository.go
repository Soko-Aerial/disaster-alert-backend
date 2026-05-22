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

type FCMTokenRepository struct {
	collection *mongo.Collection
}

func NewFCMTokenRepository(db *mongo.Database) *FCMTokenRepository {
	return &FCMTokenRepository{
		collection: db.Collection("fcm_tokens"),
	}
}

func (r *FCMTokenRepository) SaveOrUpdateToken(token models.FCMToken) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"userId": token.UserID,
		"token":  token.Token,
	}

	update := bson.M{
		"$set": bson.M{
			"deviceType": token.DeviceType,
			"isActive":   true,
			"updatedAt":  token.UpdatedAt,
		},
		"$setOnInsert": bson.M{
			"userId":    token.UserID,
			"token":     token.Token,
			"createdAt": token.CreatedAt,
		},
	}

	opts := options.Update().SetUpsert(true)

	_, err := r.collection.UpdateOne(ctx, filter, update, opts)

	return err
}

func (r *FCMTokenRepository) FindActiveTokensByUserID(userID primitive.ObjectID) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.M{
		"userId":   userID,
		"isActive": true,
	})

	if err != nil {
		return nil, err
	}

	defer cursor.Close(ctx)

	var tokens []string

	for cursor.Next(ctx) {
		var fcmToken models.FCMToken

		if err := cursor.Decode(&fcmToken); err != nil {
			return nil, err
		}

		tokens = append(tokens, fcmToken.Token)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return tokens, nil
}

func (r *FCMTokenRepository) FindAllActiveTokens() ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.M{
		"isActive": true,
	})

	if err != nil {
		return nil, err
	}

	defer cursor.Close(ctx)

	var tokens []string

	for cursor.Next(ctx) {
		var fcmToken models.FCMToken

		if err := cursor.Decode(&fcmToken); err != nil {
			return nil, err
		}

		tokens = append(tokens, fcmToken.Token)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return tokens, nil
}