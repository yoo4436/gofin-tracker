package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCORSMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name            string
		origin          string
		method          string
		wantStatus      int
		wantAllowOrigin string
	}{
		{name: "allowed origin", origin: "https://app.example.com", method: http.MethodGet, wantStatus: http.StatusOK, wantAllowOrigin: "https://app.example.com"},
		{name: "allowed preflight", origin: "http://localhost:5173", method: http.MethodOptions, wantStatus: http.StatusNoContent, wantAllowOrigin: "http://localhost:5173"},
		{name: "rejected origin", origin: "https://evil.example.com", method: http.MethodGet, wantStatus: http.StatusForbidden},
		{name: "server to server", method: http.MethodGet, wantStatus: http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(corsMiddleware("https://app.example.com,http://localhost:5173"))
			router.Any("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

			request := httptest.NewRequest(tt.method, "/test", nil)
			if tt.origin != "" {
				request.Header.Set("Origin", tt.origin)
			}
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			if response.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, tt.wantStatus)
			}
			if got := response.Header().Get("Access-Control-Allow-Origin"); got != tt.wantAllowOrigin {
				t.Fatalf("Access-Control-Allow-Origin = %q, want %q", got, tt.wantAllowOrigin)
			}
		})
	}
}
