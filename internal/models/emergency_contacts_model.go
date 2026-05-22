package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EmergencyContact struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID       primitive.ObjectID `json:"userId" bson:"userId"`

	Name         string             `json:"name" bson:"name"`
	Phone        string             `json:"phone" bson:"phone"`
	Email        string             `json:"email" bson:"email"`
	Relationship string             `json:"relationship" bson:"relationship"`

	Type         string             `json:"type" bson:"type"`
	Organization string             `json:"organization" bson:"organization"`
	Address      string             `json:"address" bson:"address"`

	IsPrimary    bool               `json:"isPrimary" bson:"isPrimary"`
	IsGovernment bool               `json:"isGovernment" bson:"isGovernment"`
	IsActive     bool               `json:"isActive" bson:"isActive"`

	CreatedAt    time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt    time.Time          `json:"updatedAt" bson:"updatedAt"`
}