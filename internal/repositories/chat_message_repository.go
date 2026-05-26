package repositories

import (
	"context"
	"time"

	"disaster_alert_backend/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ChatMessageRepository struct {
	collection *mongo.Collection
}

func NewChatMessageRepository(db *mongo.Database) *ChatMessageRepository {
	return &ChatMessageRepository{
		collection: db.Collection("chat_messages"),
	}
}

func (r *ChatMessageRepository) Create(message models.ChatMessage) (*models.ChatMessage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	now := time.Now().UTC()

	message.ID = primitive.NewObjectID()
	message.CreatedAt = now
	message.UpdatedAt = now

	if message.Status == "" {
		message.Status = "sent"
	}

	_, err := r.collection.InsertOne(ctx, message)
	if err != nil {
		return nil, err
	}

	return &message, nil
}

func (r *ChatMessageRepository) FindByConversationID(
	conversationID primitive.ObjectID,
) ([]models.ChatMessage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"conversationId": conversationID,
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	messages := make([]models.ChatMessage, 0)

	for cursor.Next(ctx) {
		var message models.ChatMessage
		if err := cursor.Decode(&message); err != nil {
			return nil, err
		}

		messages = append(messages, message)
	}

	return messages, cursor.Err()
}

func (r *ChatMessageRepository) MarkMessagesRead(
	conversationID primitive.ObjectID,
	userID primitive.ObjectID,
	messageIDs []primitive.ObjectID,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"conversationId": conversationID,
		"_id": bson.M{
			"$in": messageIDs,
		},
	}

	update := bson.M{
		"$set": bson.M{
			"status":    "read",
			"updatedAt": time.Now().UTC(),
		},
		"$addToSet": bson.M{
			"readBy": userID,
		},
	}

	_, err := r.collection.UpdateMany(ctx, filter, update)
	return err
}