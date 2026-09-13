package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ReportLocation struct {
	Latitude  float64 `bson:"latitude" json:"latitude"`
	Longitude float64 `bson:"longitude" json:"longitude"`
	Address   string  `bson:"address,omitempty" json:"address,omitempty"`
	Country   string  `bson:"country,omitempty" json:"country,omitempty"`
	Region    string  `bson:"region,omitempty" json:"region,omitempty"`
}

type ReportMedia struct {
	URL      string `bson:"url" json:"url"`
	Type     string `bson:"type" json:"type"`
	PublicID string `bson:"publicId" json:"publicId"`
}

type Report struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID             primitive.ObjectID `bson:"userId" json:"userId"`
	Category           string             `bson:"category" json:"category"`
	AccessCategoryID   string             `bson:"accessCategoryId,omitempty" json:"accessCategoryId,omitempty"`
	AccessCategorySlug string             `bson:"accessCategorySlug,omitempty" json:"accessCategorySlug,omitempty"`
	AccessCategoryName string             `bson:"accessCategoryName,omitempty" json:"accessCategoryName,omitempty"`

	OwnerOrganisationID string         `bson:"ownerOrganisationId,omitempty" json:"ownerOrganisationId,omitempty"`
	LeadOrganisationID  string         `bson:"leadOrganisationId,omitempty" json:"leadOrganisationId,omitempty"`
	AssignedOrgIDs      []string       `bson:"assignedOrgIds,omitempty" json:"assignedOrgIds,omitempty"`
	VisibleToOrgIDs     []string       `bson:"visibleToOrgIds,omitempty" json:"visibleToOrgIds,omitempty"`
	Description         string         `bson:"description" json:"description"`
	TimeOfOccurrence    string         `bson:"timeOfOccurrence" json:"timeOfOccurrence"`
	Location            ReportLocation `bson:"location" json:"location"`

	MediaURLs []string      `bson:"mediaUrls,omitempty" json:"mediaUrls,omitempty"`
	Media     []ReportMedia `bson:"media,omitempty" json:"media,omitempty"`

	Status    string    `bson:"status" json:"status"`
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time `bson:"updatedAt" json:"updatedAt"`
}
