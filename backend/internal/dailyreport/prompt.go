package dailyreport

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"index-system-backend/backend/internal/grounding"
	"text/template"
)

//go:embed prompts/format_report.tmpl
var formatPromptSource string

var formatPromptTemplate = template.Must(template.New("daily-report-format").Parse(formatPromptSource))

type formatPromptData struct {
	ReportDate  string
	Research    string
	SourcesJSON string
}

func buildFormatPrompt(reportDate, research string, metadata grounding.Metadata) (string, error) {
	sourcesJSON, err := json.Marshal(metadata.Sources)
	if err != nil {
		return "", fmt.Errorf("encode grounded sources: %w", err)
	}
	data := formatPromptData{
		ReportDate:  reportDate,
		Research:    research,
		SourcesJSON: string(sourcesJSON),
	}
	return executePrompt(formatPromptTemplate, data)
}

func executePrompt(prompt *template.Template, data any) (string, error) {
	var output bytes.Buffer
	if err := prompt.Execute(&output, data); err != nil {
		return "", fmt.Errorf("render daily report prompt: %w", err)
	}
	return output.String(), nil
}
