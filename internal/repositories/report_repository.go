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

type ReportRepository struct {
	collection *mongo.Collection
}

func NewReportRepository(db *mongo.Database) *ReportRepository {
	return &ReportRepository{
		collection: db.Collection("reports"),
	}
}

func (r *ReportRepository) Create(report models.Report) (*models.Report, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := r.collection.InsertOne(ctx, report)
	if err != nil {
		return nil, err
	}

	insertedID, ok := result.InsertedID.(primitive.ObjectID)
	if ok {
		report.ID = insertedID
	}

	return &report, nil
}

func (r *ReportRepository) FindAll() ([]models.Report, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "createdAt", Value: -1}})

	cursor, err := r.collection.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		return nil, err
	}

	defer cursor.Close(ctx)

	var reports []models.Report

	for cursor.Next(ctx) {
		var report models.Report

		if err := cursor.Decode(&report); err != nil {
			return nil, err
		}

		reports = append(reports, report)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return reports, nil
}

func (r *ReportRepository) FindByID(reportID primitive.ObjectID) (*models.Report, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var report models.Report

	err := r.collection.FindOne(ctx, bson.M{
		"_id": reportID,
	}).Decode(&report)

	if err != nil {
		return nil, err
	}

	return &report, nil
}

func (r *ReportRepository) UpdateStatus(
	reportID primitive.ObjectID,
	status string,
) (*models.Report, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	now := time.Now().UTC()

	update := bson.M{
		"$set": bson.M{
			"status":    status,
			"updatedAt": now,
		},
	}

	options := options.FindOneAndUpdate().
		SetReturnDocument(options.After)

	var report models.Report

	err := r.collection.FindOneAndUpdate(
		ctx,
		bson.M{"_id": reportID},
		update,
		options,
	).Decode(&report)

	if err != nil {
		return nil, err
	}

	return &report, nil
}