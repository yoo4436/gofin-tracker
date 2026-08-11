package dailyreport

import (
	"time"

	"index-system-backend/backend/internal/grounding"
)

const PromptVersion = "2026-08-10-v3-gdelt"

type Source struct {
	Title       string `json:"title"`
	Publisher   string `json:"publisher"`
	URL         string `json:"url"`
	PublishedAt string `json:"published_at"`
}

type Item struct {
	Category string   `json:"category"`
	Headline string   `json:"headline"`
	Overview string   `json:"overview"`
	Analysis string   `json:"analysis"`
	Sources  []Source `json:"sources"`
}

type Report struct {
	ReportDate string `json:"report_date"`
	Title      string `json:"title"`
	Summary    string `json:"summary"`
	Items      []Item `json:"items"`
}

type GenerateOptions struct {
	AsOf   time.Time
	DryRun bool
}

type GroundingMetadata = grounding.Metadata
type GroundingSource = grounding.Source

type GenerateResult struct {
	ID            int               `json:"id,omitempty"`
	Created       bool              `json:"created"`
	Status        string            `json:"status"`
	ReportDate    string            `json:"report_date"`
	Model         string            `json:"model"`
	PromptVersion string            `json:"prompt_version"`
	Report        *Report           `json:"report,omitempty"`
	Grounding     GroundingMetadata `json:"grounding"`
}

type PublishResult struct {
	ID          int       `json:"id"`
	Status      string    `json:"status"`
	PublishedAt time.Time `json:"published_at"`
}
