package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"strings"

	"disaster_alert_backend/config"
	"disaster_alert_backend/internal/app"
	"disaster_alert_backend/internal/database"
	"disaster_alert_backend/internal/handlers"
	"disaster_alert_backend/internal/queue"
	"disaster_alert_backend/internal/repositories"
	"disaster_alert_backend/internal/services"
	"disaster_alert_backend/internal/aggregator"
	"disaster_alert_backend/internal/sources"
	"disaster_alert_backend/internal/scheduler"
)

func main() {
	cfg := config.LoadConfig()

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	db, err := database.ConnectMongoDB(cfg)
	if err != nil {
		log.Fatal("Failed to connect MongoDB:", err)
	}
	defer db.Disconnect()

	firebaseApp, err := config.InitFirebase(cfg)
	if err != nil {
		log.Fatal("Failed to initialize Firebase:", err)
	}

	cloudinaryService, err := services.NewCloudinaryService(cfg)
	if err != nil {
		log.Fatal("Failed to initialize Cloudinary:", err)
	}

	userRepository := repositories.NewUserRepository(db.Database)
	fcmTokenRepository := repositories.NewFCMTokenRepository(db.Database)
	reportRepository := repositories.NewReportRepository(db.Database)
	assistanceRepository := repositories.NewAssistanceRepository(db.Database)
	sosRepository := repositories.NewSOSRepository(db.Database)
	alertRepository := repositories.NewAlertRepository(db.Database)
	emergencyContactRepository := repositories.NewEmergencyContactRepository(db.Database)
	emergencyMessageRepository := repositories.NewEmergencyMessageRepository(db.Database)
	alertPreferenceRepository := repositories.NewAlertPreferenceRepository(db.Database)
	conversationRepo := repositories.NewConversationRepository(db.Database)
	chatMessageRepo := repositories.NewChatMessageRepository(db.Database)


	//mockAlertSource := sources.NewMockAlertSource()
	externalSources := make([]sources.AlertSource, 0)

	if isEnabled(cfg.GDACSEnabled) {
		gdacsSource := sources.NewGDACSSource()
		externalSources = append(externalSources, gdacsSource)
		log.Println("GDACS source enabled")
	} else {
		log.Println("GDACS source disabled")
	}

	if isEnabled(cfg.NASAFIRMSEnabled) {
		nasaFIRMSSource := sources.NewNASAFIRMSSource(
			cfg.NASAFIRMSMapKey,
			cfg.NASAFIRMSSource,
			cfg.NASAFIRMSArea,
			cfg.NASAFIRMSDayRange,
			cfg.NASAFIRMSLimit,
		)

		externalSources = append(externalSources, nasaFIRMSSource)
		log.Println("NASA FIRMS source enabled")
	} else {
		log.Println("NASA FIRMS source disabled")
	}

	if isEnabled(cfg.GDELTEnabled) {
		gdeltSource := sources.NewGDELTSource(
			cfg.GDELTLimit,
			cfg.GDELTTimespan,
		)

		externalSources = append(externalSources, gdeltSource)
		log.Println("GDELT source enabled")
	} else {
		log.Println("GDELT source disabled")
	}

	if isEnabled(cfg.ReliefWebEnabled) {
		reliefWebSource := sources.NewReliefWebSource(
			cfg.ReliefWebAppName,
			cfg.ReliefWebLimit,
		)

		externalSources = append(externalSources, reliefWebSource)
		log.Println("ReliefWeb source enabled")
	} else {
		log.Println("ReliefWeb source disabled")
	}

	alertAggregator := aggregator.NewAlertAggregator(
		alertRepository,
		externalSources,
	)

	if isEnabled(cfg.AlertSyncEnabled) {
		alertSyncScheduler := scheduler.NewAlertSyncSchedulerWithDelay(
			alertAggregator,
			time.Duration(cfg.AlertSyncIntervalMinutes)*time.Minute,
			time.Duration(cfg.AlertSyncInitialDelaySeconds)*time.Second,
		)

		go alertSyncScheduler.Start(ctx)

		log.Printf(
			"Alert sync scheduler enabled. Interval: %dm InitialDelay: %ds\n",
			cfg.AlertSyncIntervalMinutes,
			cfg.AlertSyncInitialDelaySeconds,
		)
	} else {
		log.Println("Alert sync scheduler disabled")
	}

	passwordService := services.NewPasswordService()
	jwtService := services.NewJWTService(cfg)

	authService := services.NewAuthService(
		userRepository,
		passwordService,
		jwtService,
	)

	notificationService := services.NewNotificationService(
		fcmTokenRepository,
	)

	firebaseMessagingService := services.NewFirebaseMessagingService(
		firebaseApp.MessagingClient,
		fcmTokenRepository,
	)

	notificationQueue := queue.NewNotificationQueue(
		firebaseMessagingService,
		100,
		5,
	)

	notificationQueue.Start(ctx)
	defer notificationQueue.Stop()

	reportService := services.NewReportService(reportRepository, alertRepository)
	assistanceService := services.NewAssistanceService(assistanceRepository)
	sosService := services.NewSOSService(sosRepository)
	weatherService := services.NewWeatherService(
		cfg.OpenWeatherAPIKey,
		isEnabled(cfg.OpenWeatherEnabled),
		alertRepository,
	)
	
	alertService := services.NewAlertService(
		alertRepository,
		notificationQueue,
	)

	emergencyMessageService := services.NewEmergencyMessageService(
		emergencyMessageRepository,
		emergencyContactRepository,
	)

	alertPreferenceService := services.NewAlertPreferenceService(
		alertPreferenceRepository,
	)
	

	authHandler := handlers.NewAuthHandler(authService)

	notificationHandler := handlers.NewNotificationHandler(
		notificationService,
		firebaseMessagingService,
		notificationQueue,
	)

	emergencyContactService := services.NewEmergencyContactService(
		emergencyContactRepository,
	)

	userLocationService := services.NewUserLocationService(
		userRepository,
	)

	userProfileDetailsService := services.NewUserProfileDetailsService(
		userRepository,
	)

	chatService := services.NewChatService(
		conversationRepo,
		chatMessageRepo,
	)

	reportHandler := handlers.NewReportHandler(reportService, cloudinaryService)
	assistanceHandler := handlers.NewAssistanceHandler(assistanceService)
	sosHandler := handlers.NewSOSHandler(sosService)
	alertHandler := handlers.NewAlertHandler(alertService)
	aggregatorHandler := handlers.NewAggregatorHandler(alertAggregator)
	weatherHandler := handlers.NewWeatherHandler(weatherService)
	cleanupHandler := handlers.NewCleanupHandler(alertRepository)
	externalSourceHandler := handlers.NewExternalSourceHandler(cfg)
	emergencyContactHandler := handlers.NewEmergencyContactHandler(
		emergencyContactService,
	)
	emergencyMessageHandler := handlers.NewEmergencyMessageHandler(
		emergencyMessageService,
	)
	alertPreferenceHandler := handlers.NewAlertPreferenceHandler(
		alertPreferenceService,
	)

	userLocationHandler := handlers.NewUserLocationHandler(
		userLocationService,
	)

	userProfileDetailsHandler := handlers.NewUserProfileDetailsHandler(
		userProfileDetailsService,
	)

	chatHandler := handlers.NewChatHandler(chatService)



	router := app.SetupRouter(
		authHandler,
		notificationHandler,
		reportHandler,
		assistanceHandler,
		sosHandler,
		alertHandler,
		aggregatorHandler,
		weatherHandler,
		cleanupHandler,
		externalSourceHandler,
		emergencyContactHandler,
		emergencyMessageHandler,
		alertPreferenceHandler,
		userLocationHandler,
		userProfileDetailsHandler,
		chatHandler,
		jwtService,
	)


	port := os.Getenv("PORT")
	if port == "" {
		port = cfg.Port
	}

	if port == "" {
		port = "8080"
	}

	serverAddress := "0.0.0.0:" + port

	log.Println("Server running on", serverAddress)

	if err := router.Run(serverAddress); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func isEnabled(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))

	return value == "true" ||
		value == "1" ||
		value == "yes" ||
		value == "enabled"
}