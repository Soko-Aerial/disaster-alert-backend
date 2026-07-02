package sources

import (
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

const gdeltDocAPIURL = "https://api.gdeltproject.org/api/v2/doc/doc"

type GDELTSource struct {
	limit    int
	timespan string
	client   *http.Client
}

func NewGDELTSource(limitValue string, timespan string) *GDELTSource {
	limit := 20

	if parsedLimit, err := strconv.Atoi(limitValue); err == nil && parsedLimit > 0 {
		limit = parsedLimit
	}

	if limit > 50 {
		limit = 50
	}

	timespan = strings.TrimSpace(timespan)
	if timespan == "" {
		timespan = "24h"
	}

	return &GDELTSource{
		limit:    limit,
		timespan: timespan,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (s *GDELTSource) Name() string {
	return "gdelt"
}

func (s *GDELTSource) FetchAlerts() ([]models.Alert, error) {
	query := buildGDELTQuery()

	requestURL := fmt.Sprintf(
		"%s?query=%s&mode=ArtList&format=json&sort=DateDesc&maxrecords=%d&timespan=%s",
		gdeltDocAPIURL,
		url.QueryEscape(query),
		s.limit,
		url.QueryEscape(s.timespan),
	)

	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	res, err := s.doRequestWithRetry(req, 3)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf(
			"GDELT request failed with status %d: %s",
			res.StatusCode,
			string(bodyBytes),
		)
	}

	var response gdeltResponse

	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return nil, fmt.Errorf("failed to decode GDELT response: %w", err)
	}

	alerts := make([]models.Alert, 0)

	for _, article := range response.Articles {

		alert, ok := s.articleToAlert(article)
		if !ok {
			continue
		}

		alerts = append(alerts, alert)
	}

	log.Printf("GDELT fetched %d intelligence alerts\n", len(alerts))

	return alerts, nil
}

func buildGDELTQuery() string {
	return `(flood OR earthquake OR wildfire OR "forest fire" OR cyclone OR storm OR drought OR epidemic OR outbreak OR conflict OR violence OR explosion OR emergency OR "disaster response")`
}

func (s *GDELTSource) articleToAlert(article gdeltArticle) (models.Alert, bool) {
	now := time.Now().UTC()

	title := strings.TrimSpace(article.Title)
	if title == "" {
		return models.Alert{}, false
	}

	articleURL := strings.TrimSpace(article.URL)
	if articleURL == "" {
		return models.Alert{}, false
	}

	eventTime := parseGDELTSeenDate(article.SeenDate)

	category := mapGDELTCategory(article)
	severity := mapGDELTSeverity(article)

	country := strings.TrimSpace(article.SourceCountry)
	if country == "" {
		country = strings.TrimSpace(article.Language)
	}

	expiresAt := now.Add(7 * 24 * time.Hour)

	description := fmt.Sprintf(
		"%s Source: %s. Domain: %s. Language: %s.",
		title,
		firstNonEmpty(article.SourceCountry, "unknown"),
		firstNonEmpty(article.Domain, "unknown"),
		firstNonEmpty(article.Language, "unknown"),
	)

	imageURLs := []string{}
	if strings.TrimSpace(article.SocialImage) != "" {
		imageURLs = append(imageURLs, strings.TrimSpace(article.SocialImage))
	}

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
			Country:   country,
			Region:    country,
			Address:   country,
		},
		RadiusKm: mapGDELTRadiusKm(category),
		SafetyInstructions: []string{
			"Treat this as news intelligence until verified by official sources",
			"Check the original article and official emergency channels",
			"Do not take operational action based on one media report only",
		},
		SourceType: "external",
		SourceName: "gdelt",
		ExternalID: buildGDELTExternalID(articleURL),
		SourceURL:  articleURL,
		ImageURLs:  imageURLs,
		EventTime:  eventTime,
		ExpiresAt:  &expiresAt,
		Confidence: mapGDELTConfidence(severity),
		IsVerified: false,
		Tags: []string{
			category,
			severity,
			"gdelt",
			strings.ToLower(strings.TrimSpace(country)),
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	return alert, true
}

type gdeltResponse struct {
	Articles []gdeltArticle `json:"articles"`
}

type gdeltArticle struct {
	URL           string `json:"url"`
	URLMobile     string `json:"url_mobile"`
	Title         string `json:"title"`
	SeenDate      string `json:"seendate"`
	SocialImage   string `json:"socialimage"`
	Domain        string `json:"domain"`
	Language      string `json:"language"`
	SourceCountry string `json:"sourcecountry"`
}

func buildGDELTExternalID(articleURL string) string {
	return "gdelt-" + hashString(articleURL)
}

func hashString(value string) string {
	hash := uint32(2166136261)

	for _, char := range value {
		hash ^= uint32(char)
		hash *= 16777619
	}

	return fmt.Sprintf("%x", hash)
}

func parseGDELTSeenDate(value string) *time.Time {
	value = strings.TrimSpace(value)

	if value == "" {
		return nil
	}

	layouts := []string{
		"20060102150400",
		"20060102T150400Z",
		time.RFC3339,
		"2006-01-02T15:04:05Z",
	}

	for _, layout := range layouts {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return &parsed
		}
	}

	return nil
}

func mapGDELTCategory(article gdeltArticle) string {
	text := strings.ToLower(article.Title + " " + article.Domain)

	switch {
	case strings.Contains(text, "flood"):
		return "flood"
	case strings.Contains(text, "earthquake"):
		return "earthquake"
	case strings.Contains(text, "wildfire"), strings.Contains(text, "forest fire"), strings.Contains(text, "fire"):
		return "fire"
	case strings.Contains(text, "cyclone"),
		strings.Contains(text, "storm"),
		strings.Contains(text, "hurricane"),
		strings.Contains(text, "rain"):
		return "weather"
	case strings.Contains(text, "drought"):
		return "drought"
	case strings.Contains(text, "epidemic"),
		strings.Contains(text, "outbreak"),
		strings.Contains(text, "cholera"),
		strings.Contains(text, "disease"):
		return "health"
	case strings.Contains(text, "conflict"),
		strings.Contains(text, "war"),
		strings.Contains(text, "violence"),
		strings.Contains(text, "attack"),
		strings.Contains(text, "explosion"):
		return "conflict"
	default:
		return "disaster"
	}
}

func mapGDELTSeverity(article gdeltArticle) string {
	text := strings.ToLower(article.Title)

	switch {
	case strings.Contains(text, "catastrophic"),
		strings.Contains(text, "massive"),
		strings.Contains(text, "deadly"),
		strings.Contains(text, "kills"),
		strings.Contains(text, "killed"),
		strings.Contains(text, "emergency"),
		strings.Contains(text, "severe"),
		strings.Contains(text, "major"):
		return "critical"

	case strings.Contains(text, "warning"),
		strings.Contains(text, "alert"),
		strings.Contains(text, "evacuation"),
		strings.Contains(text, "threat"):
		return "medium"

	default:
		return "medium"
	}
}

func mapGDELTRadiusKm(category string) float64 {
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
		return 100
	default:
		return 150
	}
}

func mapGDELTConfidence(severity string) float64 {
	switch severity {
	case "critical":
		return 0.80
	case "high":
		return 0.75
	case "medium":
		return 0.60
	default:
		return 0.50
	}
}

func (s *GDELTSource) doRequestWithRetry(
	req *http.Request,
	maxAttempts int,
) (*http.Response, error) {
	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		clonedReq := req.Clone(req.Context())

		res, err := s.client.Do(clonedReq)
		if err == nil {
			return res, nil
		}

		lastErr = err

		log.Printf(
			"GDELT request attempt %d/%d failed: %v\n",
			attempt,
			maxAttempts,
			err,
		)

		time.Sleep(time.Duration(attempt) * 2 * time.Second)
	}

	return nil, lastErr
}
