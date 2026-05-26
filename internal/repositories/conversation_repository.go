package repositories

import (
	"context"
	"time"

	"disaster_alert_backend/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ConversationRepository struct {
	collection *mongo.Collection
}

func NewConversationRepository(db *mongo.Database) *ConversationRepository {
	return &ConversationRepository{
		collection: db.Collection("conversations"),
	}
}

func (r *ConversationRepository) Create(conversation models.Conversation) (*models.Conversation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	now := time.Now().UTC()

	conversation.ID = primitive.NewObjectID()
	conversation.CreatedAt = now
	conversation.UpdatedAt = now

	if conversation.Status == "" {
		conversation.Status = "open"
	}

	_, err := r.collection.InsertOne(ctx, conversation)
	if err != nil {
		return nil, err
	}

	return &conversation, nil
}

func (r *ConversationRepository) FindByUserID(userID primitive.ObjectID) ([]models.Conversation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"userId": userID,
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	conversations := make([]models.Conversation, 0)

	for cursor.Next(ctx) {
		var conversation models.Conversation
		if err := cursor.Decode(&conversation); err != nil {
			return nil, err
		}

		conversations = append(conversations, conversation)
	}

	return conversations, cursor.Err()
}

func (r *ConversationRepository) FindByID(id primitive.ObjectID, userID primitive.ObjectID) (*models.Conversation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"_id":    id,
		"userId": userID,
	}

	var conversation models.Conversation

	err := r.collection.FindOne(ctx, filter).Decode(&conversation)
	if err != nil {
		return nil, err
	}

	return &conversation, nil
}

func (r *ConversationRepository) UpdateLastMessage(
	id primitive.ObjectID,
	lastMessage string,
	lastMessageAt time.Time,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"lastMessage":   lastMessage,
			"lastMessageAt": lastMessageAt,
			"updatedAt":     time.Now().UTC(),
		},
	}

	_, err := r.collection.UpdateByID(ctx, id, update)
	return err
}