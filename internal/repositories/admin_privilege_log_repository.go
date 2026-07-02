package repositories

import (
	"context"

	"disaster_alert_backend/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AdminPrivilegeLogRepository struct {
	collection *mongo.Collection
}

func NewAdminPrivilegeLogRepository(db *mongo.Database) *AdminPrivilegeLogRepository {
	return &AdminPrivilegeLogRepository{
		collection: db.Collection("admin_privilege_logs"),
	}
}

func (r *AdminPrivilegeLogRepository) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "createdAt", Value: -1},
			},
		},
		{
			Keys: bson.D{
				{Key: "privilegeCodeId", Value: 1},
				{Key: "createdAt", Value: -1},
			},
		},
		{
			Keys: bson.D{
				{Key: "action", Value: 1},
				{Key: "allowed", Value: 1},
			},
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	return err
}

func (r *AdminPrivilegeLogRepository) Create(
	ctx context.Context,
	log *models.AdminPrivilegeLog,
) error {
	_, err := r.collection.InsertOne(ctx, log)
	return err
}

func (r *AdminPrivilegeLogRepository) GetAll(
	ctx context.Context,
	limit int64,
	privilegeCodeID string,
) ([]models.AdminPrivilegeLog, error) {
	if limit <= 0 {
		limit = 100
	}

	if limit > 500 {
		limit = 500
	}

	filter := bson.M{}

	if privilegeCodeID != "" {
		filter["privilegeCodeId"] = privilegeCodeID
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}}).
		SetLimit(limit)

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}

	defer cursor.Close(ctx)

	var logs []models.AdminPrivilegeLog

	if err := cursor.All(ctx, &logs); err != nil {
		return nil, err
	}

	return logs, nil
}