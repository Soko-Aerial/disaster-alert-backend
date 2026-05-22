package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AssistanceLocation struct {
	Latitude  float64 `bson:"latitude" json:"latitude"`
	Longitude float64 `bson:"longitude" json:"longitude"`
	Address   string  `bson:"address,omitempty" json:"address,omitempty"`
}

type AssistanceRequest struct {
	ID                  primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID              primitive.ObjectID `bson:"userId" json:"userId"`
	AssistanceType      string             `bson:"assistanceType" json:"assistanceType"`
	UrgencyLevel         string             `bson:"urgencyLevel" json:"urgencyLevel"`
	AffectedIndividuals  int                `bson:"affectedIndividuals" json:"affectedIndividuals"`
	OtherInformation     string             `bson:"otherInformation,omitempty" json:"otherInformation,omitempty"`
	Location             AssistanceLocation `bson:"location" json:"location"`
	Status               string             `bson:"status" json:"status"`
	CreatedAt            time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt            time.Time          `bson:"updatedAt" json:"updatedAt"`
}