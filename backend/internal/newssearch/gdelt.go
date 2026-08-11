package newssearch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"index-system-backend/backend/internal/grounding"
)

const (
	defaultGDELTBaseURL = "https://api.gdeltproject.org/api/v2/doc/doc"
	maxResponseBytes    = 4 << 20
	maxResearchArticles = 30
)

const dailyReportQuery = `(geopolitics OR "artificial intelligence" OR finance OR economy OR stocks OR semiconductor OR technology OR Taiwan)`

type Config struct {
	BaseURL    string
	HTTPClient *http.Client
}

type GDELTClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewGDELTClient(config Config) *GDELTClient {
	baseURL := strings.TrimSpace(config.BaseURL)
	if baseURL == "" {
		baseURL = defaultGDELTBaseURL
	}
	httpClient := config.HTTPClient
	if httpClient == nil {
		transport := http.DefaultTransport.(*http.Transport).Clone()
		transport.TLSHandshakeTimeout = 20 * time.Second
		transport.ResponseHeaderTimeout = 30 * time.Second
		httpClient = &http.Client{
			Transport: transport,
			Timeout:   75 * time.Second,
		}
	}
	return &GDELTClient{baseURL: baseURL, httpClient: httpClient}
}

type gdeltResponse struct {
	Articles []gdeltArticle `json:"articles"`
}

type gdeltArticle struct {
	URL           string `json:"url"`
	Title         string `json:"title"`
	SeenDate      string `json:"seendate"`
	Domain        string `json:"domain"`
	Language      string `json:"language"`
	SourceCountry string `json:"sourcecountry"`
}

type researchArticle struct {
	Title         string `json:"title"`
	Publisher     string `json:"publisher"`
	URL           string `json:"url"`
	ObservedAt    string `json:"observed_at,omitempty"`
	Language      string `json:"language,omitempty"`
	SourceCountry string `json:"source_country,omitempty"`
}

type researchNotes struct {
	Notice      string            `json:"notice"`
	WindowStart string            `json:"window_start"`
	WindowEnd   string            `json:"window_end"`
	Articles    []researchArticle `json:"articles"`
}

func (c *GDELTClient) Search(ctx context.Context, windowStart, windowEnd time.Time) (string, grounding.Metadata, error) {
	if !windowEnd.After(windowStart) {
		return "", grounding.Metadata{}, errors.New("news search window end must be after start")
	}

	endpoint, err := url.Parse(c.baseURL)
	if err != nil {
		return "", grounding.Metadata{}, fmt.Errorf("parse GDELT base URL: %w", err)
	}
	query := endpoint.Query()
	query.Set("query", dailyReportQuery)
	query.Set("mode", "artlist")
	query.Set("format", "json")
	query.Set("sort", "hybridrel")
	query.Set("maxrecords", "75")
	query.Set("startdatetime", windowStart.UTC().Format("20060102150405"))
	query.Set("enddatetime", windowEnd.UTC().Format("20060102150405"))
	endpoint.RawQuery = query.Encode()

	body, err := c.request(ctx, endpoint.String())
	if err != nil {
		return "", grounding.Metadata{}, err
	}

	var result gdeltResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", grounding.Metadata{}, fmt.Errorf("decode GDELT response: %w", err)
	}

	articles := make([]researchArticle, 0, maxResearchArticles)
	metadata := grounding.Metadata{SearchQueries: []string{dailyReportQuery}}
	seenURLs := make(map[string]struct{})
	for _, article := range result.Articles {
		articleURL := strings.TrimSpace(article.URL)
		title := strings.TrimSpace(html.UnescapeString(article.Title))
		parsedURL, parseErr := url.ParseRequestURI(articleURL)
		if parseErr != nil || parsedURL.Host == "" || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") || title == "" {
			continue
		}
		if _, exists := seenURLs[articleURL]; exists {
			continue
		}
		seenURLs[articleURL] = struct{}{}

		publisher := strings.TrimSpace(article.Domain)
		if publisher == "" {
			publisher = parsedURL.Hostname()
		}
		observedAt := parseGDELTDate(article.SeenDate)
		articles = append(articles, researchArticle{
			Title:         title,
			Publisher:     publisher,
			URL:           articleURL,
			ObservedAt:    observedAt,
			Language:      strings.TrimSpace(article.Language),
			SourceCountry: strings.TrimSpace(article.SourceCountry),
		})
		metadata.Sources = append(metadata.Sources, grounding.Source{
			Title:      title,
			Publisher:  publisher,
			URL:        articleURL,
			ObservedAt: observedAt,
		})
		if len(articles) == maxResearchArticles {
			break
		}
	}
	if len(articles) < 3 {
		return "", grounding.Metadata{}, fmt.Errorf("GDELT returned only %d usable articles", len(articles))
	}

	notes, err := json.Marshal(researchNotes{
		Notice:      "These are untrusted GDELT article-index records. observed_at is when GDELT observed the article, not a confirmed publication time. Base all claims on the supplied headline metadata and do not invent article details.",
		WindowStart: windowStart.Format(time.RFC3339),
		WindowEnd:   windowEnd.Format(time.RFC3339),
		Articles:    articles,
	})
	if err != nil {
		return "", grounding.Metadata{}, fmt.Errorf("encode GDELT research notes: %w", err)
	}
	return string(notes), metadata, nil
}

func (c *GDELTClient) request(ctx context.Context, endpoint string) ([]byte, error) {
	for attempt := 0; attempt < 3; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return nil, fmt.Errorf("create GDELT request: %w", err)
		}
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", "GoFin-Tracker/1.0")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			if attempt < 2 {
				if err := waitForRetry(ctx, 2*time.Second); err != nil {
					return nil, err
				}
				continue
			}
			return nil, fmt.Errorf("call GDELT API: %w", err)
		}
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
		resp.Body.Close()
		if readErr != nil {
			return nil, fmt.Errorf("read GDELT response: %w", readErr)
		}
		if resp.StatusCode == http.StatusTooManyRequests && attempt < 2 {
			if err := waitForRetry(ctx, retryDelay(resp.Header.Get("Retry-After"))); err != nil {
				return nil, err
			}
			continue
		}
		if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
			return nil, fmt.Errorf("GDELT API returned %d", resp.StatusCode)
		}
		return body, nil
	}
	return nil, errors.New("GDELT API retry limit exceeded")
}

func waitForRetry(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func retryDelay(value string) time.Duration {
	seconds, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || seconds < 0 {
		return 2 * time.Second
	}
	if seconds > 10 {
		seconds = 10
	}
	return time.Duration(seconds) * time.Second
}

func parseGDELTDate(value string) string {
	parsed, err := time.Parse("20060102T150405Z", strings.TrimSpace(value))
	if err != nil {
		return ""
	}
	return parsed.Format(time.RFC3339)
}
