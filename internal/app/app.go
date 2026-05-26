package app

import (
	"disaster_alert_backend/internal/routes"
	"disaster_alert_backend/internal/handlers"
	"disaster_alert_backend/internal/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter(
	authHandler *handlers.AuthHandler, 
	notificationHandler *handlers.NotificationHandler,
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
	jwtService *services.JWTService) *gin.Engine {
	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	routes.RegisterRoutes(
		router, 
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

	return router
}