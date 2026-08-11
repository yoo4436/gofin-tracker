BEGIN;

ALTER TABLE public.daily_reports
  ADD COLUMN report_date date,
  ADD COLUMN ai_model text,
  ADD COLUMN prompt_version text;

-- Preserve the calendar date of reports that were published before AI generation existed.
UPDATE public.daily_reports
SET report_date = (published_at AT TIME ZONE 'Asia/Taipei')::date
WHERE report_date IS NULL
  AND published_at IS NOT NULL;

CREATE UNIQUE INDEX idx_daily_reports_generated_date
  ON public.daily_reports (report_date)
  WHERE report_date IS NOT NULL;

COMMIT;
