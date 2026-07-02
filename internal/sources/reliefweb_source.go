package sources

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"disaster_alert_backend/internal/models"
)

const reliefWebReportsURL = "https://api.reliefweb.int/v2/reports"

type ReliefWebSource struct {
	appName string
	limit   int
	client  *http.Client
}

func NewReliefWebSource(appName string, limitValue string) *ReliefWebSource {
	appName = strings.TrimSpace(appName)
	if appName == "" {
		appName = "disaster-alert-backend"
	}

	limit := 20
	if parsedLimit, err := strconv.Atoi(limitValue); err == nil && parsedLimit > 0 {
		limit = parsedLimit
	}

	if limit > 100 {
		limit = 100
	}

	return &ReliefWebSource{
		appName: appName,
		limit:   limit,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *ReliefWebSource) Name() string {
	return "reliefweb"
}

func (s *ReliefWebSource) FetchAlerts() ([]models.Alert, error) {
	requestBody := reliefWebRequest{
		Limit: s.limit,
		Query: reliefWebQuery{
			Value:    `disaster OR flood OR earthquake OR cyclone OR wildfire OR drought OR epidemic OR outbreak OR cholera OR conflict OR violence OR displacement OR emergency OR "state of emergency" OR evacuation`,
			Fields:   []string{"title", "body", "headline"},
			Operator: "OR",
		},
		Preset: "latest",
		Fields: reliefWebFields{
			Include: []string{
				"id",
				"url",
				"title",
				"body",
				"headline",
				"date.created",
				"date.original",
				"country.name",
				"country.iso3",
				"disaster.name",
				"disaster.type.name",
				"source.name",
				"format.name",
				"primary_country.name",
				"primary_country.iso3",
			},
		},
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	requestURL := fmt.Sprintf(
		"%s?appname=%s",
		reliefWebReportsURL,
		url.QueryEscape(s.appName),
	)

	req, err := http.NewRequest(
		http.MethodPost,
		requestURL,
		bytes.NewReader(bodyBytes),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	res, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	responseBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf(
			"ReliefWeb request failed with status %d: %s",
			res.StatusCode,
			string(responseBytes),
		)
	}

	var response reliefWebResponse

	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return nil, fmt.Errorf("failed to decode ReliefWeb response: %w", err)
	}

	alerts := make([]models.Alert, 0)

	for _, item := range response.Data {
		alert, ok := s.reportToAlert(item)
		if !ok {
			continue
		}

		alerts = append(alerts, alert)
	}

	log.Printf("ReliefWeb fetched %d humanitarian alerts\n", len(alerts))

	return alerts, nil
}

func (s *ReliefWebSource) reportToAlert(item reliefWebItem) (models.Alert, bool) {
	now := time.Now().UTC()

	fields := item.Fields

	title := strings.TrimSpace(fields.Title)
	if title == "" {
		return models.Alert{}, false
	}

	countryName := firstReliefWebCountryName(fields)
	countryCode := firstReliefWebCountryCode(fields)

	description := firstNonEmpty(
		fields.Headline,
		shortenText(stripHTML(fields.Body), 600),
		title,
	)

	eventTime := parseReliefWebTime(firstNonEmpty(
		fields.Date.Original,
		fields.Date.Created,
	))

	expiresAt := now.Add(14 * 24 * time.Hour)

	sourceURL := firstNonEmpty(
		fields.URL,
		fmt.Sprintf("https://reliefweb.int/report/%s", item.ID),
	)

	category := mapReliefWebCategory(fields)
	severity := mapReliefWebSeverity(fields)

	alert := models.Alert{
		Title:       title,
		Description: description,
		Summary:     description,
		Category:    category,
		Severity:    severity,
		Status:      "active",
		Location: models.AlertLocation{
			Latitude:  0,
			Longitude: 0,
			Country:   countryName,
			Region:    countryName,
			Address:   countryName,
		},
		RadiusKm: mapReliefWebRadiusKm(category),
		SafetyInstructions: []string{
			"Follow updates from verified humanitarian and emergency sources",
			"Monitor official instructions from local authorities",
			"Confirm local impact before taking operational action",
		},
		SourceType: "external",
		SourceName: "reliefweb",
		ExternalID: fmt.Sprintf("reliefweb-%s", item.ID),
		SourceURL:  sourceURL,
		EventTime:  eventTime,
		ExpiresAt:  &expiresAt,
		Confidence: mapReliefWebConfidence(severity),
		IsVerified: true,
		Tags: []string{
			category,
			severity,
			"reliefweb",
			strings.ToLower(strings.TrimSpace(countryName)),
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	if countryCode != "" && alert.Location.Country == "" {
		alert.Location.Country = countryCode
	}

	return alert, true
}

type reliefWebRequest struct {
	Limit  int             `json:"limit"`
	Query  reliefWebQuery  `json:"query"`
	Preset string          `json:"preset"`
	Fields reliefWebFields `json:"fields"`
}

type reliefWebQuery struct {
	Value    string   `json:"value"`
	Fields   []string `json:"fields"`
	Operator string   `json:"operator"`
}

type reliefWebFields struct {
	Include []string `json:"include"`
}

type reliefWebResponse struct {
	Data []reliefWebItem `json:"data"`
}

type reliefWebItem struct {
	ID     string          `json:"id"`
	Score  float64         `json:"score"`
	Fields reliefWebReport `json:"fields"`
}

type reliefWebReport struct {
	URL            string                 `json:"url"`
	Title          string                 `json:"title"`
	Body           string                 `json:"body"`
	Headline       string                 `json:"headline"`
	Date           reliefWebDate          `json:"date"`
	Country        []reliefWebNameCode    `json:"country"`
	PrimaryCountry reliefWebNameCode      `json:"primary_country"`
	Disaster       []reliefWebDisaster    `json:"disaster"`
	Source         []reliefWebNameOnly    `json:"source"`
	Format         []reliefWebNameOnly    `json:"format"`
	Extra          map[string]interface{} `json:"-"`
}

type reliefWebDate struct {
	Created  string `json:"created"`
	Original string `json:"original"`
}

type reliefWebNameCode struct {
	Name string `json:"name"`
	ISO3 string `json:"iso3"`
}

type reliefWebDisaster struct {
	Name string            `json:"name"`
	Type reliefWebNameOnly `json:"type"`
}

type reliefWebNameOnly struct {
	Name string `json:"name"`
}

func firstReliefWebCountryName(fields reliefWebReport) string {
	if strings.TrimSpace(fields.PrimaryCountry.Name) != "" {
		return strings.TrimSpace(fields.PrimaryCountry.Name)
	}

	if len(fields.Country) > 0 {
		return strings.TrimSpace(fields.Country[0].Name)
	}

	return ""
}

func firstReliefWebCountryCode(fields reliefWebReport) string {
	if strings.TrimSpace(fields.PrimaryCountry.ISO3) != "" {
		return strings.TrimSpace(fields.PrimaryCountry.ISO3)
	}

	if len(fields.Country) > 0 {
		return strings.TrimSpace(fields.Country[0].ISO3)
	}

	return ""
}

func mapReliefWebCategory(fields reliefWebReport) string {
	text := strings.ToLower(fields.Title + " " + fields.Headline + " " + fields.Body)

	for _, disaster := range fields.Disaster {
		text += " " + strings.ToLower(disaster.Name)
		text += " " + strings.ToLower(disaster.Type.Name)
	}

	switch {
	case strings.Contains(text, "flood"):
		return "flood"
	case strings.Contains(text, "earthquake"):
		return "earthquake"
	case strings.Contains(text, "cyclone"), strings.Contains(text, "storm"), strings.Contains(text, "hurricane"):
		return "weather"
	case strings.Contains(text, "wildfire"), strings.Contains(text, "fire"):
		return "fire"
	case strings.Contains(text, "drought"):
		return "drought"
	case strings.Contains(text, "epidemic"), strings.Contains(text, "outbreak"), strings.Contains(text, "cholera"), strings.Contains(text, "disease"):
		return "health"
	case strings.Contains(text, "conflict"), strings.Contains(text, "war"), strings.Contains(text, "violence"), strings.Contains(text, "displacement"):
		return "conflict"
	default:
		return "disaster"
	}
}

func mapReliefWebSeverity(fields reliefWebReport) string {
	text := strings.ToLower(fields.Title + " " + fields.Headline + " " + fields.Body)

	switch {
	case strings.Contains(text, "catastrophic"),
		strings.Contains(text, "critical"),
		strings.Contains(text, "state of emergency"),
		strings.Contains(text, "deadly"),
		strings.Contains(text, "killed"),
		strings.Contains(text, "deaths"),
		strings.Contains(text, "mass casualties"):
		return "critical"

	case strings.Contains(text, "emergency"),
		strings.Contains(text, "severe"),
		strings.Contains(text, "major"),
		strings.Contains(text, "urgent"),
		strings.Contains(text, "evacuation"),
		strings.Contains(text, "displaced"),
		strings.Contains(text, "displacement"):
		return "high"

	case strings.Contains(text, "warning"),
		strings.Contains(text, "alert"),
		strings.Contains(text, "watch"),
		strings.Contains(text, "risk"):
		return "medium"

	default:
		return "medium"
	}
}
func mapReliefWebRadiusKm(category string) float64 {
	switch category {
	case "flood":
		return 120
	case "earthquake":
		return 150
	case "weather":
		return 150
	case "fire":
		return 80
	case "drought":
		return 300
	case "health":
		return 200
	case "conflict":
		return 200
	default:
		return 150
	}
}

func mapReliefWebConfidence(severity string) float64 {
	switch severity {
	case "critical":
		return 0.90
	case "high":
		return 0.85
	case "medium":
		return 0.70
	default:
		return 0.60
	}
}

func parseReliefWebTime(value string) *time.Time {
	value = strings.TrimSpace(value)

	if value == "" {
		return nil
	}

	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04:05-07:00",
		"2006-01-02T15:04:05+00:00",
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

func stripHTML(value string) string {
	value = strings.ReplaceAll(value, "\n", " ")
	value = strings.ReplaceAll(value, "\r", " ")
	value = strings.ReplaceAll(value, "\t", " ")

	replacements := []string{
		"<p>", " ",
		"</p>", " ",
		"<br>", " ",
		"<br/>", " ",
		"<br />", " ",
		"<strong>", "",
		"</strong>", "",
		"<em>", "",
		"</em>", "",
		"&nbsp;", " ",
		"&amp;", "&",
	}

	for i := 0; i < len(replacements); i += 2 {
		value = strings.ReplaceAll(value, replacements[i], replacements[i+1])
	}

	return strings.Join(strings.Fields(value), " ")
}

func shortenText(value string, maxLength int) string {
	value = strings.TrimSpace(value)

	if value == "" {
		return ""
	}

	if len(value) <= maxLength {
		return value
	}

	return value[:maxLength] + "..."
}
