package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	AlertDeliveryTypeInitial    = "initial"
	AlertDeliveryTypeEscalation = "escalation"
	AlertDeliveryTypeUpdate     = "update"

	AlertDeliveryStatusPending = "pending"
	AlertDeliveryStatusSent    = "sent"
	AlertDeliveryStatusFailed  = "failed"
	AlertDeliveryStatusSkipped = "skipped"
)

type AlertDeliveryBatch struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id"`

	AlertID primitive.ObjectID `bson:"alertId" json:"alertId"`

	DeliveryType string `bson:"deliveryType" json:"deliveryType"`

	TargetingMode string `bson:"targetingMode" json:"targetingMode"`

	DangerRecipients    int `bson:"dangerRecipients" json:"dangerRecipients"`
	AwarenessRecipients int `bson:"awarenessRecipients" json:"awarenessRecipients"`
	TotalRecipients     int `bson:"totalRecipients" json:"totalRecipients"`

	SuccessCount int `bson:"successCount" json:"successCount"`
	FailureCount int `bson:"failureCount" json:"failureCount"`
	SkippedCount int `bson:"skippedCount" json:"skippedCount"`

	ExcludedNoLocation    int `bson:"excludedNoLocation,omitempty" json:"excludedNoLocation,omitempty"`
	ExcludedOldLocation   int `bson:"excludedOldLocation,omitempty" json:"excludedOldLocation,omitempty"`
	ExcludedPreferenceOff int `bson:"excludedPreferenceOff,omitempty" json:"excludedPreferenceOff,omitempty"`
	ExcludedNoFCMToken    int `bson:"excludedNoFcmToken,omitempty" json:"excludedNoFcmToken,omitempty"`

	SentByPrivilegeCodeID string `bson:"sentByPrivilegeCodeId,omitempty" json:"sentByPrivilegeCodeId,omitempty"`
	SentByOrganisationID  string `bson:"sentByOrganisationId,omitempty" json:"sentByOrganisationId,omitempty"`
	SentByOrganisationName string `bson:"sentByOrganisationName,omitempty" json:"sentByOrganisationName,omitempty"`

	StartedAt   time.Time  `bson:"startedAt" json:"startedAt"`
	CompletedAt *time.Time `bson:"completedAt,omitempty" json:"completedAt,omitempty"`

	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time `bson:"updatedAt" json:"updatedAt"`
}

type AlertRecipientDelivery struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id"`

	AlertID         primitive.ObjectID `bson:"alertId" json:"alertId"`
	DeliveryBatchID primitive.ObjectID `bson:"deliveryBatchId" json:"deliveryBatchId"`
	UserID          primitive.ObjectID `bson:"userId" json:"userId"`

	DeliveryType string `bson:"deliveryType" json:"deliveryType"`
	Zone          string `bson:"zone" json:"zone"`

	Status string `bson:"status" json:"status"`

	// DedupKey prevents duplicate spam.
	// Example: alertId:userId:initial:danger
	DedupKey string `bson:"dedupKey" json:"dedupKey"`

	DistanceKm float64 `bson:"distanceKm,omitempty" json:"distanceKm,omitempty"`

	FailureReason string `bson:"failureReason,omitempty" json:"failureReason,omitempty"`

	SentAt    *time.Time `bson:"sentAt,omitempty" json:"sentAt,omitempty"`
	CreatedAt time.Time  `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time  `bson:"updatedAt" json:"updatedAt"`
}