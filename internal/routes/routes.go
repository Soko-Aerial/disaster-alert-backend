package routes

import (
	"net/http"

	"disaster_alert_backend/internal/handlers"
	"disaster_alert_backend/internal/middleware"
	"disaster_alert_backend/internal/services"
	"disaster_alert_backend/internal/utils"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(
	router *gin.Engine,
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
	jwtService *services.JWTService,
	adminAPIKey string,
) {

	router.GET("/health", func(c *gin.Context) {
		utils.SuccessResponse(
			c,
			http.StatusOK,
			"Disaster Alert API is running",
			gin.H{
				"status": "healthy",
			},
		)
	})
	
	api := router.Group("/api/v1")

	// Health check
	api.GET("/health", func(c *gin.Context) {
		utils.SuccessResponse(
			c,
			http.StatusOK,
			"Disaster Alert API is running",
			gin.H{
				"status": "healthy",
			},
		)
	})

	// -------------------------
	// Public Auth Routes
	// -------------------------
	auth := api.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.GET("/me", middleware.AuthMiddleware(jwtService), authHandler.Me)
	}

	// -------------------------
	// Public / Semi-public Routes
	// -------------------------

	weather := api.Group("/weather")
	{
		weather.GET("/alerts", weatherHandler.GetWeatherAlerts)
		weather.GET("/current", weatherHandler.GetCurrentWeather)
		weather.GET("/forecast", weatherHandler.GetWeatherForecast)
	}

	news := api.Group("/news")
	{
		news.GET("", newsHandler.GetNews)
	}


	externalSources := api.Group("/external-sources")
	{
		externalSources.GET("/status", externalSourceHandler.GetExternalSourceStatus)
	}

	// -------------------------
	// ADMIN API ROUTE
	// -------------------------

	admin := api.Group("/admin")
		admin.Use(middleware.AdminAPIKeyMiddleware(adminAPIKey))
		{
			adminReports := admin.Group("/reports")
			{
				adminReports.GET("", reportHandler.GetReports)
				adminReports.GET("/:id", reportHandler.GetReportByID)
				adminReports.PUT("/:id/approve", reportHandler.ApproveReport)
			}

			adminAssistance := admin.Group("/assistance")
			{
				adminAssistance.GET("", assistanceHandler.GetAssistanceRequests)
				adminAssistance.GET("/:id", assistanceHandler.GetAssistanceRequestByID)
			}

			adminSOS := admin.Group("/sos")
			{
				adminSOS.GET("", sosHandler.GetSOSRequests)
				adminSOS.GET("/:id", sosHandler.GetSOSByID)
				adminSOS.PUT("/:id/status", sosHandler.UpdateSOSStatus)
			}

			adminAlerts := admin.Group("/alerts")
			{
				adminAlerts.POST("", alertHandler.CreateAlert)
				adminAlerts.GET("", alertHandler.GetAlerts)

				adminAlerts.GET("/active", alertHandler.GetActiveAlerts)
				adminAlerts.POST("/sync-external", aggregatorHandler.SyncExternalAlerts)
				adminAlerts.GET("/local", alertHandler.GetLocalAlerts)
				adminAlerts.GET("/global", alertHandler.GetGlobalAlerts)
				adminAlerts.GET("/weather", alertHandler.GetWeatherAlerts)
				adminAlerts.GET("/health", alertHandler.GetHealthAlerts)

				adminAlerts.GET("/:id", alertHandler.GetAlertByID)
				adminAlerts.PUT("/:id/status", alertHandler.UpdateAlertStatus)
				adminAlerts.DELETE("/:id", alertHandler.DeleteAlert)
			}

			adminNotifications := admin.Group("/notifications")
			{
				adminNotifications.POST("/test/all", notificationHandler.SendTestToAll)
			}

			adminMaintenance := admin.Group("/maintenance")
			{
				adminMaintenance.POST("/cleanup-expired-alerts", cleanupHandler.CleanupExpiredExternalAlerts)
			}

			adminChats := admin.Group("/chats")
			{
				adminChats.GET("/conversations", chatHandler.GetConversations)
				adminChats.GET("/conversations/:id/messages", chatHandler.GetConversationMessages)
				adminChats.POST("/conversations/:id/messages", chatHandler.SendMessage)
				adminChats.PUT("/conversations/:id/read", chatHandler.MarkConversationRead)
			}
		}





	// -------------------------
	// Protected Routes
	// Everything inside this group requires JWT
	// -------------------------
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware(jwtService))
	{
		notifications := protected.Group("/notifications")
		{
			notifications.POST("/token", notificationHandler.SaveFCMToken)
			notifications.POST("/test/me", notificationHandler.SendTestToMe)
			notifications.POST("/test/all", notificationHandler.SendTestToAll)

			notifications.GET("", appNotificationHandler.GetMyNotifications)
			notifications.GET("/unread-count", appNotificationHandler.GetUnreadCount)
			notifications.PUT("/read", appNotificationHandler.MarkRead)
			notifications.PUT("/read-all", appNotificationHandler.MarkAllRead)
		}

		reports := protected.Group("/reports")
		{
			reports.POST("", reportHandler.CreateReport)
			reports.GET("", reportHandler.GetReports)
			reports.GET("/:id", reportHandler.GetReportByID)
			reports.PUT("/:id/approve", reportHandler.ApproveReport)
		}

		assistance := protected.Group("/assistance")
		{
			assistance.POST("", assistanceHandler.CreateAssistanceRequest)
			assistance.GET("", assistanceHandler.GetAssistanceRequests)
			assistance.GET("/:id", assistanceHandler.GetAssistanceRequestByID)
		}

		sos := protected.Group("/sos")
		{
			sos.POST("", sosHandler.CreateSOSRequest)
			sos.GET("", sosHandler.GetSOSRequests)
			sos.GET("/:id", sosHandler.GetSOSByID)
			sos.PUT("/:id/status", sosHandler.UpdateSOSStatus)
		}

		alerts := protected.Group("/alerts")
		{
			alerts.POST("", alertHandler.CreateAlert)
			alerts.GET("", alertHandler.GetAlerts)

			// Important: these specific routes must come before /:id
			alerts.GET("/active", alertHandler.GetActiveAlerts)
			// alerts.GET("/nearby", alertHandler.GetNearbyAlerts)
			// alerts.GET("/critical-global", alertHandler.GetCriticalGlobalAlerts)
			alerts.POST("/sync-external", aggregatorHandler.SyncExternalAlerts)
			alerts.GET("/local", alertHandler.GetLocalAlerts)
			alerts.GET("/global", alertHandler.GetGlobalAlerts)
			alerts.GET("/weather", alertHandler.GetWeatherAlerts)
			alerts.GET("/health", alertHandler.GetHealthAlerts)

			alerts.GET("/:id", alertHandler.GetAlertByID)
			alerts.PUT("/:id/status", alertHandler.UpdateAlertStatus)
			alerts.DELETE("/:id", alertHandler.DeleteAlert)
		}

		maintenance := protected.Group("/maintenance")
		{
			maintenance.POST("/cleanup-expired-alerts", cleanupHandler.CleanupExpiredExternalAlerts)
		}

		emergencyContacts := protected.Group("/emergency-contacts")
		{
			emergencyContacts.POST("", emergencyContactHandler.CreateContact)
			emergencyContacts.GET("", emergencyContactHandler.GetContacts)
			emergencyContacts.GET("/:id", emergencyContactHandler.GetContactByID)
			emergencyContacts.PUT("/:id", emergencyContactHandler.UpdateContact)
			emergencyContacts.DELETE("/:id", emergencyContactHandler.DeleteContact)
		}

		emergencyMessages := protected.Group("/emergency-messages")
		{
			emergencyMessages.POST("", emergencyMessageHandler.CreateMessage)
			emergencyMessages.GET("", emergencyMessageHandler.GetMessages)
			emergencyMessages.POST("/send", emergencyMessageHandler.SendMessage)

			emergencyMessages.GET("/:id", emergencyMessageHandler.GetMessageByID)
			emergencyMessages.DELETE("/:id", emergencyMessageHandler.DeleteMessage)

		}

		user := protected.Group("/user")
		{
			user.GET("/alert-preferences", alertPreferenceHandler.GetPreferences)
			user.PUT("/alert-preferences", alertPreferenceHandler.UpdatePreferences)

			user.GET("/location", userLocationHandler.GetLocation)
			user.PUT("/location", userLocationHandler.UpdateLocation)

			user.GET("/profile-details", userProfileDetailsHandler.GetProfileDetails)
			user.PUT("/profile-details", userProfileDetailsHandler.UpdateProfileDetails)
		}

		chats := protected.Group("/chats")
		{
			chats.POST("/conversations", chatHandler.CreateConversation)
			chats.GET("/conversations", chatHandler.GetConversations)

			chats.GET("/conversations/:id/messages", chatHandler.GetConversationMessages)
			chats.POST("/conversations/:id/messages", chatHandler.SendMessage)
			
			chats.PUT("/conversations/:id/read", chatHandler.MarkConversationRead)
		}
	}
}