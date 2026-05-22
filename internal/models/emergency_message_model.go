package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EmergencyMessageContactSnapshot struct {
	ContactID    primitive.ObjectID `json:"contactId" bson:"contactId"`
	Name         string             `json:"name" bson:"name"`
	Phone        string             `json:"phone" bson:"phone"`
	Email        string             `json:"email" bson:"email"`
	Relationship string             `json:"relationship" bson:"relationship"`
	Type         string             `json:"type" bson:"type"`
}

type EmergencyMessageLocation struct {
	Latitude  float64 `json:"latitude" bson:"latitude"`
	Longitude float64 `json:"longitude" bson:"longitude"`
	Address   string  `json:"address" bson:"address"`
}

type EmergencyMessage struct {
	ID     primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID primitive.ObjectID `json:"userId" bson:"userId"`

	Title   string `json:"title" bson:"title"`
	Message string `json:"message" bson:"message"`

	Contacts []EmergencyMessageContactSnapshot `json:"contacts" bson:"contacts"`

	Location *EmergencyMessageLocation `json:"location,omitempty" bson:"location,omitempty"`

	Status string `json:"status" bson:"status"`
	// draft, pending, sent, failed, cancelled

	Type string `json:"type" bson:"type"`
	// free_text, sos, assistance, report, template

	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt"`
	SentAt    *time.Time `json:"sentAt,omitempty" bson:"sentAt,omitempty"`
}