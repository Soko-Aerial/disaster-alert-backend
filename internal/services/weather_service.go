package services

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"disaster_alert_backend/internal/models"
	"disaster_alert_backend/internal/repositories"
)

const openWeatherCurrentURL = "https://api.openweathermap.org/data/2.5/weather"
const openWeatherForecastURL = "https://api.openweathermap.org/data/2.5/forecast"

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

// ===============================
// CURRENT WEATHER FOR UI
// ===============================

type CurrentWeatherResponse struct {
	Temperature float64   `json:"temperature"`
	FeelsLike   float64   `json:"feelsLike"`
	Condition   string    `json:"condition"`
	Description string    `json:"description"`
	WindKmh     float64   `json:"windKmh"`
	Humidity    int       `json:"humidity"`
	RainChance  int       `json:"rainChance"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func (s *WeatherService) FetchCurrentWeather(
	lat float64,
	lng float64,
	country string,
) (*CurrentWeatherResponse, error) {
	if !s.enabled {
		return nil, fmt.Errorf("OpenWeather source is disabled")
	}

	if s.apiKey == "" {
		return nil, fmt.Errorf("OPENWEATHER_API_KEY is not configured")
	}

	response, err := s.fetchCurrentOpenWeather(lat, lng)
	if err != nil {
		return nil, err
	}

	condition := openWeatherCondition{
		Main:        "Weather",
		Description: "Current weather",
	}

	if len(response.Weather) > 0 {
		condition = response.Weather[0]
	}

	return &CurrentWeatherResponse{
		Temperature: response.Main.Temp,
		FeelsLike:   response.Main.FeelsLike,
		Condition:   condition.Main,
		Description: condition.Description,
		WindKmh:     response.Wind.Speed * 3.6,
		Humidity:    response.Main.Humidity,
		RainChance:  rainChanceFromCondition(condition),
		UpdatedAt:   time.Now().UTC(),
	}, nil
}

// ===============================
// FORECAST FOR UI
// ===============================

type WeatherForecastItem struct {
	TimeLabel   string  `json:"timeLabel"`
	Temperature float64 `json:"temperature"`
	Condition   string  `json:"condition"`
	Description string  `json:"description"`
}

func (s *WeatherService) FetchWeatherForecast(
	lat float64,
	lng float64,
	country string,
) ([]WeatherForecastItem, error) {
	if !s.enabled {
		return []WeatherForecastItem{}, fmt.Errorf("OpenWeather source is disabled")
	}

	if s.apiKey == "" {
		return []WeatherForecastItem{}, fmt.Errorf("OPENWEATHER_API_KEY is not configured")
	}

	requestURL := fmt.Sprintf(
		"%s?lat=%f&lon=%f&appid=%s&units=metric",
		openWeatherForecastURL,
		lat,
		lng,
		url.QueryEscape(s.apiKey),
	)

	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		return []WeatherForecastItem{}, err
	}

	req.Header.Set("Accept", "application/json")

	res, err := s.client.Do(req)
	if err != nil {
		return []WeatherForecastItem{}, err
	}
	defer res.Body.Close()

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return []WeatherForecastItem{}, err
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		log.Printf(
			"OpenWeather forecast failed. Status: %d Body: %s\n",
			res.StatusCode,
			string(bodyBytes),
		)

		return []WeatherForecastItem{}, fmt.Errorf(
			"OpenWeather forecast request failed with status %d: %s",
			res.StatusCode,
			string(bodyBytes),
		)
	}

	var response openWeatherForecastResponse
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return []WeatherForecastItem{}, fmt.Errorf("failed to decode OpenWeather forecast response: %w", err)
	}

	forecast := make([]WeatherForecastItem, 0)

	for index, item := range response.List {
		if index >= 8 {
			break
		}

		condition := openWeatherCondition{
			Main:        "Weather",
			Description: "Forecast weather",
		}

		if len(item.Weather) > 0 {
			condition = item.Weather[0]
		}

		forecast = append(forecast, WeatherForecastItem{
			TimeLabel:   forecastTimeLabel(item.DateText, index),
			Temperature: item.Main.Temp,
			Condition:   condition.Main,
			Description: condition.Description,
		})
	}

	return forecast, nil
}

// ===============================
// WEATHER ALERTS
// ===============================

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

	response, err := s.fetchCurrentOpenWeather(lat, lng)
	if err != nil {
		return []models.Alert{}, err
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

// ===============================
// OPENWEATHER API HELPERS
// ===============================

func (s *WeatherService) fetchCurrentOpenWeather(
	lat float64,
	lng float64,
) (*openWeatherResponse, error) {
	requestURL := fmt.Sprintf(
		"%s?lat=%f&lon=%f&appid=%s&units=metric",
		openWeatherCurrentURL,
		lat,
		lng,
		url.QueryEscape(s.apiKey),
	)

	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	res, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		log.Printf(
			"OpenWeather failed. Status: %d Body: %s\n",
			res.StatusCode,
			string(bodyBytes),
		)

		return nil, fmt.Errorf(
			"OpenWeather request failed with status %d: %s",
			res.StatusCode,
			string(bodyBytes),
		)
	}

	var response openWeatherResponse
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return nil, fmt.Errorf("failed to decode OpenWeather response: %w", err)
	}

	return &response, nil
}

// ===============================
// OPENWEATHER MODELS
// ===============================

type openWeatherResponse struct {
	Weather []openWeatherCondition `json:"weather"`
	Main    openWeatherMain        `json:"main"`
	Wind    openWeatherWind        `json:"wind"`
	Name    string                 `json:"name"`
}

type openWeatherForecastResponse struct {
	List []openWeatherForecastListItem `json:"list"`
}

type openWeatherForecastListItem struct {
	DateText string                 `json:"dt_txt"`
	Main     openWeatherMain        `json:"main"`
	Weather  []openWeatherCondition `json:"weather"`
	Pop      float64                `json:"pop"`
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

// ===============================
// WEATHER ALERT MAPPING
// ===============================

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

func mapCurrentWeatherCategory(condition openWeatherCondition) string {
	id := condition.ID

	switch {
	case id >= 200 && id < 300:
		return "thunderstorm"
	case id >= 300 && id < 400:
		return "drizzle"
	case id >= 500 && id < 600:
		return "rain"
	case id >= 600 && id < 700:
		return "snow"
	case id >= 700 && id < 800:
		return "visibility"
	case id == 800:
		return "clear"
	case id > 800:
		return "clouds"
	default:
		return "weather"
	}
}

func mapCurrentWeatherSeverity(condition openWeatherCondition) string {
	id := condition.ID

	switch {
	case id >= 200 && id < 300:
		return "high"
	case id >= 500 && id < 600:
		if id >= 502 {
			return "high"
		}
		return "medium"
	case id >= 600 && id < 700:
		return "medium"
	case id >= 700 && id < 800:
		return "medium"
	default:
		return "low"
	}
}

func mapCurrentWeatherRadiusKm(condition openWeatherCondition) float64 {
	id := condition.ID

	switch {
	case id >= 200 && id < 300:
		return 30
	case id >= 500 && id < 600:
		return 20
	case id >= 600 && id < 700:
		return 25
	case id >= 700 && id < 800:
		return 15
	default:
		return 10
	}
}

func mapCurrentWeatherConfidence(condition openWeatherCondition) float64 {
	id := condition.ID

	switch {
	case id >= 200 && id < 300:
		return 0.85
	case id >= 500 && id < 600:
		return 0.80
	case id >= 600 && id < 700:
		return 0.78
	case id >= 700 && id < 800:
		return 0.75
	default:
		return 0.65
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

// ===============================
// UI HELPER LOGIC
// ===============================

func rainChanceFromCondition(condition openWeatherCondition) int {
	id := condition.ID

	switch {
	case id >= 200 && id < 300:
		return 90
	case id >= 300 && id < 400:
		return 65
	case id >= 500 && id < 600:
		return 80
	case id >= 600 && id < 700:
		return 70
	case id >= 700 && id < 800:
		return 30
	case id > 800:
		return 20
	default:
		return 5
	}
}

func forecastTimeLabel(dateText string, index int) string {
	if index == 0 {
		return "Now"
	}

	parsed, err := time.Parse("2006-01-02 15:04:05", dateText)
	if err != nil {
		return dateText
	}

	return parsed.Format("15:04")
}