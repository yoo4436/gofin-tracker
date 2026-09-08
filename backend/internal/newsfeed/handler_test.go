package newsfeed

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

type providerStub struct {
	feed Feed
	err  error
}

func (p providerStub) Fetch(_ context.Context, query string) (Feed, error) {
	p.feed.Query = query
	return p.feed, p.err
}

func TestHandlerReturnsFeed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/news", NewHandler(providerStub{feed: Feed{Articles: []Article{}}}))

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/news?q=AI", nil))
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"query":"AI"`) {
		t.Fatalf("unexpected response: %d %s", response.Code, response.Body.String())
	}
}

func TestHandlerValidatesQueryAndHidesProviderError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/news", NewHandler(providerStub{err: errors.New("upstream detail")}))

	tooLong := httptest.NewRecorder()
	router.ServeHTTP(tooLong, httptest.NewRequest(http.MethodGet, "/news?q="+strings.Repeat("a", maxQueryLength+1), nil))
	if tooLong.Code != http.StatusBadRequest {
		t.Fatalf("long query status = %d, want %d", tooLong.Code, http.StatusBadRequest)
	}

	failure := httptest.NewRecorder()
	router.ServeHTTP(failure, httptest.NewRequest(http.MethodGet, "/news?q=AI", nil))
	if failure.Code != http.StatusBadGateway || strings.Contains(failure.Body.String(), "upstream detail") {
		t.Fatalf("unexpected provider failure response: %d %s", failure.Code, failure.Body.String())
	}
}
