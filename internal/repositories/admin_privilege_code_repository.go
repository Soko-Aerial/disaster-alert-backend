package repositories

import (
	"context"
	"errors"
	"time"

	"disaster_alert_backend/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AdminPrivilegeCodeRepository struct {
	collection *mongo.Collection
}

func NewAdminPrivilegeCodeRepository(db *mongo.Database) *AdminPrivilegeCodeRepository {
	return &AdminPrivilegeCodeRepository{
		collection: db.Collection("admin_privilege_codes"),
	}
}

func (r *AdminPrivilegeCodeRepository) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "codeHash", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{
				{Key: "status", Value: 1},
				{Key: "organisationId", Value: 1},
				{Key: "levelId", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "createdAt", Value: -1},
			},
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	return err
}

func (r *AdminPrivilegeCodeRepository) Create(
	ctx context.Context,
	code *models.AdminPrivilegeCode,
) error {
	result, err := r.collection.InsertOne(ctx, code)
	if err != nil {
		return err
	}

	if objectID, ok := result.InsertedID.(primitive.ObjectID); ok {
		code.ID = objectID
	}

	return nil
}

func (r *AdminPrivilegeCodeRepository) FindByCodeHash(
	ctx context.Context,
	codeHash string,
) (*models.AdminPrivilegeCode, error) {
	var code models.AdminPrivilegeCode

	err := r.collection.FindOne(ctx, bson.M{
		"codeHash": codeHash,
	}).Decode(&code)

	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &code, nil
}

func (r *AdminPrivilegeCodeRepository) FindByID(
	ctx context.Context,
	id string,
) (*models.AdminPrivilegeCode, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var code models.AdminPrivilegeCode

	err = r.collection.FindOne(ctx, bson.M{
		"_id": objectID,
	}).Decode(&code)

	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &code, nil
}

func (r *AdminPrivilegeCodeRepository) GetAll(
	ctx context.Context,
	limit int64,
) ([]models.AdminPrivilegeCode, error) {
	if limit <= 0 {
		limit = 50
	}

	if limit > 200 {
		limit = 200
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}}).
		SetLimit(limit)

	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}

	defer cursor.Close(ctx)

	var codes []models.AdminPrivilegeCode

	if err := cursor.All(ctx, &codes); err != nil {
		return nil, err
	}

	return codes, nil
}

func (r *AdminPrivilegeCodeRepository) Revoke(
	ctx context.Context,
	id string,
	revokedBy string,
	reason string,
) (*models.AdminPrivilegeCode, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()

	update := bson.M{
		"$set": bson.M{
			"status":       models.PrivilegeCodeStatusRevoked,
			"revokedAt":    now,
			"revokedBy":    revokedBy,
			"revokeReason": reason,
			"updatedAt":    now,
		},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var updated models.AdminPrivilegeCode

	err = r.collection.FindOneAndUpdate(
		ctx,
		bson.M{"_id": objectID},
		update,
		opts,
	).Decode(&updated)

	if err != nil {
		return nil, err
	}

	return &updated, nil
}

func (r *AdminPrivilegeCodeRepository) MarkUsed(
	ctx context.Context,
	id primitive.ObjectID,
) error {
	now := time.Now().UTC()

	_, err := r.collection.UpdateByID(
		ctx,
		id,
		bson.M{
			"$inc": bson.M{
				"usageCount": 1,
			},
			"$set": bson.M{
				"lastUsedAt": now,
				"updatedAt":  now,
			},
		},
	)

	return err
}