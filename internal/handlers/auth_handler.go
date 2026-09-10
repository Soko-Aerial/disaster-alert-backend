package handlers

import (
	"net/http"

	"disaster_alert_backend/internal/dto"
	"disaster_alert_backend/internal/services"
	"disaster_alert_backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type AuthHandler struct {
	authService *services.AuthService
	validator   *validator.Validate
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		validator:   validator.New(),
	}
}


// Register godoc
// @Summary Register a new user
// @Description Creates a new mobile/user account.
// @Description
// @Description WHEN TO USE THIS ENDPOINT:
// @Description Use this endpoint when a new user wants to create an account in the mobile app.
// @Description
// @Description REQUIRED FIELDS:
// @Description - name: User's full name. Minimum 2 characters.
// @Description - email: User's email address. Must be a valid email.
// @Description - password: User's password. Minimum 6 characters.
// @Description
// @Description OPTIONAL FIELD:
// @Description - phone: User's phone number.
// @Description
// @Description EXAMPLE REQUEST BODY:
// @Description {
// @Description   "name": "Kwame Mensah",
// @Description   "email": "kwame@example.com",
// @Description   "password": "StrongPass123",
// @Description   "phone": "+233241234567"
// @Description }
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "User registration payload. name, email, and password are required."
// @Success 201 {object} map[string]interface{} "User registered successfully. Response includes token and user."
// @Failure 400 {object} map[string]interface{} "Invalid request body, validation failed, or user already exists."
// @Failure 500 {object} map[string]interface{} "Server error while registering user."
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(
			c,
			http.StatusBadRequest,
			"Invalid request body",
			err.Error(),
		)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		utils.ErrorResponse(
			c,
			http.StatusBadRequest,
			"Validation failed",
			err.Error(),
		)
		return
	}

	response, err := h.authService.Register(req)
	if err != nil {
		utils.ErrorResponse(
			c,
			http.StatusBadRequest,
			err.Error(),
			nil,
		)
		return
	}

	utils.SuccessResponse(
		c,
		http.StatusCreated,
		"User registered successfully",
		response,
	)
}


// Login godoc
// @Summary Login user
// @Description Authenticates a registered user and returns a JWT token.
// @Description
// @Description WHEN TO USE THIS ENDPOINT:
// @Description Use this endpoint when a user wants to sign in to the mobile app.
// @Description
// @Description REQUIRED FIELDS:
// @Description - email: Registered user email.
// @Description - password: Account password.
// @Description
// @Description HOW TO USE THE TOKEN:
// @Description 1. Copy the token from the response.
// @Description 2. Click Authorize in Swagger.
// @Description 3. Paste it under BearerAuth like this: Bearer YOUR_JWT_TOKEN.
// @Description 4. Use protected endpoints like /reports, /alerts, /sos, /assistance, /notifications, and /chats.
// @Description
// @Description EXAMPLE REQUEST BODY:
// @Description {
// @Description   "email": "kwame@example.com",
// @Description   "password": "StrongPass123"
// @Description }
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "User login payload. email and password are required."
// @Success 200 {object} map[string]interface{} "Login successful. Response includes token and user."
// @Failure 400 {object} map[string]interface{} "Invalid request body or validation failed."
// @Failure 401 {object} map[string]interface{} "Invalid email or password."
// @Failure 500 {object} map[string]interface{} "Server error while logging in."
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(
			c,
			http.StatusBadRequest,
			"Invalid request body",
			err.Error(),
		)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		utils.ErrorResponse(
			c,
			http.StatusBadRequest,
			"Validation failed",
			err.Error(),
		)
		return
	}

	response, err := h.authService.Login(req)
	if err != nil {
		utils.ErrorResponse(
			c,
			http.StatusUnauthorized,
			err.Error(),
			nil,
		)
		return
	}

	utils.SuccessResponse(
		c,
		http.StatusOK,
		"Login successful",
		response,
	)
}

// Me godoc
// @Summary Get current user
// @Description Returns the profile of the currently authenticated user.
// @Description
// @Description AUTH REQUIRED:
// @Description This endpoint requires a valid user JWT token.
// @Description Click Authorize and paste: Bearer YOUR_JWT_TOKEN.
// @Description
// @Description WHEN TO USE THIS ENDPOINT:
// @Description Use this endpoint after login to confirm the token is valid and fetch the logged-in user's details.
// @Tags Auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{} "Current user fetched successfully."
// @Failure 401 {object} map[string]interface{} "Missing, invalid, or expired user token."
// @Failure 404 {object} map[string]interface{} "User not found."
// @Failure 500 {object} map[string]interface{} "Server error while fetching current user."
// @Router /auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	userIDValue, exists := c.Get("userId")

	if !exists {
		utils.ErrorResponse(
			c,
			http.StatusUnauthorized,
			"User not found in request context",
			nil,
		)
		return
	}

	userID, ok := userIDValue.(string)

	if !ok {
		utils.ErrorResponse(
			c,
			http.StatusUnauthorized,
			"Invalid user id in request context",
			nil,
		)
		return
	}

	user, err := h.authService.GetCurrentUser(userID)

	if err != nil {
		utils.ErrorResponse(
			c,
			http.StatusNotFound,
			err.Error(),
			nil,
		)
		return
	}

	utils.SuccessResponse(
		c,
		http.StatusOK,
		"Current user fetched successfully",
		user,
	)
}