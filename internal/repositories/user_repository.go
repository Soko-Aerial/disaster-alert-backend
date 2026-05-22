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


type UserRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(db *mongo.Database) *UserRepository {
	return &UserRepository{
		collection: db.Collection("users"),
	}
}

func (r *UserRepository) CreateUser(user models.User) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		return nil, err
	}

	insertedID, ok := result.InsertedID.(primitive.ObjectID)

	if ok {
		user.ID =insertedID
	}

	return &user, nil
}

func (r *UserRepository) FindUserByEmail(email string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var user models.User

	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) FindUserByID(userID primitive.ObjectID) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var user models.User

	err := r.collection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) FindLocationByUserID(
	userID primitive.ObjectID,
) (*models.UserLocation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"_id": userID,
	}

	var user models.User
	err := r.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		return nil, err
	}

	return user.Location, nil
}

func (r *UserRepository) UpdateLocationByUserID(
	userID primitive.ObjectID,
	location models.UserLocation,
) (*models.UserLocation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	now := time.Now().UTC()

	filter := bson.M{
		"_id": userID,
	}

	update := bson.M{
		"$set": bson.M{
			"location":  location,
			"updatedAt": now,
		},
	}

	opts := options.FindOneAndUpdate().
		SetReturnDocument(options.After)

	var updatedUser models.User
	err := r.collection.FindOneAndUpdate(
		ctx,
		filter,
		update,
		opts,
	).Decode(&updatedUser)

	if err != nil {
		return nil, err
	}

	return updatedUser.Location, nil
}

func (r *UserRepository) FindProfileDetailsByUserID(
	userID primitive.ObjectID,
) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"_id": userID,
	}

	var user models.User
	err := r.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) UpdateProfileDetailsByUserID(
	userID primitive.ObjectID,
	update bson.M,
) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	update["updatedAt"] = time.Now().UTC()

	filter := bson.M{
		"_id": userID,
	}

	opts := options.FindOneAndUpdate().
		SetReturnDocument(options.After)

	var updatedUser models.User
	err := r.collection.FindOneAndUpdate(
		ctx,
		filter,
		bson.M{"$set": update},
		opts,
	).Decode(&updatedUser)

	if err != nil {
		return nil, err
	}

	return &updatedUser, nil
}