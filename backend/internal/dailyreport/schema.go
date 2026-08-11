package dailyreport

func reportSchema() map[string]any {
	stringProperty := func(description string) map[string]any {
		return map[string]any{"type": "string", "description": description}
	}

	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"report_date": stringProperty("Report date in YYYY-MM-DD format."),
			"title":       stringProperty("Concise Traditional Chinese daily report title."),
			"summary":     stringProperty("Two to four sentence Traditional Chinese overview."),
			"items": map[string]any{
				"type":     "array",
				"minItems": 3,
				"maxItems": 5,
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"category": map[string]any{
							"type": "string",
							"enum": []string{"地緣政治", "人工智慧", "財經", "台美股", "宏觀經濟", "科技產業"},
						},
						"headline": stringProperty("Factual Traditional Chinese headline."),
						"overview": stringProperty("Verified facts and key developments in Traditional Chinese."),
						"analysis": stringProperty("Multi-angle impact analysis with inferences clearly distinguished from facts."),
						"sources": map[string]any{
							"type":     "array",
							"minItems": 1,
							"maxItems": 5,
							"items": map[string]any{
								"type": "object",
								"properties": map[string]any{
									"title":        stringProperty("Source article or document title."),
									"publisher":    stringProperty("Publisher or organization name."),
									"url":          stringProperty("Clickable source URL used for this item."),
									"published_at": stringProperty("Source publication time in RFC3339, or empty if unavailable."),
								},
								"required": []string{"title", "publisher", "url", "published_at"},
							},
						},
					},
					"required": []string{"category", "headline", "overview", "analysis", "sources"},
				},
			},
		},
		"required": []string{"report_date", "title", "summary", "items"},
	}
}
