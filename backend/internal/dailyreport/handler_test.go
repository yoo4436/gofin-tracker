package dailyreport

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

type fakeGenerateService struct {
	called bool
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
