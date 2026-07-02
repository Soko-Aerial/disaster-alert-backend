package app

import (
	"time"

	"disaster_alert_backend/internal/handlers"
	"disaster_alert_backend/internal/routes"
	"disaster_alert_backend/internal/services"
	"disaster_alert_backend/internal/websocket"

	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter(
	authHandler *handlers.AuthHandler,
	notificationHandler *handlers.NotificationHandler,
	appNotificationHandler *handlers.AppNotificationHandler,
	reportHandler *handlers.ReportHandler,
	assistanceHandler *handlers.AssistanceHandler,
	sosHandler *handlers.SOSHandler,
	alertHandler *handlers.AlertHandler,
	aggregatorHandler *handlers.AggregatorHandler,
	weatherHandler *handlers.WeatherHandler,
	cleanupHandler *handlers.CleanupHandler,
	externalSourceHandler *handlers.ExternalSourceHandler,
	emergencyContactHandler *handlers.EmergencyContactHandler,
	emergencyMessageHandler *handlers.EmergencyMessageHandler,
	alertPreferenceHandler *handlers.AlertPreferenceHandler,
	userLocationHandler *handlers.UserLocationHandler,
	userProfileDetailsHandler *handlers.UserProfileDetailsHandler,
	chatHandler *handlers.ChatHandler,
	newsHandler *handlers.NewsHandler,
	adminPrivilegeCodeHandler *handlers.AdminPrivilegeCodeHandler,
	adminPrivilegeCodeService *services.AdminPrivilegeCodeService,
	webSocketHandler *websocket.Handler,
	jwtService *services.JWTService,
	adminAPIKey string,
) *gin.Engine {
	router := gin.Default()

	router.Use(sentrygin.New(sentrygin.Options{
		Repanic:         true,
		WaitForDelivery: false,
		Timeout:         5 * time.Second,
	}))

	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{
			"GET",
			"POST",
			"PUT",
			"PATCH",
			"DELETE",
			"OPTIONS",
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Authorization",
			"Sigtrack-Admin-API-Key",
			"X-Privilege-Code",
			"X-Admin-Actor",
		},
		AllowCredentials: true,
	}))

	routes.RegisterRoutes(
		router,
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
		adminAPIKey,
	)

	return router
}
