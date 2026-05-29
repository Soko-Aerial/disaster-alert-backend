package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AppNotification struct {
	ID 				primitive.ObjectID 	`json:"id" bson:"_id,omitempty"`

	RecipientID 	primitive.ObjectID 	`json:"recipientId" bson:"recipientId"`
	RecipientRole 	string 				`json:"recipientRole" bson:"recipientRole"`

	Title 			string 				`json:"title" bson:"title"`
	Body  			string 				`json:"body" bson:"body"`

	Type 			string 				`json:"type" bson:"type"`

	ReferenceID 	string 				`json:"referenceId,omitempty" bson:"referenceId,omitempty"`

	Data 			map[string]string 	`json:"data,omitempty" bson:"data,omitempty"`

	IsRead 			bool 				`json:"isRead" bson:"isRead"`

	CreatedAt 		time.Time			`json:"createdAt" bson:"createdAt"`
	UpdatedAt 		time.Time 			`json:"updatedAt" bson:"updatedAt"`
}