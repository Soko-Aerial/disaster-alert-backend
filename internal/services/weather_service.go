package services

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
	"log"

	"disaster_alert_backend/internal/models"
	"disaster_alert_backend/internal/repositories"
)

const openWeatherCurrentURL = "https://api.openweathermap.org/data/2.5/weather"

type WeatherService struct {
	apiKey    string
	enabled   bool
	alertRepo *repositories.AlertRepository
	client    *http.Client
}

func NewWeatherService(
	apiKey string,
	enabled bool,
	alertRepo *repositories.AlertRepository,
) *WeatherService {
	return &WeatherService{
		apiKey:    strings.TrimSpace(apiKey),
		enabled:   enabled,
		alertRepo: alertRepo,
		client: &http.Client{
			Timeout: 25 * time.Second,
		},
	}
}

func (s *WeatherService) FetchWeatherAlertsForLocation(
	lat float64,
	lng float64,
	name string,
	country string,
) ([]models.Alert, error) {
	if !s.enabled {
		return []models.Alert{}, fmt.Errorf("OpenWeather source is disabled")
	}

	if s.apiKey == "" {
		return []models.Alert{}, fmt.Errorf("OPENWEATHER_API_KEY is not configured")
	}

	requestURL := fmt.Sprintf(
		"%s?lat=%f&lon=%f&appid=%s&units=metric",
		openWeatherCurrentURL,
		lat,
		lng,
		url.QueryEscape(s.apiKey),
	)

	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		return []models.Alert{}, err
	}

	req.Header.Set("Accept", "application/json")

	res, err := s.client.Do(req)
	if err != nil {
		return []models.Alert{}, err
	}
	defer res.Body.Close()

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return []models.Alert{}, err
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		log.Printf(
			"OpenWeather failed. Status: %d Body: %s\n",
			res.StatusCode,
			string(bodyBytes),
		)

		return []models.Alert{}, fmt.Errorf(
			"OpenWeather request failed with status %d: %s",
			res.StatusCode,
			string(bodyBytes),
		)
	}

	var response openWeatherResponse
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return []models.Alert{}, fmt.Errorf("failed to decode OpenWeather response: %w", err)
	}

	alerts := make([]models.Alert, 0)

	if len(response.Weather) == 0 {
		return alerts, nil
	}

	condition := response.Weather[0]

	if !shouldCreateWeatherAdvisory(condition) {
		return alerts, nil
	}

	alert := s.weatherConditionToAlert(
		condition,
		response.Main,
		response.Wind,
		lat,
		lng,
		name,
		country,
	)

	savedAlert, err := s.alertRepo.UpsertExternalAlert(alert)
	if err != nil {
		return []models.Alert{}, err
	}

	alerts = append(alerts, *savedAlert)

		return alerts, nil
	}

type openWeatherResponse struct {
	Weather []openWeatherCondition `json:"weather"`
	Main    openWeatherMain        `json:"main"`
	Wind    openWeatherWind        `json:"wind"`
	Name    string                 `json:"name"`
}

type openWeatherCondition struct {
	ID          int    `json:"id"`
	Main        string `json:"main"`
	Description string `json:"description"`
}

type openWeatherMain struct {
	Temp      float64 `json:"temp"`
	FeelsLike float64 `json:"feels_like"`
	Humidity  int     `json:"humidity"`
}

type openWeatherWind struct {
	Speed float64 `json:"speed"`
}

func shouldCreateWeatherAdvisory(condition openWeatherCondition) bool {
	id := condition.ID

	switch {
	case id >= 200 && id < 300:
		return true // thunderstorm
	case id >= 300 && id < 400:
		return true // drizzle
	case id >= 500 && id < 600:
		return true // rain
	case id >= 600 && id < 700:
		return true // snow
	case id >= 700 && id < 800:
		return true // fog, dust, haze, smoke
	default:
		return false
	}
}

func (s *WeatherService) weatherConditionToAlert(
	condition openWeatherCondition,
	main openWeatherMain,
	wind openWeatherWind,
	lat float64,
	lng float64,
	name string,
	country string,
) models.Alert {
	now := time.Now().UTC()
	expiresAt := now.Add(6 * time.Hour)

	locationName := strings.TrimSpace(name)
	if locationName == "" {
		locationName = fmt.Sprintf("%.5f, %.5f", lat, lng)
	}

	title := fmt.Sprintf(
		"%s advisory - %s",
		strings.Title(condition.Main),
		locationName,
	)

	description := fmt.Sprintf(
		"Weather condition: %s. Temperature: %.1f°C. Feels like: %.1f°C. Humidity: %d%%. Wind speed: %.1f m/s.",
		condition.Description,
		main.Temp,
		main.FeelsLike,
		main.Humidity,
		wind.Speed,
	)

	return models.Alert{
		Title:       title,
		Description: description,
		Category:    mapCurrentWeatherCategory(condition),
		Severity:    mapCurrentWeatherSeverity(condition),
		Status:      "active",
		Location: models.AlertLocation{
			Latitude:  lat,
			Longitude: lng,
			Country:   strings.TrimSpace(country),
			Region:    locationName,
			Address:   locationName,
		},
		RadiusKm: mapCurrentWeatherRadiusKm(condition),
		SafetyInstructions: []string{
			"Monitor local weather conditions",
			"Avoid unnecessary movement if conditions worsen",
			"Follow instructions from local authorities",
		},
		SourceType: "external",
		SourceName: "openweather",
		ExternalID: buildCurrentWeatherExternalID(condition, lat, lng, locationName, country, now),
		SourceURL:  "https://openweathermap.org/",
		EventTime:  &now,
		ExpiresAt:  &expiresAt,
		Confidence: mapCurrentWeatherConfidence(condition),
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

func buildCurrentWeatherExternalID(
	condition openWeatherCondition,
	lat float64,
	lng float64,
	name string,
	country string,
	now time.Time,
) string {
	raw := fmt.Sprintf(
		"%d-%s-%s-%s-%.5f-%.5f-%s",
		condition.ID,
		condition.Main,
		name,
		country,
		lat,
		lng,
		now.Format("2006-01-02-15"),
	)

	hash := sha1.Sum([]byte(raw))
	return "openweather-current-" + hex.EncodeToString(hash[:])
}

func mapCurrentWeatherCategory(condition openWeatherCondition) string {
	id := condition.ID

	switch {
	case id >= 200 && id < 300:
		return "weather"
	case id >= 300 && id < 600:
		return "flood"
	case id >= 600 && id < 700:
		return "weather"
	case id >= 700 && id < 800:
		return "weather"
	default:
		return "weather"
	}
}

func mapCurrentWeatherSeverity(condition openWeatherCondition) string {
	id := condition.ID

	switch {
	case id >= 200 && id < 300:
		return "high" // thunderstorm
	case id >= 502 && id <= 504:
		return "high" // heavy rain
	case id == 511:
		return "high" // freezing rain
	case id >= 520 && id <= 531:
		return "high" // shower rain
	case id >= 500 && id < 600:
		return "medium"
	case id >= 700 && id < 800:
		return "medium"
	default:
		return "medium"
	}
}

func mapCurrentWeatherRadiusKm(condition openWeatherCondition) float64 {
	id := condition.ID

	switch {
	case id >= 200 && id < 300:
		return 100
	case id >= 300 && id < 600:
		return 80
	case id >= 700 && id < 800:
		return 50
	default:
		return 80
	}
}

func mapCurrentWeatherConfidence(condition openWeatherCondition) float64 {
	switch mapCurrentWeatherSeverity(condition) {
	case "high":
		return 0.80
	case "medium":
		return 0.65
	default:
		return 0.55
	}
}