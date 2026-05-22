package repositories

import (
	"context"
	"time"

	"disaster_alert_backend/internal/models"
	"disaster_alert_backend/internal/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AlertRepository struct {
	collection *mongo.Collection
}

type AlertFilter struct {
	Category   string
	Severity   string
	SourceName string
	SourceType string
	Country    string
	Limit      int
}

func NewAlertRepository(db *mongo.Database) *AlertRepository {
	return &AlertRepository{
		collection: db.Collection("alerts"),
	}
}

func (r *AlertRepository) Create(alert models.Alert) (*models.Alert, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := r.collection.InsertOne(ctx, alert)
	if err != nil {
		return nil, err
	}

	insertedID, ok := result.InsertedID.(primitive.ObjectID)
	if ok {
		alert.ID = insertedID
	}

	return &alert, nil
}

func (r *AlertRepository) FindAll() ([]models.Alert, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "createdAt", Value: -1}})

	cursor, err := r.collection.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	alerts := make([]models.Alert, 0)

	for cursor.Next(ctx) {
		var alert models.Alert

		if err := cursor.Decode(&alert); err != nil {
			return nil, err
		}

		alerts = append(alerts, alert)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return alerts, nil
}

func (r *AlertRepository) FindActive() ([]models.Alert, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	now := time.Now()

	filter := bson.M{
		"status": "active",
		"$or": []bson.M{
			{"expiresAt": bson.M{"$exists": false}},
			{"expiresAt": nil},
			{"expiresAt": bson.M{"$gt": now}},
		},
	}

	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "createdAt", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	alerts := make([]models.Alert, 0)

	for cursor.Next(ctx) {
		var alert models.Alert

		if err := cursor.Decode(&alert); err != nil {
			return nil, err
		}

		alerts = append(alerts, alert)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return alerts, nil
}

func (r *AlertRepository) FindByID(alertID primitive.ObjectID) (*models.Alert, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var alert models.Alert

	err := r.collection.FindOne(ctx, bson.M{
		"_id": alertID,
	}).Decode(&alert)

	if err != nil {
		return nil, err
	}

	return &alert, nil
}

func (r *AlertRepository) UpdateStatus(
	alertID primitive.ObjectID,
	status string,
) (*models.Alert, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.FindOneAndUpdate().
		SetReturnDocument(options.After)

	var updatedAlert models.Alert

	err := r.collection.FindOneAndUpdate(
		ctx,
		bson.M{"_id": alertID},
		bson.M{
			"$set": bson.M{
				"status":    status,
				"updatedAt": time.Now(),
			},
		},
		opts,
	).Decode(&updatedAlert)

	if err != nil {
		return nil, err
	}

	return &updatedAlert, nil
}

func (r *AlertRepository) Delete(alertID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := r.collection.DeleteOne(ctx, bson.M{
		"_id": alertID,
	})

	return err
}

func (r *AlertRepository) UpsertExternalAlert(alert models.Alert) (*models.Alert, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"sourceType": "external",
		"sourceName": alert.SourceName,
		"externalId": alert.ExternalID,
	}

	update := bson.M{
		"$set": bson.M{
			"title":              alert.Title,
			"description":        alert.Description,
			"category":           alert.Category,
			"severity":           alert.Severity,
			"status":             alert.Status,
			"location":           alert.Location,
			"radiusKm":           alert.RadiusKm,
			"safetyInstructions": alert.SafetyInstructions,
			"sourceUrl":          alert.SourceURL,
			"eventTime":          alert.EventTime,
			"expiresAt":          alert.ExpiresAt,
			"confidence":         alert.Confidence,
			"updatedAt":          alert.UpdatedAt,
		},
		"$setOnInsert": bson.M{
			"sourceType": alert.SourceType,
			"sourceName": alert.SourceName,
			"externalId": alert.ExternalID,
			"createdAt":  alert.CreatedAt,
		},
	}

	opts := options.FindOneAndUpdate().
		SetUpsert(true).
		SetReturnDocument(options.After)

	var upsertedAlert models.Alert

	err := r.collection.FindOneAndUpdate(
		ctx,
		filter,
		update,
		opts,
	).Decode(&upsertedAlert)

	if err != nil {
		return nil, err
	}

	return &upsertedAlert, nil
}

func (r *AlertRepository) FindActiveNearby(
	latitude float64,
	longitude float64,
	radiusKm float64,
) ([]models.Alert, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	now := time.Now()

	filter := bson.M{
		"status": "active",
		"$or": []bson.M{
			{"expiresAt": bson.M{"$exists": false}},
			{"expiresAt": nil},
			{"expiresAt": bson.M{"$gt": now}},
		},
	}

	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "createdAt", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}

	defer cursor.Close(ctx)

	alerts := make([]models.Alert, 0)

	for cursor.Next(ctx) {
		var alert models.Alert

		if err := cursor.Decode(&alert); err != nil {
			return nil, err
		}

		distanceKm := utils.DistanceKm(
			latitude,
			longitude,
			alert.Location.Latitude,
			alert.Location.Longitude,
		)

		if distanceKm <= radiusKm {
			alerts = append(alerts, alert)
		}
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return alerts, nil
}


func (r *AlertRepository) FindCriticalGlobal(limit int) ([]models.Alert, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	now := time.Now().UTC()

	if limit <= 0 {
		limit = 20
	}

	filter := bson.M{
		"status": "active",
		"severity": bson.M{
			"$in": []string{"high", "critical"},
		},
		"$or": []bson.M{
			{"expiresAt": bson.M{"$gt": now}},
			{"expiresAt": bson.M{"$exists": false}},
			{"expiresAt": nil},
		},
	}

	opts := options.Find().
		SetSort(bson.D{
			{Key: "severity", Value: 1},
			{Key: "eventTime", Value: -1},
			{Key: "updatedAt", Value: -1},
		}).
		SetLimit(int64(limit))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var alerts []models.Alert

	if err := cursor.All(ctx, &alerts); err != nil {
		return nil, err
	}

	return alerts, nil
}

func (r *AlertRepository) FindActiveWithFilters(filterOptions AlertFilter) ([]models.Alert, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	now := time.Now().UTC()

	if filterOptions.Limit <= 0 {
		filterOptions.Limit = 50
	}

	if filterOptions.Limit > 200 {
		filterOptions.Limit = 200
	}

	filter := bson.M{
		"status": "active",
		"$or": []bson.M{
			{"expiresAt": bson.M{"$gt": now}},
			{"expiresAt": bson.M{"$exists": false}},
			{"expiresAt": nil},
		},
	}

	if filterOptions.Category != "" {
		filter["category"] = filterOptions.Category
	}

	if filterOptions.Severity != "" {
		filter["severity"] = filterOptions.Severity
	}

	if filterOptions.SourceName != "" {
		filter["sourceName"] = filterOptions.SourceName
	}

	if filterOptions.SourceType != "" {
		filter["sourceType"] = filterOptions.SourceType
	}

	if filterOptions.Country != "" {
		filter["location.country"] = filterOptions.Country
	}

	opts := options.Find().
		SetSort(bson.D{
			{Key: "severity", Value: 1},
			{Key: "eventTime", Value: -1},
			{Key: "updatedAt", Value: -1},
		}).
		SetLimit(int64(filterOptions.Limit))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var alerts []models.Alert

	if err := cursor.All(ctx, &alerts); err != nil {
		return nil, err
	}

	return alerts, nil
}


func (r *AlertRepository) FindActiveNearbyWithFilters(
	lat float64,
	lng float64,
	radiusKm float64,
	filterOptions AlertFilter,
) ([]models.Alert, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	now := time.Now().UTC()

	if filterOptions.Limit <= 0 {
		filterOptions.Limit = 50
	}

	if filterOptions.Limit > 200 {
		filterOptions.Limit = 200
	}

	filter := bson.M{
		"status": "active",
		"$or": []bson.M{
			{"expiresAt": bson.M{"$gt": now}},
			{"expiresAt": bson.M{"$exists": false}},
			{"expiresAt": nil},
		},
	}

	if filterOptions.Category != "" {
		filter["category"] = filterOptions.Category
	}

	if filterOptions.Severity != "" {
		filter["severity"] = filterOptions.Severity
	}

	if filterOptions.SourceName != "" {
		filter["sourceName"] = filterOptions.SourceName
	}

	if filterOptions.SourceType != "" {
		filter["sourceType"] = filterOptions.SourceType
	}

	if filterOptions.Country != "" {
		filter["location.country"] = filterOptions.Country
	}

	opts := options.Find().
		SetSort(bson.D{
			{Key: "eventTime", Value: -1},
			{Key: "updatedAt", Value: -1},
		})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var allAlerts []models.Alert

	if err := cursor.All(ctx, &allAlerts); err != nil {
		return nil, err
	}

	nearbyAlerts := make([]models.Alert, 0)

	for _, alert := range allAlerts {
		distance := utils.DistanceKm(
			lat,
			lng,
			alert.Location.Latitude,
			alert.Location.Longitude,
		)

		if distance <= radiusKm {
			nearbyAlerts = append(nearbyAlerts, alert)
		}

		if len(nearbyAlerts) >= filterOptions.Limit {
			break
		}
	}

	return nearbyAlerts, nil
}

func (r *AlertRepository) DeactivateExpiredExternalAlerts() (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	now := time.Now().UTC()

	filter := bson.M{
		"sourceType": "external",
		"status":     "active",
		"expiresAt": bson.M{
			"$lt": now,
		},
	}

	update := bson.M{
		"$set": bson.M{
			"status":    "expired",
			"updatedAt": now,
		},
	}

	result, err := r.collection.UpdateMany(ctx, filter, update)
	if err != nil {
		return 0, err
	}

	return result.ModifiedCount, nil
}