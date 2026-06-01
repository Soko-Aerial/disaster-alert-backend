package services

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"disaster_alert_backend/internal/utils"
)

const serpAPIGoogleNewsURL = "https://serpapi.com/search.json"

type NewsService struct {
	apiKey  string
	enabled bool
	client  *http.Client
}

func NewNewsService(
	apiKey string,
	enabled bool,
) *NewsService {
	return &NewsService{
		apiKey:  strings.TrimSpace(apiKey),
		enabled: enabled,
		client: &http.Client{
			Timeout: 25 * time.Second,
		},
	}
}

type NewsArticle struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Content     string    `json:"content,omitempty"`
	ImageURL    string    `json:"imageUrl"`
	URL         string    `json:"url"`
	SourceName string    `json:"sourceName"`
	SourceURL  string    `json:"sourceUrl,omitempty"`
	PublishedAt time.Time `json:"publishedAt"`
	Scope       string    `json:"scope"`
	Country     string    `json:"country,omitempty"`
	Category    string    `json:"category"`
}

type serpAPIGoogleNewsResponse struct {
	NewsResults []serpAPINewsItem `json:"news_results"`
	Error       string            `json:"error"`
}

type serpAPINewsItem struct {
	Position       int                `json:"position"`
	Title          string             `json:"title"`
	Link           string             `json:"link"`
	Snippet        string             `json:"snippet"`
	Date           string             `json:"date"`
	ISODate        string             `json:"iso_date"`
	Thumbnail      string             `json:"thumbnail"`
	ThumbnailSmall string             `json:"thumbnail_small"`
	Source         serpAPINewsSource  `json:"source"`
	Stories        []serpAPINewsStory `json:"stories"`
}

type serpAPINewsStory struct {
	Position       int               `json:"position"`
	Title          string            `json:"title"`
	Link           string            `json:"link"`
	Snippet        string            `json:"snippet"`
	Date           string            `json:"date"`
	ISODate        string            `json:"iso_date"`
	Thumbnail      string            `json:"thumbnail"`
	ThumbnailSmall string            `json:"thumbnail_small"`
	Source         serpAPINewsSource `json:"source"`
}

type serpAPINewsSource struct {
	Name    string   `json:"name"`
	Icon    string   `json:"icon"`
	Authors []string `json:"authors"`
}

func (s *NewsService) FetchNews(
	scope string,
	country string,
) ([]NewsArticle, error) {
	if !s.enabled {
		return []NewsArticle{}, fmt.Errorf("news source is disabled")
	}

	if strings.TrimSpace(s.apiKey) == "" {
		return []NewsArticle{}, fmt.Errorf("SERPAPI API key is not configured")
	}

	scope = strings.ToLower(strings.TrimSpace(scope))
	if scope == "" {
		scope = "local"
	}

	if scope == "global" || scope == "worldwide" {
		scope = "world"
	}

	country = utils.NormalizeCountryCode(country)

	response, err := s.fetchSerpAPIGoogleNews(scope, country)
	if err != nil {
		return []NewsArticle{}, err
	}

	articles := make([]NewsArticle, 0)

	for _, item := range response.NewsResults {
		articles = appendNewsItem(
			articles,
			item.Title,
			item.Snippet,
			item.Link,
			item.Thumbnail,
			item.ThumbnailSmall,
			item.Date,
			item.ISODate,
			item.Source.Name,
			scope,
			country,
		)

		for _, story := range item.Stories {
			articles = appendNewsItem(
				articles,
				story.Title,
				story.Snippet,
				story.Link,
				story.Thumbnail,
				story.ThumbnailSmall,
				story.Date,
				story.ISODate,
				story.Source.Name,
				scope,
				country,
			)
		}
	}

	articles = removeDuplicateNews(articles)
	articles = filterRecentNews(articles, 7*24*time.Hour)

	sort.SliceStable(articles, func(i, j int) bool {
		return articles[i].PublishedAt.After(articles[j].PublishedAt)
	})

	if len(articles) > 10 {
		articles = articles[:10]
	}

	return articles, nil
}

func (s *NewsService) fetchSerpAPIGoogleNews(
	scope string,
	country string,
) (*serpAPIGoogleNewsResponse, error) {
	queryValues := url.Values{}
	queryValues.Set("engine", "google_news")
	queryValues.Set("q", buildNewsQuery(scope, country))
	queryValues.Set("api_key", s.apiKey)
	queryValues.Set("hl", "en")
	queryValues.Set("no_cache", "true")

	if scope == "local" && strings.TrimSpace(country) != "" {
		queryValues.Set("gl", strings.ToLower(country))
	} else {
		queryValues.Set("gl", "us")
	}

	requestURL := serpAPIGoogleNewsURL + "?" + queryValues.Encode()

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
			"SerpApi Google News request failed. Status: %d Body: %s\n",
			res.StatusCode,
			string(bodyBytes),
		)

		return nil, fmt.Errorf(
			"SerpApi Google News request failed with status %d: %s",
			res.StatusCode,
			string(bodyBytes),
		)
	}

	var response serpAPIGoogleNewsResponse
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return nil, fmt.Errorf("failed to decode SerpApi Google News response: %w", err)
	}

	if strings.TrimSpace(response.Error) != "" {
		return nil, fmt.Errorf("SerpApi error: %s", response.Error)
	}

	return &response, nil
}

func appendNewsItem(
	articles []NewsArticle,
	title string,
	description string,
	link string,
	thumbnail string,
	thumbnailSmall string,
	dateText string,
	isoDate string,
	sourceName string,
	scope string,
	country string,
) []NewsArticle {
	title = strings.TrimSpace(title)
	if title == "" {
		return articles
	}

	description = strings.TrimSpace(description)
	link = strings.TrimSpace(link)
	sourceName = strings.TrimSpace(sourceName)

	imageURL := strings.TrimSpace(thumbnail)
	if imageURL == "" {
		imageURL = strings.TrimSpace(thumbnailSmall)
	}

	publishedAt := parseSerpAPIDate(isoDate, dateText)

	return append(articles, NewsArticle{
		Title:       title,
		Description: description,
		Content:     description,
		ImageURL:    imageURL,
		URL:         link,
		SourceName:  sourceName,
		SourceURL:   "",
		PublishedAt: publishedAt,
		Scope:       scope,
		Country:     strings.ToUpper(country),
		Category: detectNewsCategory(
			title + " " + description,
		),
	})
}

func buildNewsQuery(scope string, country string) string {
	freshness := " when:7d"

	if scope == "local" && strings.TrimSpace(country) != "" {
		countryName := strings.ToUpper(country)

		return "(" +
			countryName + " disaster OR " +
			countryName + " flood OR " +
			countryName + " fire OR " +
			countryName + " emergency OR " +
			countryName + " security OR " +
			countryName + " outbreak OR " +
			countryName + " accident OR " +
			countryName + " weather warning" +
			")" + freshness
	}

	return "(disaster OR flood OR fire outbreak OR storm OR emergency OR earthquake OR conflict OR security OR disease outbreak OR public safety OR evacuation OR rescue OR weather warning)" + freshness
}

func parseSerpAPIDate(isoDate string, dateText string) time.Time {
	isoDate = strings.TrimSpace(isoDate)
	if isoDate != "" {
		if parsed, err := time.Parse(time.RFC3339, isoDate); err == nil {
			return parsed.UTC()
		}
	}

	dateText = strings.TrimSpace(dateText)
	if dateText == "" {
		return time.Now().UTC()
	}

	if parsed, err := time.Parse("01/02/2006, 03:04 PM, -0700 MST", dateText); err == nil {
		return parsed.UTC()
	}

	if parsed, err := time.Parse("01/02/2006, 03:04 PM, +0000 UTC", dateText); err == nil {
		return parsed.UTC()
	}

	if parsed, err := time.Parse("01/02/2006, 03:04 PM", dateText); err == nil {
		return parsed.UTC()
	}

	lower := strings.ToLower(dateText)
	now := time.Now().UTC()

	if strings.Contains(lower, "minute") {
		return now.Add(-30 * time.Minute)
	}

	if strings.Contains(lower, "hour") {
		return now.Add(-1 * time.Hour)
	}

	if strings.Contains(lower, "day") {
		return now.Add(-24 * time.Hour)
	}

	if strings.Contains(lower, "week") {
		return now.Add(-7 * 24 * time.Hour)
	}

	if strings.Contains(lower, "month") {
		return now.Add(-30 * 24 * time.Hour)
	}

	return now
}

func filterRecentNews(
	articles []NewsArticle,
	maxAge time.Duration,
) []NewsArticle {
	now := time.Now().UTC()
	filtered := make([]NewsArticle, 0)

	for _, article := range articles {
		if article.PublishedAt.IsZero() {
			continue
		}

		age := now.Sub(article.PublishedAt.UTC())

		if age >= 0 && age <= maxAge {
			filtered = append(filtered, article)
		}
	}

	return filtered
}

func removeDuplicateNews(articles []NewsArticle) []NewsArticle {
	seen := make(map[string]bool)
	unique := make([]NewsArticle, 0)

	for _, article := range articles {
		key := strings.ToLower(strings.TrimSpace(article.URL))
		if key == "" {
			key = strings.ToLower(strings.TrimSpace(article.Title))
		}

		if key == "" {
			continue
		}

		if seen[key] {
			continue
		}

		seen[key] = true
		unique = append(unique, article)
	}

	return unique
}

func detectNewsCategory(value string) string {
	value = strings.ToLower(value)

	switch {
	case strings.Contains(value, "flood") ||
		strings.Contains(value, "rain") ||
		strings.Contains(value, "flooding"):
		return "flood"

	case strings.Contains(value, "fire") ||
		strings.Contains(value, "wildfire") ||
		strings.Contains(value, "explosion"):
		return "fire"

	case strings.Contains(value, "storm") ||
		strings.Contains(value, "weather") ||
		strings.Contains(value, "hurricane") ||
		strings.Contains(value, "typhoon") ||
		strings.Contains(value, "cyclone"):
		return "weather"

	case strings.Contains(value, "disease") ||
		strings.Contains(value, "outbreak") ||
		strings.Contains(value, "cholera") ||
		strings.Contains(value, "malaria") ||
		strings.Contains(value, "epidemic") ||
		strings.Contains(value, "health"):
		return "health"

	case strings.Contains(value, "conflict") ||
		strings.Contains(value, "attack") ||
		strings.Contains(value, "security") ||
		strings.Contains(value, "violence") ||
		strings.Contains(value, "terror") ||
		strings.Contains(value, "war"):
		return "security"

	case strings.Contains(value, "earthquake") ||
		strings.Contains(value, "tremor"):
		return "earthquake"

	case strings.Contains(value, "landslide"):
		return "landslide"

	case strings.Contains(value, "drought") ||
		strings.Contains(value, "heatwave") ||
		strings.Contains(value, "heat wave"):
		return "drought"

	default:
		return "general"
	}
}