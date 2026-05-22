package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ReportLocation struct {
	Latitude  float64 `bson:"latitude" json:"latitude"`
	Longitude float64 `bson:"longitude" json:"longitude"`
	Address   string  `bson:"address,omitempty" json:"address,omitempty"`
}

type Report struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID           primitive.ObjectID `bson:"userId" json:"userId"`
	Category         string             `bson:"category" json:"category"`
	Description      string             `bson:"description" json:"description"`
	TimeOfOccurrence string             `bson:"timeOfOccurrence" json:"timeOfOccurrence"`
	Location         ReportLocation     `bson:"location" json:"location"`
	MediaURLs        []string           `bson:"mediaUrls,omitempty" json:"mediaUrls,omitempty"`
	Status           string             `bson:"status" json:"status"`
	CreatedAt        time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt        time.Time          `bson:"updatedAt" json:"updatedAt"`
}