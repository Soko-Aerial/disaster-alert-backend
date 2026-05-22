package sources

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"disaster_alert_backend/internal/models"
)

const gdacsBaseURL = "https://www.gdacs.org/gdacsapi/api/events/geteventlist/SEARCH"

type GDACSSource struct {
	client *http.Client
}

func NewGDACSSource() *GDACSSource {
	return &GDACSSource{
		client: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
}

func (s *GDACSSource) Name() string {
	return "gdacs"
}

func (s *GDACSSource) FetchAlerts() ([]models.Alert, error) {
	toDate := time.Now().UTC()
	fromDate := toDate.AddDate(0, -6, 0)

	url := fmt.Sprintf(
		"%s?eventlist=EQ;TC;FL;VO;DR;WF&fromdate=%s&todate=%s&alertlevel=green;orange;red",
		gdacsBaseURL,
		fromDate.Format("2006-01-02"),
		toDate.Format("2006-01-02"),
	)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	res, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("gdacs request failed with status %d: %s", res.StatusCode, string(bodyBytes))
	}

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var response gdacsFeatureCollection

	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return nil, fmt.Errorf("failed to decode gdacs response: %w", err)
	}

	alerts := make([]models.Alert, 0)

	for _, feature := range response.Features {
		alert, ok := s.featureToAlert(feature)
		if !ok {
			continue
		}

		alerts = append(alerts, alert)
	}

	log.Printf("GDACS fetched %d alerts\n", len(alerts))

	return alerts, nil
}

func (s *GDACSSource) featureToAlert(feature gdacsFeature) (models.Alert, bool) {
	now := time.Now().UTC()

	lat, lng, ok := extractPoint(feature.Geometry)
	if !ok {
		return models.Alert{}, false
	}

	eventType := firstNonEmpty(
		feature.Properties.EventType.String(),
		feature.Properties.EventTypeCode.String(),
	)

	eventID := firstNonEmpty(
		feature.Properties.EventID.String(),
		feature.Properties.EventIDString.String(),
		feature.ID,
	)

	if eventID == "" {
		return models.Alert{}, false
	}

	alertLevel := strings.ToLower(strings.TrimSpace(feature.Properties.AlertLevel.String()))

	title := firstNonEmpty(
		feature.Properties.Name.String(),
		feature.Properties.Title.String(),
		fmt.Sprintf("%s Disaster Alert", strings.ToUpper(eventType)),
	)

	description := firstNonEmpty(
		feature.Properties.Description.String(),
		feature.Properties.EventName.String(),
		title,
	)

	country := firstNonEmpty(
		feature.Properties.Country.String(),
		feature.Properties.Countries.String(),
	)

	eventTime := parseGDACSTime(firstNonEmpty(
		feature.Properties.FromDate.String(),
		feature.Properties.EventDate.String(),
		feature.Properties.Date.String(),
	))

	expiresAt := gdacsExpiryTime(eventType, now)

	sourceURL := firstNonEmpty(
		feature.Properties.URL.String(),
		feature.Properties.Link.String(),
	)

	if sourceURL == "" && eventType != "" {
		sourceURL = fmt.Sprintf(
			"https://www.gdacs.org/report.aspx?eventtype=%s&eventid=%s",
			strings.ToUpper(eventType),
			eventID,
		)
	}

	alert := models.Alert{
		Title:       title,
		Description: description,
		Category:    mapGDACSCategory(eventType),
		Severity:    mapGDACSSeverity(alertLevel),
		Status:      "active",
		Location: models.AlertLocation{
			Latitude:  lat,
			Longitude: lng,
			Country:   country,
			Region:    country,
			Address:   country,
		},
		RadiusKm:  defaultGDACSRadiusKm(eventType),
		SafetyInstructions: []string{
			"Follow official emergency instructions",
			"Stay away from affected areas",
			"Monitor verified updates from authorities",
		},
		SourceType: "external",
		SourceName: "gdacs",
		ExternalID: fmt.Sprintf(
			"gdacs-%s-%s",
			strings.ToUpper(eventType),
			eventID,
		),
		SourceURL:  sourceURL,
		EventTime:  eventTime,
		ExpiresAt:  &expiresAt,
		Confidence: mapGDACSConfidence(alertLevel),
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	return alert, true
}

type gdacsFeatureCollection struct {
	Type     string         `json:"type"`
	Features []gdacsFeature `json:"features"`
}

type gdacsFeature struct {
	ID         string          `json:"id"`
	Type       string          `json:"type"`
	Geometry   gdacsGeometry   `json:"geometry"`
	Properties gdacsProperties `json:"properties"`
}

type gdacsGeometry struct {
	Type        string          `json:"type"`
	Coordinates json.RawMessage `json:"coordinates"`
}

type gdacsProperties struct {
	EventID       flexibleString `json:"eventid"`
	EventIDString flexibleString `json:"eventidstring"`
	EventType     flexibleString `json:"eventtype"`
	EventTypeCode flexibleString `json:"eventtypecode"`
	AlertLevel    flexibleString `json:"alertlevel"`
	Name          flexibleString `json:"name"`
	Title         flexibleString `json:"title"`
	EventName     flexibleString `json:"eventname"`
	Description   flexibleString `json:"description"`
	Country       flexibleString `json:"country"`
	Countries     flexibleString `json:"countries"`
	FromDate      flexibleString `json:"fromdate"`
	EventDate     flexibleString `json:"eventdate"`
	Date          flexibleString `json:"date"`
	URL           flexibleString `json:"url"`
	Link          flexibleString `json:"link"`
}

func gdacsExpiryTime(eventType string, now time.Time) time.Time {
	switch strings.ToUpper(strings.TrimSpace(eventType)) {
	case "EQ":
		return now.Add(7 * 24 * time.Hour)
	case "WF":
		return now.Add(10 * 24 * time.Hour)
	case "FL":
		return now.Add(10 * 24 * time.Hour)
	case "TC":
		return now.Add(10 * 24 * time.Hour)
	case "DR":
		return now.Add(30 * 24 * time.Hour)
	case "VO":
		return now.Add(30 * 24 * time.Hour)
	default:
		return now.Add(14 * 24 * time.Hour)
	}
}

func extractPoint(geometry gdacsGeometry) (float64, float64, bool) {
	if len(geometry.Coordinates) == 0 {
		return 0, 0, false
	}

	switch strings.ToLower(geometry.Type) {
	case "point":
		var coords []float64
		if err := json.Unmarshal(geometry.Coordinates, &coords); err != nil {
			return 0, 0, false
		}

		if len(coords) < 2 {
			return 0, 0, false
		}

		lng := coords[0]
		lat := coords[1]

		return lat, lng, true

	case "polygon":
		var coords [][][]float64
		if err := json.Unmarshal(geometry.Coordinates, &coords); err != nil {
			return 0, 0, false
		}

		if len(coords) == 0 || len(coords[0]) == 0 || len(coords[0][0]) < 2 {
			return 0, 0, false
		}

		lng := coords[0][0][0]
		lat := coords[0][0][1]

		return lat, lng, true

	case "multipolygon":
		var coords [][][][]float64
		if err := json.Unmarshal(geometry.Coordinates, &coords); err != nil {
			return 0, 0, false
		}

		if len(coords) == 0 ||
			len(coords[0]) == 0 ||
			len(coords[0][0]) == 0 ||
			len(coords[0][0][0]) < 2 {
			return 0, 0, false
		}

		lng := coords[0][0][0][0]
		lat := coords[0][0][0][1]

		return lat, lng, true
	}

	return 0, 0, false
}

func mapGDACSCategory(eventType string) string {
	switch strings.ToUpper(strings.TrimSpace(eventType)) {
	case "EQ":
		return "earthquake"
	case "TC":
		return "weather"
	case "FL":
		return "flood"
	case "VO":
		return "volcano"
	case "DR":
		return "drought"
	case "WF":
		return "fire"
	default:
		return "disaster"
	}
}

func mapGDACSSeverity(alertLevel string) string {
	switch strings.ToLower(strings.TrimSpace(alertLevel)) {
	case "green":
		return "low"
	case "orange":
		return "high"
	case "red":
		return "critical"
	default:
		return "medium"
	}
}

func mapGDACSConfidence(alertLevel string) float64 {
	switch strings.ToLower(strings.TrimSpace(alertLevel)) {
	case "green":
		return 0.55
	case "orange":
		return 0.80
	case "red":
		return 0.95
	default:
		return 0.65
	}
}

func defaultGDACSRadiusKm(eventType string) float64 {
	switch strings.ToUpper(strings.TrimSpace(eventType)) {
	case "EQ":
		return 150
	case "TC":
		return 500
	case "FL":
		return 100
	case "VO":
		return 80
	case "DR":
		return 300
	case "WF":
		return 50
	default:
		return 100
	}
}

func parseGDACSTime(value string) *time.Time {
	value = strings.TrimSpace(value)

	if value == "" {
		return nil
	}

	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}

	for _, layout := range layouts {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return &parsed
		}
	}

	return nil
}

type flexibleString string

func (fs *flexibleString) UnmarshalJSON(data []byte) error {
	var str string

	if err := json.Unmarshal(data, &str); err == nil {
		*fs = flexibleString(str)
		return nil
	}

	var num float64

	if err := json.Unmarshal(data, &num); err == nil {
		*fs = flexibleString(fmt.Sprintf("%.0f", num))
		return nil
	}

	var boolValue bool

	if err := json.Unmarshal(data, &boolValue); err == nil {
		if boolValue {
			*fs = "true"
		} else {
			*fs = "false"
		}
		return nil
	}

	*fs = ""
	return nil
}

func (fs flexibleString) String() string {
	return string(fs)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}

	return ""
}