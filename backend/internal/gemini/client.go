package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultBaseURL   = "https://generativelanguage.googleapis.com/v1beta"
	defaultModel     = "gemini-3.6-flash"
	maxResponseBytes = 4 << 20
)

type Config struct {
	APIKey     string
	Model      string
	BaseURL    string
	HTTPClient *http.Client
}

type Client struct {
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
}

func NewClient(config Config) *Client {
	model := strings.TrimSpace(config.Model)
	if model == "" {
		model = defaultModel
	}

	baseURL := strings.TrimRight(strings.TrimSpace(config.BaseURL), "/")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 2 * time.Minute}
	}

	return &Client{
		apiKey:     strings.TrimSpace(config.APIKey),
		model:      model,
		baseURL:    baseURL,
		httpClient: httpClient,
	}
}

func (c *Client) ModelName() string {
	return c.model
}

type generateRequest struct {
	Contents         []content      `json:"contents"`
	GenerationConfig map[string]any `json:"generationConfig,omitempty"`
}

type content struct {
	Role  string `json:"role,omitempty"`
	Parts []part `json:"parts"`
}

type part struct {
	Text string `json:"text,omitempty"`
}

type generateResponse struct {
	Candidates []struct {
		Content      content `json:"content"`
		FinishReason string  `json:"finishReason"`
	} `json:"candidates"`
	PromptFeedback struct {
		BlockReason string `json:"blockReason"`
	} `json:"promptFeedback"`
}

type apiErrorResponse struct {
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

// GenerateJSON performs the formatting stage without tools and decodes JSON
// that conforms to the supplied schema.
func (c *Client) GenerateJSON(ctx context.Context, prompt string, schema map[string]any, destination any) error {
	if destination == nil {
		return errors.New("destination must not be nil")
	}
	payload := generateRequest{
		Contents: []content{{Role: "user", Parts: []part{{Text: prompt}}}},
		GenerationConfig: map[string]any{
			"responseMimeType":   "application/json",
			"responseJsonSchema": schema,
		},
	}

	result, err := c.generate(ctx, payload)
	if err != nil {
		return err
	}
	text := responseText(result)
	if text == "" {
		return errors.New("Gemini formatting returned empty content")
	}
	if err := json.Unmarshal([]byte(text), destination); err != nil {
		return fmt.Errorf("decode Gemini structured output: %w", err)
	}
	return nil
}

func (c *Client) generate(ctx context.Context, payload generateRequest) (generateResponse, error) {
	if c.apiKey == "" {
		return generateResponse{}, errors.New("Gemini API key is not configured")
	}
	if len(payload.Contents) == 0 || len(payload.Contents[0].Parts) == 0 || strings.TrimSpace(payload.Contents[0].Parts[0].Text) == "" {
		return generateResponse{}, errors.New("prompt must not be empty")
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return generateResponse{}, fmt.Errorf("encode Gemini request: %w", err)
	}

	endpoint := fmt.Sprintf("%s/models/%s:generateContent", c.baseURL, url.PathEscape(c.model))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return generateResponse{}, fmt.Errorf("create Gemini request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return generateResponse{}, fmt.Errorf("call Gemini API: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return generateResponse{}, fmt.Errorf("read Gemini response: %w", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		var apiError apiErrorResponse
		if json.Unmarshal(responseBody, &apiError) == nil && apiError.Error.Message != "" {
			return generateResponse{}, fmt.Errorf("Gemini API returned %d: %s", resp.StatusCode, apiError.Error.Message)
		}
		return generateResponse{}, fmt.Errorf("Gemini API returned %d", resp.StatusCode)
	}

	var result generateResponse
	if err := json.Unmarshal(responseBody, &result); err != nil {
		return generateResponse{}, fmt.Errorf("decode Gemini response: %w", err)
	}
	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		if result.PromptFeedback.BlockReason != "" {
			return generateResponse{}, fmt.Errorf("Gemini blocked the request: %s", result.PromptFeedback.BlockReason)
		}
		return generateResponse{}, errors.New("Gemini returned no content")
	}
	return result, nil
}

func responseText(result generateResponse) string {
	var output strings.Builder
	for _, responsePart := range result.Candidates[0].Content.Parts {
		output.WriteString(responsePart.Text)
	}
	return output.String()
}
