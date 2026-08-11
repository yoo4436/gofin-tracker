package dailyreport

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"index-system-backend/backend/internal/grounding"
)

var ErrInvalidAIResponse = errors.New("invalid AI daily report response")

type Researcher interface {
	Search(ctx context.Context, windowStart, windowEnd time.Time) (string, grounding.Metadata, error)
}

type Generator interface {
	ModelName() string
	GenerateJSON(ctx context.Context, prompt string, schema map[string]any, destination any) error
}

type Repository interface {
	FindByDate(ctx context.Context, reportDate string) (id int, status string, found bool, err error)
	SaveDraft(ctx context.Context, reportDate, model, promptVersion string, report Report) (id int, created bool, err error)
}

type Service struct {
	researcher Researcher
	generator  Generator
	repository Repository
	location   *time.Location
}

func NewService(researcher Researcher, generator Generator, repository Repository, location *time.Location) *Service {
	return &Service{researcher: researcher, generator: generator, repository: repository, location: location}
}

func (s *Service) Generate(ctx context.Context, options GenerateOptions) (GenerateResult, error) {
	if s.generator == nil {
		return GenerateResult{}, errors.New("daily report generator is not configured")
	}
	if s.researcher == nil {
		return GenerateResult{}, errors.New("daily report researcher is not configured")
	}
	if !options.DryRun && s.repository == nil {
		return GenerateResult{}, errors.New("daily report repository is not configured")
	}

	asOf := options.AsOf
	if asOf.IsZero() {
		asOf = time.Now()
	}
	if s.location != nil {
		asOf = asOf.In(s.location)
	}
	reportDate := asOf.Format("2006-01-02")

	if !options.DryRun {
		id, status, found, err := s.repository.FindByDate(ctx, reportDate)
		if err != nil {
			return GenerateResult{}, fmt.Errorf("find daily report by date: %w", err)
		}
		if found {
			return GenerateResult{
				ID:            id,
				Status:        status,
				ReportDate:    reportDate,
				Model:         s.generator.ModelName(),
				PromptVersion: PromptVersion,
			}, nil
		}
	}

	windowStart := asOf.Add(-24 * time.Hour)
	research, metadata, err := s.researcher.Search(ctx, windowStart, asOf)
	if err != nil {
		return GenerateResult{}, fmt.Errorf("research daily news: %w", err)
	}
	if len(metadata.Sources) == 0 {
		return GenerateResult{}, fmt.Errorf("%w: search returned no grounded sources", ErrInvalidAIResponse)
	}

	formatPrompt, err := buildFormatPrompt(reportDate, research, metadata)
	if err != nil {
		return GenerateResult{}, err
	}
	var report Report
	if err := s.generator.GenerateJSON(ctx, formatPrompt, reportSchema(), &report); err != nil {
		return GenerateResult{}, fmt.Errorf("format daily report: %w", err)
	}
	if err := validateReport(report, reportDate, windowStart, asOf, metadata); err != nil {
		return GenerateResult{}, err
	}

	result := GenerateResult{
		Created:       false,
		Status:        "preview",
		ReportDate:    reportDate,
		Model:         s.generator.ModelName(),
		PromptVersion: PromptVersion,
		Report:        &report,
		Grounding:     metadata,
	}
	if options.DryRun {
		return result, nil
	}

	id, created, err := s.repository.SaveDraft(ctx, reportDate, s.generator.ModelName(), PromptVersion, report)
	if err != nil {
		return GenerateResult{}, fmt.Errorf("save daily report draft: %w", err)
	}
	result.ID = id
	result.Created = created
	result.Status = "draft"
	if !created {
		result.Report = nil
	}
	return result, nil
}

func validateReport(report Report, reportDate string, windowStart, windowEnd time.Time, metadata grounding.Metadata) error {
	allowedCategories := map[string]struct{}{
		"地緣政治": {}, "人工智慧": {}, "財經": {}, "台美股": {}, "宏觀經濟": {}, "科技產業": {},
	}
	allowedSourceURLs := make(map[string]struct{}, len(metadata.Sources))
	for _, source := range metadata.Sources {
		allowedSourceURLs[source.URL] = struct{}{}
	}
	if report.ReportDate != reportDate {
		return fmt.Errorf("%w: report date does not match request", ErrInvalidAIResponse)
	}
	if strings.TrimSpace(report.Title) == "" || utf8.RuneCountInString(report.Title) > 255 {
		return fmt.Errorf("%w: title is missing or too long", ErrInvalidAIResponse)
	}
	if strings.TrimSpace(report.Summary) == "" {
		return fmt.Errorf("%w: summary is missing", ErrInvalidAIResponse)
	}
	if len(report.Items) < 3 || len(report.Items) > 5 {
		return fmt.Errorf("%w: expected 3 to 5 report items", ErrInvalidAIResponse)
	}

	for _, item := range report.Items {
		if strings.TrimSpace(item.Category) == "" || strings.TrimSpace(item.Headline) == "" ||
			strings.TrimSpace(item.Overview) == "" || strings.TrimSpace(item.Analysis) == "" {
			return fmt.Errorf("%w: report item is incomplete", ErrInvalidAIResponse)
		}
		if utf8.RuneCountInString(item.Headline) > 500 {
			return fmt.Errorf("%w: report item headline is too long", ErrInvalidAIResponse)
		}
		if _, allowed := allowedCategories[item.Category]; !allowed {
			return fmt.Errorf("%w: report item category is invalid", ErrInvalidAIResponse)
		}
		if len(item.Sources) == 0 || len(item.Sources) > 5 {
			return fmt.Errorf("%w: every item requires 1 to 5 sources", ErrInvalidAIResponse)
		}

		seenURLs := make(map[string]struct{}, len(item.Sources))
		for _, source := range item.Sources {
			if strings.TrimSpace(source.Title) == "" || strings.TrimSpace(source.Publisher) == "" {
				return fmt.Errorf("%w: source title and publisher are required", ErrInvalidAIResponse)
			}
			parsedURL, err := url.ParseRequestURI(source.URL)
			if err != nil || parsedURL.Host == "" || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
				return fmt.Errorf("%w: source URL is invalid", ErrInvalidAIResponse)
			}
			if _, allowed := allowedSourceURLs[source.URL]; !allowed {
				return fmt.Errorf("%w: source URL was not returned by grounded search", ErrInvalidAIResponse)
			}
			if _, exists := seenURLs[source.URL]; exists {
				return fmt.Errorf("%w: duplicate source URL in report item", ErrInvalidAIResponse)
			}
			seenURLs[source.URL] = struct{}{}

			if source.PublishedAt != "" {
				publishedAt, err := time.Parse(time.RFC3339, source.PublishedAt)
				if err != nil {
					return fmt.Errorf("%w: source publication time is invalid", ErrInvalidAIResponse)
				}
				if publishedAt.Before(windowStart.Add(-5*time.Minute)) || publishedAt.After(windowEnd.Add(5*time.Minute)) {
					return fmt.Errorf("%w: source is outside the requested 24-hour window", ErrInvalidAIResponse)
				}
			}
		}
	}

	return nil
}
