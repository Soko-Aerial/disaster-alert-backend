package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Conversation struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id"`

	UserID primitive.ObjectID `bson:"userId" json:"userId"`

	CaseID   *primitive.ObjectID `bson:"caseId,omitempty" json:"caseId,omitempty"`
	CaseType string              `bson:"caseType" json:"caseType"`
	// sos, assistance, report, general, contact

	Title string `bson:"title" json:"title"`

	ParticipantIDs []primitive.ObjectID `bson:"participantIds" json:"participantIds"`

	LastMessage   string     `bson:"lastMessage,omitempty" json:"lastMessage,omitempty"`
	LastMessageAt *time.Time `bson:"lastMessageAt,omitempty" json:"lastMessageAt,omitempty"`

	Status string `bson:"status" json:"status"`
	// open, closed, archived

	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time `bson:"updatedAt" json:"updatedAt"`
}

type ChatMessage struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id"`

	ConversationID primitive.ObjectID `bson:"conversationId" json:"conversationId"`

	SenderID   primitive.ObjectID `bson:"senderId" json:"senderId"`
	SenderRole string             `bson:"senderRole" json:"senderRole"`
	// user, admin, responder

	Message string `bson:"message" json:"message"`

	Status string `bson:"status" json:"status"`
	// sent, delivered, read, failed

	ReadBy []primitive.ObjectID `bson:"readBy,omitempty" json:"readBy,omitempty"`

	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time `bson:"updatedAt" json:"updatedAt"`
}