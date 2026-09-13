package models

const (
	AlertTargetingModeRadius  = "radius"
	AlertTargetingModePolygon = "polygon"
	AlertTargetingModeRegion  = "region"
	AlertTargetingModeCountry = "country"
	AlertTargetingModeNational = "national"

	AlertTargetZoneDanger   = "danger"
	AlertTargetZoneAwareness = "awareness"

	AlertNotificationStatusDraft     = "draft"
	AlertNotificationStatusPreviewed = "previewed"
	AlertNotificationStatusSending   = "sending"
	AlertNotificationStatusSent      = "sent"
	AlertNotificationStatusPartial   = "partial"
	AlertNotificationStatusFailed    = "failed"
)

type AlertGeoPoint struct {
	Latitude  float64 `bson:"latitude" json:"latitude"`
	Longitude float64 `bson:"longitude" json:"longitude"`
}

type AlertTargeting struct {
	Mode string `bson:"mode" json:"mode"`

	// Radius targeting
	RadiusKm          float64 `bson:"radiusKm,omitempty" json:"radiusKm,omitempty"`
	AwarenessRadiusKm float64 `bson:"awarenessRadiusKm,omitempty" json:"awarenessRadiusKm,omitempty"`

	// Polygon targeting
	Polygon []AlertGeoPoint `bson:"polygon,omitempty" json:"polygon,omitempty"`

	// Administrative targeting
	Country  string `bson:"country,omitempty" json:"country,omitempty"`
	Region   string `bson:"region,omitempty" json:"region,omitempty"`
	District string `bson:"district,omitempty" json:"district,omitempty"`

	// Safety filters
	LocationFreshnessHours       int  `bson:"locationFreshnessHours" json:"locationFreshnessHours"`
	RespectUserPreferences       bool `bson:"respectUserPreferences" json:"respectUserPreferences"`
	CriticalOverridePreferences   bool `bson:"criticalOverridePreferences" json:"criticalOverridePreferences"`

	// National alert protection
	ConfirmNationalAlert bool   `bson:"confirmNationalAlert,omitempty" json:"confirmNationalAlert,omitempty"`
	NationalAlertReason  string `bson:"nationalAlertReason,omitempty" json:"nationalAlertReason,omitempty"`

	// Preview and delivery summary
	EstimatedDangerRecipients    int `bson:"estimatedDangerRecipients,omitempty" json:"estimatedDangerRecipients,omitempty"`
	EstimatedAwarenessRecipients int `bson:"estimatedAwarenessRecipients,omitempty" json:"estimatedAwarenessRecipients,omitempty"`
	EstimatedTotalRecipients     int `bson:"estimatedTotalRecipients,omitempty" json:"estimatedTotalRecipients,omitempty"`

	ActualDangerRecipients    int `bson:"actualDangerRecipients,omitempty" json:"actualDangerRecipients,omitempty"`
	ActualAwarenessRecipients int `bson:"actualAwarenessRecipients,omitempty" json:"actualAwarenessRecipients,omitempty"`
	ActualTotalRecipients     int `bson:"actualTotalRecipients,omitempty" json:"actualTotalRecipients,omitempty"`

	ExcludedNoLocation      int `bson:"excludedNoLocation,omitempty" json:"excludedNoLocation,omitempty"`
	ExcludedOldLocation     int `bson:"excludedOldLocation,omitempty" json:"excludedOldLocation,omitempty"`
	ExcludedPreferenceOff   int `bson:"excludedPreferenceOff,omitempty" json:"excludedPreferenceOff,omitempty"`
	ExcludedNoFCMToken      int `bson:"excludedNoFcmToken,omitempty" json:"excludedNoFcmToken,omitempty"`
	ExcludedOutsideTarget   int `bson:"excludedOutsideTarget,omitempty" json:"excludedOutsideTarget,omitempty"`

	NotificationStatus string `bson:"notificationStatus,omitempty" json:"notificationStatus,omitempty"`
}