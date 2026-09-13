package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SOSLocation struct {
	Latitude  float64 `bson:"latitude" json:"latitude"`
	Longitude float64 `bson:"longitude" json:"longitude"`
	Address   string  `bson:"address,omitempty" json:"address,omitempty"`
	Accuracy  float64 `bson:"accuracy,omitempty" json:"accuracy,omitempty"`
	Country   string  `bson:"country,omitempty" json:"country,omitempty"`
	Region    string  `bson:"region,omitempty" json:"region,omitempty"`
}

type SOSRequest struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID             primitive.ObjectID `bson:"userId" json:"userId"`
	EmergencyType      string             `bson:"emergencyType" json:"emergencyType"`
	AccessCategoryID   string             `bson:"accessCategoryId,omitempty" json:"accessCategoryId,omitempty"`
	AccessCategorySlug string             `bson:"accessCategorySlug,omitempty" json:"accessCategorySlug,omitempty"`
	AccessCategoryName string             `bson:"accessCategoryName,omitempty" json:"accessCategoryName,omitempty"`

	OwnerOrganisationID string              `bson:"ownerOrganisationId,omitempty" json:"ownerOrganisationId,omitempty"`
	LeadOrganisationID  string              `bson:"leadOrganisationId,omitempty" json:"leadOrganisationId,omitempty"`
	AssignedOrgIDs      []string            `bson:"assignedOrgIds,omitempty" json:"assignedOrgIds,omitempty"`
	VisibleToOrgIDs     []string            `bson:"visibleToOrgIds,omitempty" json:"visibleToOrgIds,omitempty"`
	Message             string              `bson:"message,omitempty" json:"message,omitempty"`
	Location            SOSLocation         `bson:"location" json:"location"`
	Status              string              `bson:"status" json:"status"`
	IsLiveTracking      bool                `bson:"isLiveTracking" json:"isLiveTracking"`
	ResponderID         *primitive.ObjectID `bson:"responderId,omitempty" json:"responderId,omitempty"`
	ResolvedAt          *time.Time          `bson:"resolvedAt,omitempty" json:"resolvedAt,omitempty"`
	CreatedAt           time.Time           `bson:"createdAt" json:"createdAt"`
	UpdatedAt           time.Time           `bson:"updatedAt" json:"updatedAt"`
}
