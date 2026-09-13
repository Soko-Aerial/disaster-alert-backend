package dto

type AlertGeoPointRequest struct {
	Latitude  float64 `json:"latitude" binding:"required" example:"5.6037"`
	Longitude float64 `json:"longitude" binding:"required" example:"-0.1870"`
}

type AlertTargetingRequest struct {
	// Mode controls who receives push notifications.
	// Supported values: radius, polygon, region, country, national.
	Mode string `json:"mode,omitempty" example:"radius"`

	// RadiusKm is used when mode=radius.
	// Users inside this radius receive danger-zone notification.
	RadiusKm float64 `json:"radiusKm,omitempty" example:"5"`

	// AwarenessRadiusKm is wider than RadiusKm.
	// Users between RadiusKm and AwarenessRadiusKm receive softer awareness notification.
	AwarenessRadiusKm float64 `json:"awarenessRadiusKm,omitempty" example:"10"`

	// Polygon is used when mode=polygon.
	// The admin dashboard sends the drawn map boundary as points.
	Polygon []AlertGeoPointRequest `json:"polygon,omitempty"`

	// Administrative targeting fields.
	Country  string `json:"country,omitempty" example:"Ghana"`
	Region   string `json:"region,omitempty" example:"Greater Accra"`
	District string `json:"district,omitempty" example:"Accra Metropolitan"`

	// LocationFreshnessHours prevents sending local alerts based on old user location.
	// Recommended default: 24 hours.
	LocationFreshnessHours int `json:"locationFreshnessHours,omitempty" example:"24"`

	// RespectUserPreferences checks whether the user enabled this category.
	RespectUserPreferences bool `json:"respectUserPreferences,omitempty" example:"true"`

	// CriticalOverridePreferences allows critical nearby alerts to be sent even if category preference is off.
	CriticalOverridePreferences bool `json:"criticalOverridePreferences,omitempty" example:"true"`

	// Required only for mode=national.
	ConfirmNationalAlert bool `json:"confirmNationalAlert,omitempty" example:"false"`

	// Required only for mode=national.
	NationalAlertReason string `json:"nationalAlertReason,omitempty" example:"Nationwide severe weather emergency"`
}

type AlertTargetingPreviewRequest struct {
	// Category is the alert category.
	// Examples: fire, flood, weather, robbery, security, medical, accident.
	Category string `json:"category" binding:"required" validate:"required" example:"flood"`

	// Severity controls preference override and warning level.
	// Allowed values: low, medium, high, critical.
	Severity string `json:"severity" binding:"required" validate:"required,oneof=low medium high critical" example:"high"`

	// Latitude is the main alert latitude.
	// Required for radius targeting.
	Latitude float64 `json:"latitude,omitempty" example:"5.6037"`

	// Longitude is the main alert longitude.
	// Required for radius targeting.
	Longitude float64 `json:"longitude,omitempty" example:"-0.1870"`

	// Targeting contains the map-based delivery rules.
	Targeting AlertTargetingRequest `json:"targeting" binding:"required" validate:"required"`
}

type AlertTargetingPreviewResponse struct {
	TargetingMode string `json:"targetingMode"`

	DangerRecipients    int `json:"dangerRecipients"`
	AwarenessRecipients int `json:"awarenessRecipients"`
	TotalPushRecipients int `json:"totalPushRecipients"`

	ExcludedNoLocation    int `json:"excludedNoLocation"`
	ExcludedOldLocation   int `json:"excludedOldLocation"`
	ExcludedPreferenceOff int `json:"excludedPreferenceOff"`
	ExcludedNoFCMToken    int `json:"excludedNoFcmToken"`
	ExcludedOutsideTarget int `json:"excludedOutsideTarget"`

	ScannedUsers int `json:"scannedUsers"`

	LocationFreshnessHours int  `json:"locationFreshnessHours"`
	RespectUserPreferences bool `json:"respectUserPreferences"`

	Warnings []string `json:"warnings,omitempty"`
}
