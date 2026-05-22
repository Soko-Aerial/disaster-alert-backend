package database

import(
	"context"
	"fmt"
	"log"
	"time"

	"disaster_alert_backend/config"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)


type MongoDB struct{
	Client *mongo.Client
	Database *mongo.Database
}


func ConnectMongoDB(cfg *config.Config) (*MongoDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(cfg.MongoURI)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}


	database := client.Database(cfg.MongoDatabase)

	log.Println("MongoDB connected successfully")

	return &MongoDB{
		Client:   client,
		Database: database,
	}, nil
	
}

func (m *MongoDB) Disconnect() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := m.Client.Disconnect(ctx)
	if err != nil {
		log.Println("Failed to disconnect MongoDB:", err)
		return
	}

	fmt.Println("MongoDB disconnected")
}