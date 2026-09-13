package repositories

import (
	"context"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AlertTargetCandidate struct {
	UserID primitive.ObjectID

	HasLocation bool
	Latitude    float64
	Longitude   float64
	Country     string
	Region      string
	District    string

	LocationUpdatedAt *time.Time

	LocationSharingEnabled bool

	HasFCMToken bool

	AlertPreferences map[string]bool
}

func (r *UserRepository) FindAlertTargetCandidates(
	country string,
	region string,
	limit int64,
) ([]AlertTargetCandidate, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{}

	cleanCountry := strings.TrimSpace(country)
	cleanRegion := strings.TrimSpace(region)

	if cleanCountry != "" {
		filter["$or"] = []bson.M{
			{"location.country": cleanCountry},
			{"country": cleanCountry},
		}
	}

	if cleanRegion != "" {
		if _, exists := filter["$or"]; exists {
			filter = bson.M{
				"$and": []bson.M{
					filter,
					{
						"$or": []bson.M{
							{"location.region": cleanRegion},
							{"region": cleanRegion},
						},
					},
				},
			}
		} else {
			filter["$or"] = []bson.M{
				{"location.region": cleanRegion},
				{"region": cleanRegion},
			}
		}
	}

	findOptions := options.Find()

	if limit <= 0 {
		limit = 10000
	}

	findOptions.SetLimit(limit)

	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	candidates := make([]AlertTargetCandidate, 0)

	for cursor.Next(ctx) {
		var raw bson.M

		if err := cursor.Decode(&raw); err != nil {
			return nil, err
		}

		candidate := buildAlertTargetCandidateFromBSON(raw)

		if candidate.UserID.IsZero() {
			continue
		}

		candidates = append(candidates, candidate)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return candidates, nil
}

func buildAlertTargetCandidateFromBSON(raw bson.M) AlertTargetCandidate {
	candidate := AlertTargetCandidate{
		LocationSharingEnabled: true,
		HasFCMToken:            true,
		AlertPreferences:       map[string]bool{},
	}

	if id, ok := raw["_id"].(primitive.ObjectID); ok {
		candidate.UserID = id
	}

	location := getMap(raw, "location")

	latitude, hasLatitude := getFloat64FromMaps(raw, location, "latitude")
	longitude, hasLongitude := getFloat64FromMaps(raw, location, "longitude")

	candidate.HasLocation = hasLatitude && hasLongitude && latitude != 0 && longitude != 0
	candidate.Latitude = latitude
	candidate.Longitude = longitude

	candidate.Country = firstNonEmptyString(
		getString(location, "country"),
		getString(raw, "country"),
	)

	candidate.Region = firstNonEmptyString(
		getString(location, "region"),
		getString(raw, "region"),
	)

	candidate.District = firstNonEmptyString(
		getString(location, "district"),
		getString(raw, "district"),
	)

	candidate.LocationUpdatedAt = firstTimePointer(
		getTimePointer(location, "updatedAt"),
		getTimePointer(location, "lastUpdatedAt"),
		getTimePointer(raw, "locationUpdatedAt"),
		getTimePointer(raw, "lastLocationUpdatedAt"),
		getTimePointer(raw, "updatedAt"),
	)

	if value, exists := getBoolIfExists(raw, "locationSharingEnabled"); exists {
		candidate.LocationSharingEnabled = value
	}

	if value, exists := getBoolIfExists(raw, "emergencyLocationAlertsEnabled"); exists {
		candidate.LocationSharingEnabled = value
	}

	if value, exists := getBoolIfExists(location, "sharingEnabled"); exists {
		candidate.LocationSharingEnabled = value
	}

	if value, exists := getBoolIfExists(raw, "hasFcmToken"); exists {
		candidate.HasFCMToken = value
	}

	if value, exists := getBoolIfExists(raw, "hasFCMToken"); exists {
		candidate.HasFCMToken = value
	}

	if token := strings.TrimSpace(getString(raw, "fcmToken")); token != "" {
		candidate.HasFCMToken = true
	}

	candidate.AlertPreferences = readAlertPreferences(raw)

	return candidate
}

func readAlertPreferences(raw bson.M) map[string]bool {
	preferences := map[string]bool{}

	for _, key := range []string{
		"alertPreferences",
		"notificationPreferences",
		"preferences",
	} {
		value, ok := raw[key]
		if !ok {
			continue
		}

		preferenceMap, ok := value.(bson.M)
		if !ok {
			continue
		}

		for preferenceKey, preferenceValue := range preferenceMap {
			boolValue, ok := preferenceValue.(bool)
			if !ok {
				continue
			}

			preferences[strings.ToLower(strings.TrimSpace(preferenceKey))] = boolValue
		}
	}

	return preferences
}

func getMap(raw bson.M, key string) bson.M {
	value, ok := raw[key]
	if !ok {
		return bson.M{}
	}

	if result, ok := value.(bson.M); ok {
		return result
	}

	if result, ok := value.(map[string]interface{}); ok {
		return bson.M(result)
	}

	return bson.M{}
}

func getFloat64FromMaps(
	root bson.M,
	nested bson.M,
	key string,
) (float64, bool) {
	if value, ok := getFloat64IfExists(nested, key); ok {
		return value, true
	}

	if value, ok := getFloat64IfExists(root, key); ok {
		return value, true
	}

	return 0, false
}

func getFloat64IfExists(raw bson.M, key string) (float64, bool) {
	value, ok := raw[key]
	if !ok {
		return 0, false
	}

	switch typed := value.(type) {
	case float64:
		return typed, true
	case float32:
		return float64(typed), true
	case int:
		return float64(typed), true
	case int32:
		return float64(typed), true
	case int64:
		return float64(typed), true
	default:
		return 0, false
	}
}

func getString(raw bson.M, key string) string {
	value, ok := raw[key]
	if !ok || value == nil {
		return ""
	}

	if text, ok := value.(string); ok {
		return strings.TrimSpace(text)
	}

	return ""
}

func getBoolIfExists(raw bson.M, key string) (bool, bool) {
	value, ok := raw[key]
	if !ok {
		return false, false
	}

	boolValue, ok := value.(bool)
	return boolValue, ok
}

func getTimePointer(raw bson.M, key string) *time.Time {
	value, ok := raw[key]
	if !ok || value == nil {
		return nil
	}

	switch typed := value.(type) {
	case time.Time:
		return &typed
	case primitive.DateTime:
		converted := typed.Time()
		return &converted
	default:
		return nil
	}
}

func firstTimePointer(values ...*time.Time) *time.Time {
	for _, value := range values {
		if value != nil && !value.IsZero() {
			return value
		}
	}

	return nil
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}

	return ""
}
