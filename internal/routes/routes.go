package routes

import (
	"net/http"

	docs "disaster_alert_backend/docs"

	"disaster_alert_backend/internal/handlers"
	"disaster_alert_backend/internal/middleware"
	"disaster_alert_backend/internal/permissions"
	"disaster_alert_backend/internal/services"
	"disaster_alert_backend/internal/utils"
	"disaster_alert_backend/internal/websocket"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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
	adminPrivilegeCodeHandler *handlers.AdminPrivilegeCodeHandler,
	adminPrivilegeCodeService *services.AdminPrivilegeCodeService,
	webSocketHandler *websocket.Handler,
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

	docs.SwaggerInfo.Title = "Disaster Alert API"
	docs.SwaggerInfo.Description = `Mission-control API for real-time disaster intelligence, SOS escalation, assistance coordination, incident reports, alerts, chats, notifications, and administrator response operations.

	AUTHENTICATION GUIDE

	1. Mobile/User endpoints use BearerAuth.
	Header:
	Authorization: Bearer <JWT_TOKEN>

	2. Basic admin management endpoints use AdminApiKeyAuth only.
	Header:
	Sigtrack-Admin-API-Key: <ADMIN_API_KEY>

	3. Privileged admin operation endpoints require BOTH AdminApiKeyAuth and PrivilegeCodeAuth.
	Headers:
	Sigtrack-Admin-API-Key: <ADMIN_API_KEY>
	X-Privilege-Code: <GENERATED_UUID>

	ADMIN PRIVILEGE CODE FLOW

	Step 1: Click Authorize.
	Step 2: Enter your Admin API Key under AdminApiKeyAuth.
	Step 3: Call POST /admin/privilege-codes.
	Step 4: Copy the full UUID from data.code.
	Step 5: Click Authorize again and paste the UUID under PrivilegeCodeAuth.
	Step 6: Test protected admin endpoints.

	IMPORTANT:
	The full UUID is returned only once during creation.
	codePrefix is only for display and logs. Do not use codePrefix as X-Privilege-Code.`
	docs.SwaggerInfo.Version = "1.0.0"
	docs.SwaggerInfo.Host = "localhost:8080"
	docs.SwaggerInfo.BasePath = "/api/v1"
	docs.SwaggerInfo.Schemes = []string{"http"}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.GET("/docs", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/swagger/index.html")
	})

	api := router.Group("/api/v1")
	api.GET("/ws", middleware.WebSocketAuthMiddleware(jwtService), webSocketHandler.Connect)

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
		admin.GET("/ws", webSocketHandler.Connect)

		adminPrivilegeCodes := admin.Group("/privilege-codes")
		{
			adminPrivilegeCodes.POST("", adminPrivilegeCodeHandler.CreatePrivilegeCode)
			adminPrivilegeCodes.GET("", adminPrivilegeCodeHandler.GetPrivilegeCodes)
			adminPrivilegeCodes.GET("/:id", adminPrivilegeCodeHandler.GetPrivilegeCodeByID)
			adminPrivilegeCodes.POST("/validate", adminPrivilegeCodeHandler.ValidatePrivilegeCode)
			adminPrivilegeCodes.PUT("/:id/revoke", adminPrivilegeCodeHandler.RevokePrivilegeCode)
		}

		admin.GET("/privilege-logs", adminPrivilegeCodeHandler.GetPrivilegeLogs)

		adminReports := admin.Group("/reports")
		{
			adminReports.GET("", middleware.RequirePrivilegePermission(adminPrivilegeCodeService, permissions.ReportsRead), reportHandler.GetReports)
			adminReports.GET("/:id", middleware.RequirePrivilegePermission(adminPrivilegeCodeService, permissions.ReportsRead), reportHandler.GetReportByID)
			adminReports.PUT("/:id/approve", middleware.RequirePrivilegePermission(adminPrivilegeCodeService, permissions.ReportsApprove), reportHandler.ApproveReport)
		}

		adminAssistance := admin.Group("/assistance")
		{
			adminAssistance.GET("", middleware.RequirePrivilegePermission(adminPrivilegeCodeService, permissions.AssistanceRead), assistanceHandler.GetAssistanceRequests)
			adminAssistance.GET("/:id", middleware.RequirePrivilegePermission(adminPrivilegeCodeService, permissions.AssistanceRead), assistanceHandler.GetAssistanceRequestByID)
			adminAssistance.PUT("/:id/status", middleware.RequirePrivilegePermission(adminPrivilegeCodeService, permissions.AssistanceUpdateStatus), assistanceHandler.UpdateAssistanceStatus)
		}

		adminSOS := admin.Group("/sos")
		{
			adminSOS.GET("", middleware.RequirePrivilegePermission(adminPrivilegeCodeService, permissions.SOSRead), sosHandler.GetSOSRequests)
			adminSOS.GET("/:id", middleware.RequirePrivilegePermission(adminPrivilegeCodeService, permissions.SOSRead), sosHandler.GetSOSByID)
			adminSOS.PUT("/:id/status", middleware.RequirePrivilegePermission(adminPrivilegeCodeService, permissions.SOSUpdateStatus), sosHandler.UpdateSOSStatus)
		}

		adminAlerts := admin.Group("/alerts")
		{
			adminAlerts.POST("", middleware.RequirePrivilegePermission(adminPrivilegeCodeService, permissions.AlertsCreate), alertHandler.CreateAlert)
			adminAlerts.GET("", middleware.RequirePrivilegePermission(adminPrivilegeCodeService, permissions.AlertsRead), alertHandler.GetAlerts)

			adminAlerts.GET("/active", middleware.RequirePrivilegePermission(adminPrivilegeCodeService, permissions.AlertsRead), alertHandler.GetActiveAlerts)
			adminAlerts.POST("/sync-external", middleware.RequirePrivilegePermission(adminPrivilegeCodeService, permissions.AlertsCreate), aggregatorHandler.SyncExternalAlerts)
			adminAlerts.GET("/local", middleware.RequirePrivilegePermission(adminPrivilegeCodeService, permissions.AlertsRead), alertHandler.GetLocalAlerts)
			adminAlerts.GET("/global", middleware.RequirePrivilegePermission(adminPrivilegeCodeService, permissions.AlertsRead), alertHandler.GetGlobalAlerts)
			adminAlerts.GET("/weather", middleware.RequirePrivilegePermission(adminPrivilegeCodeService, permissions.AlertsRead), alertHandler.GetWeatherAlerts)
			adminAlerts.GET("/health", middleware.RequirePrivilegePermission(adminPrivilegeCodeService, permissions.AlertsRead), alertHandler.GetHealthAlerts)

			adminAlerts.GET("/:id", middleware.RequirePrivilegePermission(adminPrivilegeCodeService, permissions.AlertsRead), alertHandler.GetAlertByID)
			adminAlerts.PUT("/:id/status", middleware.RequirePrivilegePermission(adminPrivilegeCodeService, permissions.AlertsUpdate), alertHandler.UpdateAlertStatus)
			adminAlerts.DELETE("/:id", middleware.RequirePrivilegePermission(adminPrivilegeCodeService, permissions.AlertsDelete), alertHandler.DeleteAlert)
		}

		adminNotifications := admin.Group("/notifications")
		{
			adminNotifications.POST("/test/all", middleware.RequirePrivilegePermission(adminPrivilegeCodeService, permissions.NotificationsSend), notificationHandler.SendTestToAll)
		}

		adminMaintenance := admin.Group("/maintenance")
		{
			adminMaintenance.POST("/cleanup-expired-alerts", middleware.RequirePrivilegePermission(adminPrivilegeCodeService, permissions.AlertsUpdate), cleanupHandler.CleanupExpiredExternalAlerts)
		}

		adminChats := admin.Group("/chats")
		{
			adminChats.GET("/conversations", middleware.RequirePrivilegePermission(adminPrivilegeCodeService, permissions.ChatsRead), chatHandler.GetConversations)
			adminChats.GET("/conversations/:id/messages", middleware.RequirePrivilegePermission(adminPrivilegeCodeService, permissions.ChatsRead), chatHandler.GetConversationMessages)
			adminChats.POST("/conversations/:id/messages", middleware.RequirePrivilegePermission(adminPrivilegeCodeService, permissions.ChatsSend), chatHandler.SendMessage)
			adminChats.PUT("/conversations/:id/read", middleware.RequirePrivilegePermission(adminPrivilegeCodeService, permissions.ChatsRead), chatHandler.MarkConversationRead)
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
			alerts.GET("", alertHandler.GetAlerts)

			alerts.GET("/active", alertHandler.GetActiveAlerts)
			alerts.GET("/local", alertHandler.GetLocalAlerts)
			alerts.GET("/global", alertHandler.GetGlobalAlerts)
			alerts.GET("/weather", alertHandler.GetWeatherAlerts)
			alerts.GET("/health", alertHandler.GetHealthAlerts)

			alerts.GET("/:id", alertHandler.GetAlertByID)
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
