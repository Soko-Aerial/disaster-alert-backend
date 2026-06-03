package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"disaster_alert_backend/config"
	"disaster_alert_backend/internal/aggregator"
	"disaster_alert_backend/internal/app"
	"disaster_alert_backend/internal/database"
	"disaster_alert_backend/internal/handlers"
	"disaster_alert_backend/internal/queue"
	"disaster_alert_backend/internal/repositories"
	"disaster_alert_backend/internal/scheduler"
	"disaster_alert_backend/internal/services"
	"disaster_alert_backend/internal/sources"
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

	// Repositories
	userRepository := repositories.NewUserRepository(db.Database)
	fcmTokenRepository := repositories.NewFCMTokenRepository(db.Database)
	appNotificationRepository := repositories.NewAppNotificationRepository(db.Database)

	reportRepository := repositories.NewReportRepository(db.Database)
	assistanceRepository := repositories.NewAssistanceRepository(db.Database)
	sosRepository := repositories.NewSOSRepository(db.Database)
	alertRepository := repositories.NewAlertRepository(db.Database)

	emergencyContactRepository := repositories.NewEmergencyContactRepository(db.Database)
	emergencyMessageRepository := repositories.NewEmergencyMessageRepository(db.Database)
	alertPreferenceRepository := repositories.NewAlertPreferenceRepository(db.Database)

	conversationRepo := repositories.NewConversationRepository(db.Database)
	chatMessageRepo := repositories.NewChatMessageRepository(db.Database)

	// Indexes
	if err := alertRepository.EnsureIndexes(); err != nil {
		log.Println("Failed to ensure alert indexes:", err)
	} else {
		log.Println("Alert indexes ensured successfully")
	}

	if err := appNotificationRepository.EnsureIndexes(); err != nil {
		log.Println("Failed to ensure app notification indexes:", err)
	} else {
		log.Println("App notification indexes ensured successfully")
	}

	// External alert sources
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

	// Core services
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

	appNotificationService := services.NewAppNotificationService(
		appNotificationRepository,
	)

	eventNotificationService := services.NewEventNotificationService(
		userRepository,
		notificationQueue,
		appNotificationService,
	)

	// Feature services
	reportService := services.NewReportService(
		reportRepository,
		alertRepository,
		eventNotificationService,
	)

	assistanceService := services.NewAssistanceService(
		assistanceRepository,
		eventNotificationService,
	)

	sosService := services.NewSOSService(
		sosRepository,
		eventNotificationService,
	)

	weatherService := services.NewWeatherService(
		cfg.OpenWeatherAPIKey,
		isEnabled(cfg.OpenWeatherEnabled),
		alertRepository,
	)

	newsService := services.NewNewsService(
		cfg.GNewsAPIKey,
		isEnabled(cfg.GNewsEnabled),
	)

	alertService := services.NewAlertService(
		alertRepository,
		userRepository,
		notificationQueue,
		appNotificationService,
	)

	emergencyMessageService := services.NewEmergencyMessageService(
		emergencyMessageRepository,
		emergencyContactRepository,
	)

	alertPreferenceService := services.NewAlertPreferenceService(
		alertPreferenceRepository,
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
		eventNotificationService,
	)

	// Handlers
	authHandler := handlers.NewAuthHandler(authService)

	notificationHandler := handlers.NewNotificationHandler(
		notificationService,
		firebaseMessagingService,
		notificationQueue,
	)

	appNotificationHandler := handlers.NewAppNotificationHandler(
		appNotificationService,
	)

	reportHandler := handlers.NewReportHandler(
		reportService,
		cloudinaryService,
	)

	assistanceHandler := handlers.NewAssistanceHandler(
		assistanceService,
	)

	sosHandler := handlers.NewSOSHandler(
		sosService,
	)

	alertHandler := handlers.NewAlertHandler(
		alertService,
	)

	aggregatorHandler := handlers.NewAggregatorHandler(
		alertAggregator,
	)

	weatherHandler := handlers.NewWeatherHandler(
		weatherService,
	)

	newsHandler := handlers.NewNewsHandler(newsService)

	cleanupHandler := handlers.NewCleanupHandler(
		alertRepository,
	)

	externalSourceHandler := handlers.NewExternalSourceHandler(
		cfg,
	)

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

	chatHandler := handlers.NewChatHandler(
		chatService,
	)

	router := app.SetupRouter(
		authHandler,
		notificationHandler,
		appNotificationHandler,
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
		newsHandler,
		jwtService,
		cfg.AdminAPIKey,
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