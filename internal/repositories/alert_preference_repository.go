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

type AlertPreferenceRepository struct {
	collection *mongo.Collection
}

func NewAlertPreferenceRepository(db *mongo.Database) *AlertPreferenceRepository {
	return &AlertPreferenceRepository{
		collection: db.Collection("alert_preferences"),
	}
}

func (r *AlertPreferenceRepository) FindByUserID(
	userID primitive.ObjectID,
) (*models.AlertPreference, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"userId": userID,
	}

	var preference models.AlertPreference
	err := r.collection.FindOne(ctx, filter).Decode(&preference)
	if err != nil {
		return nil, err
	}

	return &preference, nil
}

func (r *AlertPreferenceRepository) UpsertByUserID(
	userID primitive.ObjectID,
	preference models.AlertPreference,
) (*models.AlertPreference, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	now := time.Now().UTC()

	filter := bson.M{
		"userId": userID,
	}

	update := bson.M{
		"$set": bson.M{
			"userId":             userID,
			"fire":               preference.Fire,
			"flood":              preference.Flood,
			"weather":            preference.Weather,
			"earthquake":         preference.Earthquake,
			"health":             preference.Health,
			"conflict":           preference.Conflict,
			"drought":            preference.Drought,
			"protests":           preference.Protests,
			"robbery":            preference.Robbery,
			"munitions":          preference.Munitions,
			"galamsey":           preference.Galamsey,
			"unverifiedActivity": preference.UnverifiedActivity,
			"criticalAlerts":     preference.CriticalAlerts,
			"updatedAt":          now,
		},
		"$setOnInsert": bson.M{
			"_id":       primitive.NewObjectID(),
			"createdAt": now,
		},
	}

	opts := options.FindOneAndUpdate().
		SetUpsert(true).
		SetReturnDocument(options.After)

	var updatedPreference models.AlertPreference
	err := r.collection.FindOneAndUpdate(
		ctx,
		filter,
		update,
		opts,
	).Decode(&updatedPreference)

	if err != nil {
		return nil, err
	}

	return &updatedPreference, nil
}