package dto

// RegisterRequest is used to create a new mobile/user account.
//
// Required fields:
// - name
// - email
// - password
//
// Optional fields:
// - phone
type RegisterRequest struct {
	// Name is the user's full name.
	// Minimum length: 2 characters.
	Name string `json:"name" validate:"required,min=2" example:"Kwame Mensah"`

	// Email is the user's login email address.
	// It must be a valid email format.
	Email string `json:"email" validate:"required,email" example:"kwame@example.com"`

	// Password is the user's account password.
	// Minimum length: 6 characters.
	Password string `json:"password" validate:"required,min=6" example:"StrongPass123"`

	// Phone is the user's phone number.
	// This is optional.
	Phone string `json:"phone,omitempty" example:"+233241234567"`
}

// LoginRequest is used to authenticate an existing user.
//
// Required fields:
// - email
// - password
type LoginRequest struct {
	// Email is the user's registered email address.
	Email string `json:"email" validate:"required,email" example:"kwame@example.com"`

	// Password is the user's account password.
	Password string `json:"password" validate:"required" example:"StrongPass123"`
}

// AuthResponse is returned after a successful register or login request.
type AuthResponse struct {
	// Token is the JWT token used to access protected user endpoints.
	//
	// Use it in Swagger Authorize as:
	// Bearer YOUR_JWT_TOKEN
	Token string `json:"token" example:"Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`

	// User contains the authenticated user's profile information.
	User interface{} `json:"user"`
}