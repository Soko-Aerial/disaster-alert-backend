package config

import (
	"context"
	"errors"
	"log"
	"strings"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

type FirebaseApp struct {
	App             *firebase.App
	MessagingClient *messaging.Client
}

func InitFirebase(cfg *Config) (*FirebaseApp, error) {
	ctx := context.Background()

	credentialsJSON := strings.TrimSpace(cfg.FirebaseCredentialsJSON)
	credentialsFile := strings.TrimSpace(cfg.FirebaseCredentialsFile)

	var opt option.ClientOption

	if credentialsJSON != "" {
		opt = option.WithCredentialsJSON([]byte(credentialsJSON))
		log.Println("Firebase Admin SDK using credentials from FIREBASE_CREDENTIALS_JSON")
	} else if credentialsFile != "" {
		opt = option.WithCredentialsFile(credentialsFile)
		log.Println("Firebase Admin SDK using credentials file:", credentialsFile)
	} else {
		return nil, errors.New("firebase credentials are not configured")
	}

	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, err
	}

	messagingClient, err := app.Messaging(ctx)
	if err != nil {
		return nil, err
	}

	log.Println("Firebase Admin SDK initialized successfully")

	return &FirebaseApp{
		App:             app,
		MessagingClient: messagingClient,
	}, nil
}