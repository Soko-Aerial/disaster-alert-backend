package repositories

import (
	"context"
	"time"

	"disaster_alert_backend/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type EmergencyContactRepository struct {
	collection *mongo.Collection
}

func NewEmergencyContactRepository(db *mongo.Database) *EmergencyContactRepository {
	return &EmergencyContactRepository{
		collection: db.Collection("emergency_contacts"),
	}
}

func (r *EmergencyContactRepository) Create(contact models.EmergencyContact) (*models.EmergencyContact, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	contact.ID = primitive.NewObjectID()
	contact.CreatedAt = time.Now().UTC()
	contact.UpdatedAt = time.Now().UTC()
	contact.IsActive = true

	_, err := r.collection.InsertOne(ctx, contact)
	if err != nil {
		return nil, err
	}

	return &contact, nil
}

func (r *EmergencyContactRepository) FindByUserID(userID primitive.ObjectID) ([]models.EmergencyContact, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"userId":   userID,
		"isActive": true,
	}

	opts := options.Find().
		SetSort(bson.D{
			{Key: "isPrimary", Value: -1},
			{Key: "createdAt", Value: -1},
		})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	contacts := make([]models.EmergencyContact, 0)
	if err := cursor.All(ctx, &contacts); err != nil {
		return nil, err
	}

	return contacts, nil
}

func (r *EmergencyContactRepository) FindByID(
	id primitive.ObjectID,
	userID primitive.ObjectID,
) (*models.EmergencyContact, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"_id":    id,
		"userId": userID,
	}

	var contact models.EmergencyContact
	err := r.collection.FindOne(ctx, filter).Decode(&contact)
	if err != nil {
		return nil, err
	}

	return &contact, nil
}

func (r *EmergencyContactRepository) Update(
	id primitive.ObjectID,
	userID primitive.ObjectID,
	update bson.M,
) (*models.EmergencyContact, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	update["updatedAt"] = time.Now().UTC()

	filter := bson.M{
		"_id":    id,
		"userId": userID,
	}

	opts := options.FindOneAndUpdate().
		SetReturnDocument(options.After)

	var contact models.EmergencyContact
	err := r.collection.FindOneAndUpdate(
		ctx,
		filter,
		bson.M{"$set": update},
		opts,
	).Decode(&contact)

	if err != nil {
		return nil, err
	}

	return &contact, nil
}

func (r *EmergencyContactRepository) SoftDelete(
	id primitive.ObjectID,
	userID primitive.ObjectID,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"_id":    id,
		"userId": userID,
	}

	update := bson.M{
		"$set": bson.M{
			"isActive":  false,
			"updatedAt": time.Now().UTC(),
		},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}