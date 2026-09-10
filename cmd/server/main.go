package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/getsentry/sentry-go"

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
	"disaster_alert_backend/internal/websocket"
)

// @title Disaster Alert API
// @version 1.0.0
// @description Disaster Alert API is a mission-control backend for disaster intelligence, community incident reports, SOS escalation, assistance coordination, emergency messaging, public alerts, chats, notifications, and administrator response operations.
// @description
// @description ================================
// @description QUICK START
// @description ================================
// @description
// @description Use this Swagger page to test the API safely.
// @description
// @description For mobile/user endpoints:
// @description 1. Register or login using /auth/register or /auth/login.
// @description 2. Copy the token from the login response.
// @description 3. Click Authorize.
// @description 4. Paste it like this: Bearer YOUR_JWT_TOKEN.
// @description 5. Test endpoints such as /reports, /alerts, /sos, /assistance, /notifications, and /chats.
// @description
// @description For admin endpoints:
// @description 1. Enter the Admin API Key under AdminApiKeyAuth.
// @description 2. Create or validate a privilege code.
// @description 3. Enter the full privilege code under PrivilegeCodeAuth.
// @description 4. Test admin endpoints such as /admin/reports, /admin/alerts, /admin/sos, /admin/assistance, and /admin/chats.
// @description
// @description ================================
// @description AUTHENTICATION GUIDE
// @description ================================
// @description
// @description 1. Mobile/User endpoints use BearerAuth.
// @description Required header:
// @description Authorization: Bearer <JWT_TOKEN>
// @description
// @description 2. Basic admin management endpoints use AdminApiKeyAuth.
// @description Required header:
// @description Sigtrack-Admin-API-Key: <ADMIN_API_KEY>
// @description
// @description 3. Privileged admin operation endpoints require BOTH AdminApiKeyAuth and PrivilegeCodeAuth.
// @description Required headers:
// @description Sigtrack-Admin-API-Key: <ADMIN_API_KEY>
// @description X-Privilege-Code: <GENERATED_UUID>
// @description
// @description ================================
// @description ADMIN PRIVILEGE CODE FLOW
// @description ================================
// @description
// @description Step 1: Click Authorize in Swagger.
// @description Step 2: Enter your Admin API Key under AdminApiKeyAuth.
// @description Step 3: Call POST /admin/privilege-codes to generate a privilege UUID.
// @description Step 4: Copy the full UUID from data.code in the response.
// @description Step 5: Click Authorize again and paste the UUID under PrivilegeCodeAuth.
// @description Step 6: Test protected admin endpoints.
// @description
// @description Important:
// @description The full UUID is returned only once during creation.
// @description codePrefix is only for display and audit logs. Do not use codePrefix as X-Privilege-Code.
// @description
// @description ================================
// @description COMMON REQUEST NOTES
// @description ================================
// @description
// @description POST and PUT endpoints usually require a request body.
// @description Open the endpoint in Swagger and check the Parameters section.
// @description For JSON endpoints, send Content-Type: application/json.
// @description For upload endpoints, send Content-Type: multipart/form-data.
// @description
// @description Required fields are marked as required in the request schema or form fields.
// @description Optional fields can be omitted unless your frontend needs them.
// @description Date/time fields should use ISO format where possible, for example: 2026-09-10T08:30:00Z.
// @description
// @description ================================
// @description ADMIN PERMISSION CATALOG
// @description ================================
// @description
// @description Reports:
// @description - reports:read
// @description - reports:approve
// @description
// @description Assistance:
// @description - assistance:read
// @description - assistance:update_status
// @description
// @description SOS:
// @description - sos:read
// @description - sos:update_status
// @description
// @description Alerts:
// @description - alerts:read
// @description - alerts:create
// @description - alerts:update
// @description - alerts:delete
// @description
// @description Chats:
// @description - chats:read
// @description - chats:send
// @description
// @description Notifications:
// @description - notifications:read
// @description - notifications:send
// @description
// @description Privilege Management:
// @description - privilege_codes:create
// @description - privilege_codes:read
// @description - privilege_codes:revoke
// @description - audit_logs:read
// @BasePath /api/v1
// @schemes http https

// @securityDefinitions.apikey AdminApiKeyAuth
// @in header
// @name Sigtrack-Admin-API-Key

// @securityDefinitions.apikey PrivilegeCodeAuth
// @in header
// @name X-Privilege-Code

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func main() {
	cfg := config.LoadConfig()

	sentryDSN := cfg.SentryDSN

	if sentryDSN != "" {

		appEnv := strings.TrimSpace(cfg.AppEnv)
		if appEnv == "" {
			appEnv = "development"
		}

		if err := sentry.Init(sentry.ClientOptions{
			Dsn:                  cfg.SentryDSN,
			Environment:          cfg.AppEnv,
			Debug:                cfg.SentryDebug,
			SendDefaultPII:       false,
			EnableTracing:        true,
			TracesSampleRate:     cfg.SentryTracesSampleRate,
			DisableLogs:          false,
			DisableClientReports: true,

			TraceIgnoreStatusCodes: [][]int{
				{401},
				{403},
				{404},
			},

			BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
				if event.Request != nil && event.Request.Headers != nil {
					delete(event.Request.Headers, "Sigtrack-Admin-API-Key")
					delete(event.Request.Headers, "X-Privilege-Code")
					delete(event.Request.Headers, "Authorization")
				}

				return event
			},
		}); err != nil {
			log.Printf("Sentry initialization failed: %v", err)
		} else {
			log.Println("Sentry initialized successfully")
		}

		defer sentry.Flush(2 * time.Second)
	} else {
		log.Println("SENTRY_DSN not set. Sentry disabled.")
	}
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	db, err := database.ConnectMongoDB(cfg)
	if err != nil {
		sentry.CaptureException(err)
		sentry.Flush(2 * time.Second)
		log.Fatal("Failed to connect MongoDB:", err)
	}
	defer db.Disconnect()

	firebaseApp, err := config.InitFirebase(cfg)
	if err != nil {
		sentry.CaptureException(err)
		sentry.Flush(2 * time.Second)
		log.Fatal("Failed to initialize Firebase:", err)
	}

	cloudinaryService, err := services.NewCloudinaryService(cfg)
	if err != nil {
		sentry.CaptureException(err)
		sentry.Flush(2 * time.Second)
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

	adminPrivilegeCodeRepository := repositories.NewAdminPrivilegeCodeRepository(db.Database)
	adminPrivilegeLogRepository := repositories.NewAdminPrivilegeLogRepository(db.Database)

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

	if err := adminPrivilegeCodeRepository.EnsureIndexes(ctx); err != nil {
		log.Println("Failed to ensure admin privilege code indexes:", err)
	} else {
		log.Println("Admin privilege code indexes ensured successfully")
	}

	if err := adminPrivilegeLogRepository.EnsureIndexes(ctx); err != nil {
		log.Println("Failed to ensure admin privilege log indexes:", err)
	} else {
		log.Println("Admin privilege log indexes ensured successfully")
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
	wsHub := websocket.NewHub()
	wsBroadcaster := websocket.NewBroadcaster(wsHub)

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
		wsBroadcaster,
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
		userRepository,
		eventNotificationService,
		wsBroadcaster,
	)

	assistanceService := services.NewAssistanceService(
		assistanceRepository,
		userRepository,
		eventNotificationService,
		wsBroadcaster,
	)

	sosService := services.NewSOSService(
		sosRepository,
		userRepository,
		eventNotificationService,
		wsBroadcaster,
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
		wsBroadcaster,
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

	webSocketHandler := websocket.NewHandler(
		wsHub,
		userRepository,
	)

	userProfileDetailsService := services.NewUserProfileDetailsService(
		userRepository,
	)

	chatService := services.NewChatService(
		conversationRepo,
		chatMessageRepo,
		userRepository,
		eventNotificationService,
		wsBroadcaster,
	)

	adminPrivilegeCodeService := services.NewAdminPrivilegeCodeService(
		adminPrivilegeCodeRepository,
		adminPrivilegeLogRepository,
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

	adminPrivilegeCodeHandler := handlers.NewAdminPrivilegeCodeHandler(
		adminPrivilegeCodeService,
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
		adminPrivilegeCodeHandler,
		adminPrivilegeCodeService,
		webSocketHandler,
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
		sentry.CaptureException(err)
		sentry.Flush(2 * time.Second)
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
