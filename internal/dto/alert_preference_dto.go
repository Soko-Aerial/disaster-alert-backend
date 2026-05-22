package dto

type UpdateAlertPreferenceRequest struct {
	Fire               *bool `json:"fire"`
	Flood              *bool `json:"flood"`
	Weather            *bool `json:"weather"`
	Earthquake         *bool `json:"earthquake"`
	Health             *bool `json:"health"`
	Conflict           *bool `json:"conflict"`
	Drought            *bool `json:"drought"`
	Protests           *bool `json:"protests"`
	Robbery            *bool `json:"robbery"`
	Munitions          *bool `json:"munitions"`
	Galamsey           *bool `json:"galamsey"`
	UnverifiedActivity *bool `json:"unverifiedActivity"`
	CriticalAlerts     *bool `json:"criticalAlerts"`
}