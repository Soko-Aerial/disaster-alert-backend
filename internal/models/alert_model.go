package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AlertLocation struct {
	Latitude  float64 `bson:"latitude" json:"latitude"`
	Longitude float64 `bson:"longitude" json:"longitude"`
	Address   string  `bson:"address,omitempty" json:"address,omitempty"`
	Country   string  `bson:"country,omitempty" json:"country,omitempty"`
	Region    string  `bson:"region,omitempty" json:"region,omitempty"`
}

type Alert struct {
	ID                 primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	Title              string              `bson:"title" json:"title"`
	Description        string              `bson:"description" json:"description"`
	Category           string              `bson:"category" json:"category"`
	Severity           string              `bson:"severity" json:"severity"`
	Status             string              `bson:"status" json:"status"`
	Location           AlertLocation       `bson:"location" json:"location"`
	RadiusKm           float64             `bson:"radiusKm" json:"radiusKm"`
	SafetyInstructions []string            `bson:"safetyInstructions,omitempty" json:"safetyInstructions,omitempty"`
	SourceType         string              `bson:"sourceType" json:"sourceType"`
	SourceName         string              `bson:"sourceName" json:"sourceName"`
	ExternalID         string              `bson:"externalId,omitempty" json:"externalId,omitempty"`
	SourceURL          string              `bson:"sourceUrl,omitempty" json:"sourceUrl,omitempty"`
	LinkedReportID     *primitive.ObjectID `bson:"linkedReportId,omitempty" json:"linkedReportId,omitempty"`
	LinkedSOSID        *primitive.ObjectID `bson:"linkedSosId,omitempty" json:"linkedSosId,omitempty"`
	LinkedAssistanceID *primitive.ObjectID `bson:"linkedAssistanceId,omitempty" json:"linkedAssistanceId,omitempty"`
	CreatedBy         *primitive.ObjectID `bson:"createdBy,omitempty" json:"createdBy,omitempty"`
	EventTime          *time.Time          `bson:"eventTime,omitempty" json:"eventTime,omitempty"`
	ExpiresAt         *time.Time          `bson:"expiresAt,omitempty" json:"expiresAt,omitempty"`
	Confidence         float64             `bson:"confidence,omitempty" json:"confidence,omitempty"`
	CreatedAt          time.Time           `bson:"createdAt" json:"createdAt"`
	UpdatedAt          time.Time           `bson:"updatedAt" json:"updatedAt"`
}