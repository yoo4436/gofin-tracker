package dailyreport

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) FindByDate(ctx context.Context, reportDate string) (int, string, bool, error) {
	var id int
	var status string
	err := r.db.QueryRowContext(ctx, `
		SELECT id, status
		FROM daily_reports
		WHERE report_date = $1
	`, reportDate).Scan(&id, &status)
	if err == sql.ErrNoRows {
		return 0, "", false, nil
	}
	if err != nil {
		return 0, "", false, err
	}
	return id, status, true, nil
}

func (r *SQLRepository) SaveDraft(ctx context.Context, reportDate, model, promptVersion string, report Report) (int, bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, false, err
	}
	defer tx.Rollback()

	var reportID int
	err = tx.QueryRowContext(ctx, `
		INSERT INTO daily_reports (
			title, summary, content, is_premium, status, published_at,
			report_date, ai_model, prompt_version
		)
		VALUES ($1, $2, $3, false, 'draft', NULL, $4, $5, $6)
		ON CONFLICT (report_date) WHERE report_date IS NOT NULL DO NOTHING
		RETURNING id
	`, report.Title, report.Summary, RenderMarkdown(report), reportDate, model, promptVersion).Scan(&reportID)
	if err == sql.ErrNoRows {
		if err := tx.QueryRowContext(ctx, `SELECT id FROM daily_reports WHERE report_date = $1`, reportDate).Scan(&reportID); err != nil {
			return 0, false, fmt.Errorf("find concurrently created report: %w", err)
		}
		return reportID, false, nil
	}
	if err != nil {
		return 0, false, err
	}

	for itemIndex, item := range report.Items {
		var itemID int
		body := "### 焦點新聞總覽\n\n" + item.Overview + "\n\n### 多角度深入分析與長期影響\n\n" + item.Analysis
		if err := tx.QueryRowContext(ctx, `
			INSERT INTO daily_report_items (report_id, position, headline, body)
			VALUES ($1, $2, $3, $4)
			RETURNING id
		`, reportID, itemIndex+1, item.Headline, body).Scan(&itemID); err != nil {
			return 0, false, err
		}

		for sourceIndex, source := range item.Sources {
			var publishedAt any
			if source.PublishedAt != "" {
				parsed, err := time.Parse(time.RFC3339, source.PublishedAt)
				if err != nil {
					return 0, false, err
				}
				publishedAt = parsed
			}

			var sourceID int
			if err := tx.QueryRowContext(ctx, `
				INSERT INTO news_sources (url, title, publisher_name, source_published_at)
				VALUES ($1, NULLIF($2, ''), NULLIF($3, ''), $4)
				ON CONFLICT (url) DO UPDATE SET
					title = COALESCE(EXCLUDED.title, news_sources.title),
					publisher_name = COALESCE(EXCLUDED.publisher_name, news_sources.publisher_name),
					source_published_at = COALESCE(EXCLUDED.source_published_at, news_sources.source_published_at)
				RETURNING id
			`, source.URL, source.Title, source.Publisher, publishedAt).Scan(&sourceID); err != nil {
				return 0, false, err
			}

			if _, err := tx.ExecContext(ctx, `
				INSERT INTO report_item_sources (report_item_id, source_id, position)
				VALUES ($1, $2, $3)
			`, itemID, sourceID, sourceIndex+1); err != nil {
				return 0, false, err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, false, err
	}
	return reportID, true, nil
}
