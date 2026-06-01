package handlers

import (
	"net/http"
	"strings"

	"disaster_alert_backend/internal/services"
	"disaster_alert_backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type NewsHandler struct {
	newsService *services.NewsService
}

func NewNewsHandler(newsService *services.NewsService) *NewsHandler {
	return &NewsHandler{
		newsService: newsService,
	}
}

func (h *NewsHandler) GetNews(c *gin.Context) {
	scope := strings.ToLower(strings.TrimSpace(c.DefaultQuery("scope", "local")))
	country := strings.TrimSpace(c.DefaultQuery("country", ""))

	if scope == "" {
		scope = "local"
	}

	if scope != "local" &&
		scope != "world" &&
		scope != "global" &&
		scope != "worldwide" {
		utils.ErrorResponse(
			c,
			http.StatusBadRequest,
			"Invalid news scope. Use local or world",
			nil,
		)
		return
	}

	articles, err := h.newsService.FetchNews(scope, country)
	if err != nil {
		utils.ErrorResponse(
			c,
			http.StatusInternalServerError,
			"Failed to fetch news: "+err.Error(),
			err,
		)
		return
	}

	utils.SuccessResponse(
		c,
		http.StatusOK,
		"News fetched successfully",
		articles,
	)
}