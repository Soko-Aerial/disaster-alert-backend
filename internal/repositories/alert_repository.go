package repositories

import (
	"context"
	"math"
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
	Category       string
	Severity       string
	SourceName     string
	SourceType     string
	Country        string
	ExcludeCountry string
	Limit          int
}

func NewAlertRepository(db *mongo.Database) *AlertRepository {
	return &AlertRepository{
		collection: db.Collection("alerts"),
	}
}

func (r *AlertRepository) Create(alert models.Alert) (*models.Alert, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	now := time.Now().UTC()

	if alert.ID.IsZero() {
		alert.ID = primitive.NewObjectID()
	}

	if alert.CreatedAt.IsZero() {
		alert.CreatedAt = now
	}

	alert.UpdatedAt = now

	if alert.Status == "" {
		alert.Status = "active"
	}

	alert.Location.Country = utils.NormalizeCountryCode(alert.Location.Country)

	if alert.ExpiresAt == nil {
		expiresAt := now.Add(7 * 24 * time.Hour)
		alert.ExpiresAt = &expiresAt
	}

	_, err := r.collection.InsertOne(ctx, alert)
	if err != nil {
		return nil, err
	}

	return &alert, nil
}

func (r *AlertRepository) FindAll() ([]models.Alert, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	findOptions := options.Find()
	findOptions.SetSort(alertPrioritySort())
	findOptions.SetLimit(100)

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
	return r.FindActiveWithFilters(AlertFilter{
		Limit: 100,
	})
}

func (r *AlertRepository) FindActiveWithFilters(
	filterOptions AlertFilter,
) ([]models.Alert, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := buildActiveAlertFilter(filterOptions)

	findOptions := options.Find()
	findOptions.SetSort(alertPrioritySort())

	if filterOptions.Limit > 0 {
		findOptions.SetLimit(int64(filterOptions.Limit))
	} else {
		findOptions.SetLimit(50)
	}

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

func (r *AlertRepository) FindByID(
	alertID primitive.ObjectID,
) (*models.Alert, error) {
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
				"updatedAt": time.Now().UTC(),
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

func (r *AlertRepository) UpsertExternalAlert(
	alert models.Alert,
) (*models.Alert, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	now := time.Now().UTC()

	alert.Location.Country = utils.NormalizeCountryCode(alert.Location.Country)

	if alert.Status == "" {
		alert.Status = "active"
	}

	if alert.ExpiresAt == nil {
		baseTime := now

		if alert.EventTime != nil {
			baseTime = *alert.EventTime
		}

		expiresAt := baseTime.Add(7 * 24 * time.Hour)
		alert.ExpiresAt = &expiresAt
	}

	if alert.CreatedAt.IsZero() {
		alert.CreatedAt = now
	}

	alert.UpdatedAt = now

	filter := bson.M{
		"sourceType": alert.SourceType,
		"sourceName": alert.SourceName,
		"externalId": alert.ExternalID,
	}

	update := bson.M{
		"$set": bson.M{
			"title":              alert.Title,
			"description":        alert.Description,
			"summary":            alert.Summary,
			"category":           alert.Category,
			"severity":           alert.Severity,
			"status":             alert.Status,
			"location":           alert.Location,
			"radiusKm":           alert.RadiusKm,
			"safetyInstructions": alert.SafetyInstructions,
			"sourceUrl":          alert.SourceURL,
			"imageUrls":          alert.ImageURLs,
			"videoUrls":          alert.VideoURLs,
			"tags":               alert.Tags,
			"priorityScore":      alert.PriorityScore,
			"priorityLabel":      alert.PriorityLabel,
			"isBreaking":         alert.IsBreaking,
			"isVerified":         alert.IsVerified,
			"eventTime":          alert.EventTime,
			"expiresAt":          alert.ExpiresAt,
			"confidence":         alert.Confidence,
			"lastSyncedAt":       alert.LastSyncedAt,
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

func (r *AlertRepository) DeactivateExpiredExternalAlerts() (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	now := time.Now().UTC()

	filter := bson.M{
		"sourceType": bson.M{
			"$in": []string{
				"external",
				"external_api",
				"weather",
				"news",
				"system",
				"user_report",
			},
		},
		"status": "active",
		"expiresAt": bson.M{
			"$lte": now,
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

func (r *AlertRepository) FindActiveNearby(
	latitude float64,
	longitude float64,
	radiusKm float64,
) ([]models.Alert, error) {
	return r.FindActiveNearbyWithFilters(
		latitude,
		longitude,
		radiusKm,
		AlertFilter{},
	)
}

func (r *AlertRepository) FindActiveNearbyWithFilters(
	latitude float64,
	longitude float64,
	radiusKm float64,
	filterOptions AlertFilter,
) ([]models.Alert, error) {
	if radiusKm <= 0 {
		radiusKm = 100
	}

	if filterOptions.Limit <= 0 {
		filterOptions.Limit = 200
	}

	alerts, err := r.FindActiveWithFilters(filterOptions)
	if err != nil {
		return nil, err
	}

	nearbyAlerts := make([]models.Alert, 0)

	for _, alert := range alerts {
		distance := distanceKm(
			latitude,
			longitude,
			alert.Location.Latitude,
			alert.Location.Longitude,
		)

		alertRadius := alert.RadiusKm
		if alertRadius <= 0 {
			alertRadius = radiusKm
		}

		if distance <= radiusKm || distance <= alertRadius {
			nearbyAlerts = append(nearbyAlerts, alert)
		}
	}

	return nearbyAlerts, nil
}

func (r *AlertRepository) FindCriticalGlobal(
	limit int,
) ([]models.Alert, error) {
	if limit <= 0 {
		limit = 20
	}

	return r.FindActiveWithFilters(AlertFilter{
		Severity: "critical",
		Limit:    limit,
	})
}

func (r *AlertRepository) EnsureIndexes() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "status", Value: 1},
				{Key: "expiresAt", Value: 1},
				{Key: "location.country", Value: 1},
				{Key: "createdAt", Value: -1},
			},
		},
		{
			Keys: bson.D{
				{Key: "eventTime", Value: -1},
			},
		},
		{
			Keys: bson.D{
				{Key: "status", Value: 1},
				{Key: "priorityScore", Value: -1},
				{Key: "isBreaking", Value: -1},
				{Key: "confidence", Value: -1},
				{Key: "eventTime", Value: -1},
				{Key: "createdAt", Value: -1},
			},
		},
		{
			Keys: bson.D{
				{Key: "sourceType", Value: 1},
				{Key: "sourceName", Value: 1},
				{Key: "externalId", Value: 1},
			},
			Options: options.Index().
				SetUnique(true).
				SetPartialFilterExpression(bson.M{
					"sourceType": "external",
					"externalId": bson.M{
						"$type": "string",
						"$gt":   "",
					},
				}),
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	return err
}

func buildActiveAlertFilter(filterOptions AlertFilter) bson.M {
	now := time.Now().UTC()

	filter := bson.M{
		"status": "active",
		"$or": []bson.M{
			{"expiresAt": bson.M{"$exists": false}},
			{"expiresAt": nil},
			{"expiresAt": bson.M{"$gt": now}},
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
		filter["location.country"] = utils.NormalizeCountryCode(
			filterOptions.Country,
		)
	} else if filterOptions.ExcludeCountry != "" {
		filter["location.country"] = bson.M{
			"$ne": utils.NormalizeCountryCode(
				filterOptions.ExcludeCountry,
			),
			"$exists": true,
		}
	}

	return filter
}

func distanceKm(
	lat1 float64,
	lon1 float64,
	lat2 float64,
	lon2 float64,
) float64 {
	const earthRadiusKm = 6371.0

	dLat := degreesToRadians(lat2 - lat1)
	dLon := degreesToRadians(lon2 - lon1)

	lat1Rad := degreesToRadians(lat1)
	lat2Rad := degreesToRadians(lat2)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusKm * c
}

func degreesToRadians(value float64) float64 {
	return value * math.Pi / 180
}

func alertPrioritySort() bson.D {
	return bson.D{
		{Key: "priorityScore", Value: -1},
		{Key: "isBreaking", Value: -1},
		{Key: "confidence", Value: -1},
		{Key: "eventTime", Value: -1},
		{Key: "createdAt", Value: -1},
	}
}
