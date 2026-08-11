package dailyreport

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const adminTokenHeader = "X-Daily-Report-Token"

type GenerateService interface {
	Generate(ctx context.Context, options GenerateOptions) (GenerateResult, error)
}

type generateRequest struct {
	AsOf   string `json:"as_of"`
	DryRun bool   `json:"dry_run"`
}

func NewGenerateHandler(service GenerateService, adminToken string) gin.HandlerFunc {
	adminToken = strings.TrimSpace(adminToken)

	return func(c *gin.Context) {
		if service == nil || adminToken == "" {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "daily_report_service_unavailable"})
			return
		}
		if subtle.ConstantTimeCompare([]byte(c.GetHeader(adminTokenHeader)), []byte(adminToken)) != 1 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		var request generateRequest
		decoder := json.NewDecoder(http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&request); err != nil && err != io.EOF {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "body must be valid JSON"})
			return
		}
		if err := decoder.Decode(&struct{}{}); err != io.EOF {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "body must contain one JSON object"})
			return
		}

		var asOf time.Time
		if request.AsOf != "" {
			parsed, err := time.Parse(time.RFC3339, request.AsOf)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "as_of must use RFC3339 format"})
				return
			}
			asOf = parsed
		}

		result, err := service.Generate(c.Request.Context(), GenerateOptions{AsOf: asOf, DryRun: request.DryRun})
		if err != nil {
			log.Printf("generate daily report: %v", err)
			status := http.StatusBadGateway
			code := "daily_report_generation_failed"
			if errors.Is(err, ErrInvalidAIResponse) {
				code = "invalid_ai_response"
			}
			c.JSON(status, gin.H{"error": code})
			return
		}

		status := http.StatusOK
		if result.Created {
			status = http.StatusCreated
		}
		c.JSON(status, result)
	}
}
