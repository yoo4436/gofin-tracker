package dailyreport

import (
	"strings"
	"testing"
)

func TestBuildFormatPromptIncludesAllowedSources(t *testing.T) {
	prompt, err := buildFormatPrompt("2026-08-10", "research notes", GroundingMetadata{
		Sources: []GroundingSource{{
			Title:      "Example",
			Publisher:  "example.com",
			URL:        "https://example.com/news",
			ObservedAt: "2026-08-10T00:15:00Z",
		}},
	})
	if err != nil {
		t.Fatalf("buildFormatPrompt returned an error: %v", err)
	}
	if !strings.Contains(prompt, "https://example.com/news") ||
		!strings.Contains(prompt, "example.com") ||
		!strings.Contains(prompt, "2026-08-10T00:15:00Z") ||
		!strings.Contains(prompt, "research notes") {
		t.Fatalf("format prompt is missing grounded inputs: %s", prompt)
	}
}
