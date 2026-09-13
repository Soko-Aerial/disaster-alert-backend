package services

import (
	"math"

	"disaster_alert_backend/internal/models"
)

const earthRadiusKm = 6371.0

func alertDistanceKm(
	lat1 float64,
	lon1 float64,
	lat2 float64,
	lon2 float64,
) float64 {
	lat1Rad := degreesToRadians(lat1)
	lon1Rad := degreesToRadians(lon1)
	lat2Rad := degreesToRadians(lat2)
	lon2Rad := degreesToRadians(lon2)

	dLat := lat2Rad - lat1Rad
	dLon := lon2Rad - lon1Rad

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusKm * c
}

func degreesToRadians(value float64) float64 {
	return value * math.Pi / 180
}

func pointInsidePolygon(
	latitude float64,
	longitude float64,
	polygon []models.AlertGeoPoint,
) bool {
	if len(polygon) < 3 {
		return false
	}

	inside := false
	j := len(polygon) - 1

	for i := 0; i < len(polygon); i++ {
		latI := polygon[i].Latitude
		lonI := polygon[i].Longitude
		latJ := polygon[j].Latitude
		lonJ := polygon[j].Longitude

		intersects := ((lonI > longitude) != (lonJ > longitude)) &&
			(latitude < (latJ-latI)*(longitude-lonI)/(lonJ-lonI)+latI)

		if intersects {
			inside = !inside
		}

		j = i
	}

	return inside
}

func polygonCentroid(polygon []models.AlertGeoPoint) models.AlertGeoPoint {
	if len(polygon) == 0 {
		return models.AlertGeoPoint{}
	}

	var latitudeSum float64
	var longitudeSum float64

	for _, point := range polygon {
		latitudeSum += point.Latitude
		longitudeSum += point.Longitude
	}

	return models.AlertGeoPoint{
		Latitude:  latitudeSum / float64(len(polygon)),
		Longitude: longitudeSum / float64(len(polygon)),
	}
}
