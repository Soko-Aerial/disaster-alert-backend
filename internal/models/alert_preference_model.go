package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AlertPreference struct {
	ID     primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID primitive.ObjectID `json:"userId" bson:"userId"`

	Fire               bool `json:"fire" bson:"fire"`
	Flood              bool `json:"flood" bson:"flood"`
	Weather            bool `json:"weather" bson:"weather"`
	Earthquake         bool `json:"earthquake" bson:"earthquake"`
	Health             bool `json:"health" bson:"health"`
	Conflict           bool `json:"conflict" bson:"conflict"`
	Drought            bool `json:"drought" bson:"drought"`
	Protests           bool `json:"protests" bson:"protests"`
	Robbery            bool `json:"robbery" bson:"robbery"`
	Munitions          bool `json:"munitions" bson:"munitions"`
	Galamsey           bool `json:"galamsey" bson:"galamsey"`
	UnverifiedActivity bool `json:"unverifiedActivity" bson:"unverifiedActivity"`
	CriticalAlerts     bool `json:"criticalAlerts" bson:"criticalAlerts"`

	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt"`
}