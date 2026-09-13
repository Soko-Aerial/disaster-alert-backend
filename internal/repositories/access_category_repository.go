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

type AccessCategoryRepository struct {
	collection *mongo.Collection
}

func NewAccessCategoryRepository(db *mongo.Database) *AccessCategoryRepository {
	return &AccessCategoryRepository{
		collection: db.Collection("access_categories"),
	}
}

func (r *AccessCategoryRepository) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "slug", Value: 1},
				{Key: "kind", Value: 1},
				{Key: "ownerOrganisationId", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{
				{Key: "isActive", Value: 1},
				{Key: "kind", Value: 1},
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

func (r *AccessCategoryRepository) Create(ctx context.Context, category *models.AccessCategory) error {
	result, err := r.collection.InsertOne(ctx, category)
	if err != nil {
		return err
	}

	if objectID, ok := result.InsertedID.(primitive.ObjectID); ok {
		category.ID = objectID
	}

	return nil
}

func (r *AccessCategoryRepository) UpsertSystemCategory(ctx context.Context, category *models.AccessCategory) error {
	now := time.Now().UTC()

	filter := bson.M{
		"slug":                category.Slug,
		"kind":                category.Kind,
		"ownerOrganisationId": category.OwnerOrganisationID,
	}

	update := bson.M{
		"$setOnInsert": bson.M{
			"name":                  category.Name,
			"slug":                  category.Slug,
			"description":           category.Description,
			"kind":                  category.Kind,
			"visibility":            category.Visibility,
			"ownerOrganisationId":   category.OwnerOrganisationID,
			"ownerOrganisationName": category.OwnerOrganisationName,
			"preferenceKey":         category.PreferenceKey,
			"allowedActions":        category.AllowedActions,
			"isSystem":              true,
			"isActive":              true,
			"createdAt":             now,
		},
		"$set": bson.M{
			"updatedAt": now,
		},
	}

	_, err := r.collection.UpdateOne(
		ctx,
		filter,
		update,
		options.Update().SetUpsert(true),
	)

	return err
}

func (r *AccessCategoryRepository) GetAll(
	ctx context.Context,
	kind string,
	ownerOrganisationID string,
	includeInactive bool,
	limit int64,
) ([]models.AccessCategory, error) {
	if limit <= 0 {
		limit = 100
	}

	if limit > 500 {
		limit = 500
	}

	filter := bson.M{}

	if kind != "" {
		filter["kind"] = kind
	}

	if ownerOrganisationID != "" {
		filter["ownerOrganisationId"] = ownerOrganisationID
	}

	if !includeInactive {
		filter["isActive"] = true
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}}).
		SetLimit(limit)

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var categories []models.AccessCategory
	if err := cursor.All(ctx, &categories); err != nil {
		return nil, err
	}

	return categories, nil
}

func (r *AccessCategoryRepository) FindByID(
	ctx context.Context,
	id string,
) (*models.AccessCategory, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var category models.AccessCategory
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&category)

	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &category, nil
}

func (r *AccessCategoryRepository) Update(
	ctx context.Context,
	id string,
	update bson.M,
) (*models.AccessCategory, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	update["updatedAt"] = time.Now().UTC()

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var category models.AccessCategory
	err = r.collection.FindOneAndUpdate(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": update},
		opts,
	).Decode(&category)

	if err != nil {
		return nil, err
	}

	return &category, nil
}

func (r *AccessCategoryRepository) Deactivate(
	ctx context.Context,
	id string,
) (*models.AccessCategory, error) {
	return r.Update(ctx, id, bson.M{
		"isActive": false,
	})
}
