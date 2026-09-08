package newsfeed

import (
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
)

const maxQueryLength = 80

func NewHandler(provider Provider) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := strings.TrimSpace(c.Query("q"))
		if utf8.RuneCountInString(query) > maxQueryLength {
			c.JSON(http.StatusBadRequest, gin.H{"error": "search_query_too_long"})
			return
		}
		feed, err := provider.Fetch(c.Request.Context(), query)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "news_provider_unavailable"})
			return
		}
		c.JSON(http.StatusOK, feed)
	}
}
