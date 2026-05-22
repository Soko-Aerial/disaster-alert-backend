package sources

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"disaster_alert_backend/internal/models"
)

const nasaFIRMSBaseURL = "https://firms.modaps.eosdis.nasa.gov/api/area/csv"

type NASAFIRMSSource struct {
	mapKey   string
	source   string
	area     string
	dayRange string
	limit    int
	client   *http.Client
}

func NewNASAFIRMSSource(
	mapKey string,
	source string,
	area string,
	dayRange string,
	limitValue string,
) *NASAFIRMSSource {
	if source == "" {
		source = "VIIRS_SNPP_NRT"
	}

	if area == "" {
		area = "-20,-5,25,25"
	}

	if dayRange == "" {
		dayRange = "1"
	}

	limit := 50

	if parsedLimit, err := strconv.Atoi(limitValue); err == nil && parsedLimit > 0 {
		limit = parsedLimit
	}

	if limit > 500 {
		limit = 500
	}

	return &NASAFIRMSSource{
		mapKey:   mapKey,
		source:   source,
		area:     area,
		dayRange: dayRange,
		limit: limit,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *NASAFIRMSSource) Name() string {
	return "nasa_firms"
}

func (s *NASAFIRMSSource) FetchAlerts() ([]models.Alert, error) {
	if strings.TrimSpace(s.mapKey) == "" {
		log.Println("NASA FIRMS MAP_KEY is empty. Skipping NASA FIRMS sync.")
		return []models.Alert{}, nil
	}

	url := fmt.Sprintf(
		"%s/%s/%s/%s/%s",
		nasaFIRMSBaseURL,
		s.mapKey,
		s.source,
		s.area,
		s.dayRange,
	)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "text/csv")

	res, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("NASA FIRMS request failed with status %d: %s", res.StatusCode, string(bodyBytes))
	}

	reader := csv.NewReader(res.Body)
	reader.FieldsPerRecord = -1

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read NASA FIRMS CSV: %w", err)
	}

	if len(records) <= 1 {
		log.Println("NASA FIRMS returned no fire hotspots")
		return []models.Alert{}, nil
	}

	headerMap := buildCSVHeaderMap(records[0])
	alerts := make([]models.Alert, 0)

	for _, record := range records[1:] {
		if len(alerts) >= s.limit {
			break
		}

		alert, ok := s.recordToAlert(record, headerMap)
		if !ok {
			continue
		}

		alerts = append(alerts, alert)
	}

	log.Printf(
		"NASA FIRMS fetched %d fire hotspot alerts after limit %d\n",
		len(alerts),
		s.limit,
	)

	return alerts, nil
}

func (s *NASAFIRMSSource) recordToAlert(
	record []string,
	headerMap map[string]int,
) (models.Alert, bool) {
	now := time.Now().UTC()

	lat, ok := getCSVFloat(record, headerMap, "latitude")
	if !ok {
		return models.Alert{}, false
	}

	lng, ok := getCSVFloat(record, headerMap, "longitude")
	if !ok {
		return models.Alert{}, false
	}

	acqDate := getCSVString(record, headerMap, "acq_date")
	acqTime := getCSVString(record, headerMap, "acq_time")
	confidenceRaw := getCSVString(record, headerMap, "confidence")
	satellite := getCSVString(record, headerMap, "satellite")
	instrument := getCSVString(record, headerMap, "instrument")
	brightness := getCSVString(record, headerMap, "bright_ti4")
	if brightness == "" {
		brightness = getCSVString(record, headerMap, "brightness")
	}

	eventTime := parseFIRMSTime(acqDate, acqTime)

	externalID := buildFIRMSExternalID(
		s.source,
		lat,
		lng,
		acqDate,
		acqTime,
	)

	expiresAt := now.Add(24 * time.Hour)

	description := fmt.Sprintf(
		"Satellite detected active fire hotspot near %.5f, %.5f. Source: %s. Satellite: %s. Instrument: %s. Confidence: %s.",
		lat,
		lng,
		s.source,
		satellite,
		instrument,
		confidenceRaw,
	)

	if brightness != "" {
		description += " Brightness: " + brightness + "."
	}

	alert := models.Alert{
		Title:       "Active fire hotspot detected",
		Description: description,
		Category:    "fire",
		Severity:    mapFIRMSSeverity(confidenceRaw),
		Status:      "active",
		Location: models.AlertLocation{
			Latitude:  lat,
			Longitude: lng,
			Country:   "",
			Region:    "",
			Address:   fmt.Sprintf("%.5f, %.5f", lat, lng),
		},
		RadiusKm: 5,
		SafetyInstructions: []string{
			"Avoid the affected area",
			"Report visible fire or smoke to emergency authorities",
			"Monitor official fire service updates",
		},
		SourceType: "external",
		SourceName: "nasa_firms",
		ExternalID: externalID,
		SourceURL:  "https://firms.modaps.eosdis.nasa.gov/map/",
		EventTime:  eventTime,
		ExpiresAt:  &expiresAt,
		Confidence: mapFIRMSConfidence(confidenceRaw),
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	return alert, true
}

func buildCSVHeaderMap(headers []string) map[string]int {
	headerMap := make(map[string]int)

	for index, header := range headers {
		normalized := strings.ToLower(strings.TrimSpace(header))
		headerMap[normalized] = index
	}

	return headerMap
}

func getCSVString(record []string, headerMap map[string]int, key string) string {
	index, exists := headerMap[strings.ToLower(key)]
	if !exists || index >= len(record) {
		return ""
	}

	return strings.TrimSpace(record[index])
}

func getCSVFloat(record []string, headerMap map[string]int, key string) (float64, bool) {
	value := getCSVString(record, headerMap, key)
	if value == "" {
		return 0, false
	}

	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, false
	}

	return parsed, true
}

func parseFIRMSTime(acqDate string, acqTime string) *time.Time {
	acqDate = strings.TrimSpace(acqDate)
	acqTime = strings.TrimSpace(acqTime)

	if acqDate == "" {
		return nil
	}

	for len(acqTime) < 4 {
		acqTime = "0" + acqTime
	}

	dateTimeValue := fmt.Sprintf(
		"%s %s:%s",
		acqDate,
		acqTime[0:2],
		acqTime[2:4],
	)

	parsed, err := time.Parse("2006-01-02 15:04", dateTimeValue)
	if err != nil {
		return nil
	}

	return &parsed
}

func buildFIRMSExternalID(
	source string,
	lat float64,
	lng float64,
	acqDate string,
	acqTime string,
) string {
	return fmt.Sprintf(
		"nasa-firms-%s-%.5f-%.5f-%s-%s",
		strings.ToLower(source),
		lat,
		lng,
		strings.ReplaceAll(acqDate, "-", ""),
		acqTime,
	)
}

func mapFIRMSSeverity(confidence string) string {
	confidence = strings.ToLower(strings.TrimSpace(confidence))

	switch confidence {
	case "h", "high":
		return "high"
	case "n", "nominal", "medium":
		return "medium"
	case "l", "low":
		return "low"
	}

	value, err := strconv.ParseFloat(confidence, 64)
	if err != nil {
		return "medium"
	}

	switch {
	case value >= 80:
		return "high"
	case value >= 40:
		return "medium"
	default:
		return "low"
	}
}

func mapFIRMSConfidence(confidence string) float64 {
	confidence = strings.ToLower(strings.TrimSpace(confidence))

	switch confidence {
	case "h", "high":
		return 0.90
	case "n", "nominal", "medium":
		return 0.70
	case "l", "low":
		return 0.45
	}

	value, err := strconv.ParseFloat(confidence, 64)
	if err != nil {
		return 0.65
	}

	if value < 0 {
		value = 0
	}

	if value > 100 {
		value = 100
	}

	return value / 100
}