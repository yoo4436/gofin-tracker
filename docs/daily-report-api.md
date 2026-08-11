# AI 日報產生 API

此端點供後端排程或管理工具呼叫。它先透過免費的 GDELT DOC API 搜尋指定時間點往前 24 小時的新聞索引，再使用 Gemini 將資料轉換成固定 JSON 格式。草稿不會自動發布，也不使用付費的 Gemini Google Search grounding。

## 前置設定

1. 套用 `supabase/migrations/20260810083000_ai_daily_report_generation.sql`。
2. 在本機 `.env` 或 Render 環境變數設定：

```env
GEMINI_API_KEY="your-key"
GEMINI_MODEL="gemini-3.6-flash"
GDELT_BASE_URL="https://api.gdeltproject.org/api/v2/doc/doc"
DAILY_REPORT_API_TOKEN="a-long-random-token"
```

`GEMINI_API_KEY` 與 `DAILY_REPORT_API_TOKEN` 只能存在後端，不可加入 Vue 前端。

每次新日期只會產生一次 Gemini 文字模型呼叫。GDELT 不需要 API Key；`GDELT_BASE_URL` 可省略。Gemini 只能使用 GDELT 回傳的來源網址，若模型輸出其他網址，後端會拒絕儲存。

## 預覽測試

`dry_run` 不會寫入資料庫，適合確認提示詞與 Gemini 回傳格式：

```powershell
$headers = @{ "X-Daily-Report-Token" = "your-token" }
$body = @{
  dry_run = $true
  as_of = "2026-08-10T08:30:00+08:00"
} | ConvertTo-Json

Invoke-RestMethod `
  -Method Post `
  -Uri "http://localhost:8080/api/v1/admin/daily-reports/generate" `
  -Headers $headers `
  -ContentType "application/json" `
  -Body $body
```

`as_of` 可省略；省略時使用伺服器當下時間，並轉換為 `Asia/Taipei`。

## 建立草稿

```powershell
Invoke-RestMethod `
  -Method Post `
  -Uri "http://localhost:8080/api/v1/admin/daily-reports/generate" `
  -Headers @{ "X-Daily-Report-Token" = "your-token" } `
  -ContentType "application/json" `
  -Body "{}"
```

成功時會建立 `daily_reports.status = 'draft'`，並寫入 `daily_report_items`、`news_sources`、`report_item_sources`。同一個 `report_date` 重複呼叫時會回傳既有日報，不會重複建立或再次呼叫 Gemini。

## 發布草稿

發布端點使用相同的後端 Token，並透過資料庫的 `publish_daily_report()` 函式驗證內容、新聞項目與來源。成功後才會設定 `status = 'published'` 與 `published_at`。

```powershell
$reportID = 3

Invoke-RestMethod `
  -Method Post `
  -Uri "http://localhost:8080/api/v1/admin/daily-reports/$reportID/publish" `
  -Headers @{ "X-Daily-Report-Token" = "your-token" }
```

成功時回傳 HTTP 200：

```json
{
  "id": 3,
  "status": "published",
  "published_at": "2026-08-11T09:00:00+08:00"
}
```

不存在的日報回傳 404；不是草稿回傳 409；內容、項目或來源不完整時回傳 422。此 Token 不可放入前端或公開客戶端。

## 排程契約

未來的每日 08:30 排程只需向上述端點傳送 `{}`。排程時間、重試與失敗通知應由 GitHub Actions、Render Cron Job 或其他排程服務負責；新聞搜尋、分析與草稿儲存則由本端點負責。

目前使用的版本化格式提示詞位於 `backend/internal/dailyreport/prompts/format_report.tmpl`。
