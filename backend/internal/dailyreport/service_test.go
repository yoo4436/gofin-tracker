package dailyreport

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

type fakeGenerator struct {
	formatCalled bool
	report       Report
}

type fakeResearcher struct {
	searchCalled bool
}

func TestValidateReportRejectsSourceOutsideGroundingMetadata(t *testing.T) {
	report := validReport()
	report.Items[0].Sources[0].URL = "https://untrusted.example/news"
	windowEnd := time.Date(2026, 8, 10, 8, 30, 0, 0, time.FixedZone("Asia/Taipei", 8*60*60))

	err := validateReport(report, "2026-08-10", windowEnd.Add(-24*time.Hour), windowEnd, GroundingMetadata{
		Sources: []GroundingSource{{Title: "來源", URL: "https://example.com/news"}},
	})
	if !errors.Is(err, ErrInvalidAIResponse) {
		t.Fatalf("expected ErrInvalidAIResponse, got %v", err)
	}
}

func (f *fakeGenerator) ModelName() string { return "test-model" }

func (f *fakeResearcher) Search(_ context.Context, _, _ time.Time) (string, GroundingMetadata, error) {
	f.searchCalled = true
	return "research notes", GroundingMetadata{
		SearchQueries: []string{"test query"},
		Sources:       []GroundingSource{{Title: "來源", URL: "https://example.com/news"}},
	}, nil
}

func (f *fakeGenerator) GenerateJSON(_ context.Context, _ string, _ map[string]any, destination any) error {
	f.formatCalled = true
	encoded, _ := json.Marshal(f.report)
	return json.Unmarshal(encoded, destination)
}

type fakeRepository struct {
	found  bool
	saved  bool
	id     int
	status string
}

func (r *fakeRepository) FindByDate(context.Context, string) (int, string, bool, error) {
	return r.id, r.status, r.found, nil
}

func (r *fakeRepository) SaveDraft(_ context.Context, _ string, _, _ string, _ Report) (int, bool, error) {
	r.saved = true
	return 42, true, nil
}

func TestGenerateDryRun(t *testing.T) {
	asOf := time.Date(2026, 8, 10, 8, 30, 0, 0, time.FixedZone("Asia/Taipei", 8*60*60))
	researcher := &fakeResearcher{}
	generator := &fakeGenerator{report: validReport()}
	repository := &fakeRepository{}
	service := NewService(researcher, generator, repository, asOf.Location())

	result, err := service.Generate(context.Background(), GenerateOptions{AsOf: asOf, DryRun: true})
	if err != nil {
		t.Fatalf("Generate returned an error: %v", err)
	}
	if result.Status != "preview" || result.Report == nil {
		t.Fatalf("unexpected result: %#v", result)
	}
	if repository.saved {
		t.Fatal("dry run must not save a draft")
	}
}

func TestGenerateReturnsExistingReportWithoutCallingGemini(t *testing.T) {
	generator := &fakeGenerator{report: validReport()}
	researcher := &fakeResearcher{}
	repository := &fakeRepository{found: true, id: 9, status: "draft"}
	service := NewService(researcher, generator, repository, time.UTC)

	result, err := service.Generate(context.Background(), GenerateOptions{AsOf: time.Date(2026, 8, 10, 0, 30, 0, 0, time.UTC)})
	if err != nil {
		t.Fatalf("Generate returned an error: %v", err)
	}
	if result.ID != 9 || result.Status != "draft" {
		t.Fatalf("unexpected result: %#v", result)
	}
	if researcher.searchCalled || generator.formatCalled {
		t.Fatal("news search and Gemini must not be called when the date already exists")
	}
}

func validReport() Report {
	items := make([]Item, 3)
	for i := range items {
		items[i] = Item{
			Category: "人工智慧",
			Headline: "測試新聞",
			Overview: "已證實內容",
			Analysis: "分析內容",
			Sources: []Source{{
				Title:       "來源",
				Publisher:   "Publisher",
				URL:         "https://example.com/news",
				PublishedAt: "2026-08-10T07:00:00+08:00",
			}},
		}
	}
	return Report{
		ReportDate: "2026-08-10",
		Title:      "每日新聞深度分析",
		Summary:    "今日摘要內容。",
		Items:      items,
	}
}
