package dailyreport

import (
	"fmt"
	"strings"
)

func RenderMarkdown(report Report) string {
	var output strings.Builder
	fmt.Fprintf(&output, "# %s\n\n> %s\n\n", report.Title, report.Summary)

	for index, item := range report.Items {
		fmt.Fprintf(&output, "## %d. [%s] %s\n\n", index+1, item.Category, item.Headline)
		fmt.Fprintf(&output, "### 焦點新聞總覽\n\n%s\n\n", item.Overview)
		fmt.Fprintf(&output, "### 多角度深入分析與長期影響\n\n%s\n\n", item.Analysis)
		output.WriteString("### 資料來源\n\n")
		for _, source := range item.Sources {
			label := escapeMarkdownText(source.Title)
			if label == "" {
				label = escapeMarkdownText(source.Publisher)
			}
			fmt.Fprintf(&output, "- [%s](%s)", label, source.URL)
			if source.Publisher != "" {
				fmt.Fprintf(&output, " — %s", escapeMarkdownText(source.Publisher))
			}
			if source.PublishedAt != "" {
				fmt.Fprintf(&output, "（%s）", source.PublishedAt)
			}
			output.WriteString("\n")
		}
		output.WriteString("\n---\n\n")
	}

	output.WriteString("> 本文為資訊摘要與風險提示，不構成投資建議。\n")
	return output.String()
}

func escapeMarkdownText(value string) string {
	replacer := strings.NewReplacer("[", "\\[", "]", "\\]")
	return replacer.Replace(strings.TrimSpace(value))
}
