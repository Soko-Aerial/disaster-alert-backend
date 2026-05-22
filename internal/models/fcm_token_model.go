package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FCMToken struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID     primitive.ObjectID `bson:"userId" json:"userId"`
	Token      string             `bson:"token" json:"token"`
	DeviceType string             `bson:"deviceType" json:"deviceType"`
	IsActive   bool               `bson:"isActive" json:"isActive"`
	CreatedAt  time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt  time.Time          `bson:"updatedAt" json:"updatedAt"`
}