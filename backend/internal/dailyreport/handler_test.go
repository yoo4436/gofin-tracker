package dailyreport

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

type fakeGenerateService struct {
	called bool
}

type fakePublishService struct {
	called   bool
	reportID int
	result   PublishResult
	err      error
}

func (f *fakePublishService) Publish(_ context.Context, reportID int) (PublishResult, error) {
	f.called = true
	f.reportID = reportID
	return f.result, f.err
}

func (f *fakeGenerateService) Generate(context.Context, GenerateOptions) (GenerateResult, error) {
	f.called = true
	return GenerateResult{Status: "preview", ReportDate: "2026-08-10"}, nil
}

func TestGenerateHandlerRequiresToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &fakeGenerateService{}
	handler := NewGenerateHandler(service, "secret")

	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
	response := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(response)
	context.Request = request
	handler(context)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("unexpected status %d", response.Code)
	}
	if service.called {
		t.Fatal("service must not be called for unauthorized request")
	}
}

func TestGenerateHandlerAcceptsDryRun(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &fakeGenerateService{}
	handler := NewGenerateHandler(service, "secret")

	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"dry_run":true}`))
	request.Header.Set(adminTokenHeader, "secret")
	response := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(response)
	context.Request = request
	handler(context)

	if response.Code != http.StatusOK {
		t.Fatalf("unexpected status %d: %s", response.Code, response.Body.String())
	}
	if !service.called {
		t.Fatal("service was not called")
	}
}

func TestPublishHandlerRequiresToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &fakePublishService{}
	handler := NewPublishHandler(service, "secret")

	request := httptest.NewRequest(http.MethodPost, "/", nil)
	response := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(response)
	context.Request = request
	context.Params = gin.Params{{Key: "id", Value: "42"}}
	handler(context)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("unexpected status %d", response.Code)
	}
	if service.called {
		t.Fatal("service must not be called for unauthorized request")
	}
}

func TestPublishHandlerRejectsInvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &fakePublishService{}
	handler := NewPublishHandler(service, "secret")

	request := httptest.NewRequest(http.MethodPost, "/", nil)
	request.Header.Set(adminTokenHeader, "secret")
	response := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(response)
	context.Request = request
	context.Params = gin.Params{{Key: "id", Value: "invalid"}}
	handler(context)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status %d", response.Code)
	}
	if service.called {
		t.Fatal("service must not be called for an invalid report ID")
	}
}

func TestPublishHandlerPublishesDraft(t *testing.T) {
	gin.SetMode(gin.TestMode)
	publishedAt := time.Date(2026, 8, 11, 9, 0, 0, 0, time.FixedZone("Asia/Taipei", 8*60*60))
	service := &fakePublishService{result: PublishResult{ID: 42, Status: "published", PublishedAt: publishedAt}}
	handler := NewPublishHandler(service, "secret")

	request := httptest.NewRequest(http.MethodPost, "/", nil)
	request.Header.Set(adminTokenHeader, "secret")
	response := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(response)
	context.Request = request
	context.Params = gin.Params{{Key: "id", Value: "42"}}
	handler(context)

	if response.Code != http.StatusOK {
		t.Fatalf("unexpected status %d: %s", response.Code, response.Body.String())
	}
	if !service.called || service.reportID != 42 {
		t.Fatalf("unexpected publish call: called=%v reportID=%d", service.called, service.reportID)
	}
	if !strings.Contains(response.Body.String(), `"status":"published"`) {
		t.Fatalf("unexpected response: %s", response.Body.String())
	}
}

func TestPublishHandlerMapsLifecycleErrors(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{name: "not found", err: ErrDailyReportNotFound, wantStatus: http.StatusNotFound, wantCode: "daily_report_not_found"},
		{name: "not draft", err: ErrDailyReportNotDraft, wantStatus: http.StatusConflict, wantCode: "daily_report_not_draft"},
		{name: "not publishable", err: ErrDailyReportNotPublishable, wantStatus: http.StatusUnprocessableEntity, wantCode: "daily_report_not_publishable"},
		{name: "database failure", err: errors.New("database unavailable"), wantStatus: http.StatusInternalServerError, wantCode: "daily_report_publish_failed"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			service := &fakePublishService{err: test.err}
			handler := NewPublishHandler(service, "secret")

			request := httptest.NewRequest(http.MethodPost, "/", nil)
			request.Header.Set(adminTokenHeader, "secret")
			response := httptest.NewRecorder()
			context, _ := gin.CreateTestContext(response)
			context.Request = request
			context.Params = gin.Params{{Key: "id", Value: "42"}}
			handler(context)

			if response.Code != test.wantStatus {
				t.Fatalf("unexpected status %d: %s", response.Code, response.Body.String())
			}
			if !strings.Contains(response.Body.String(), test.wantCode) {
				t.Fatalf("unexpected response: %s", response.Body.String())
			}
		})
	}
}
