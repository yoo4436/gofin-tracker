# GoFin-Tracker agent 作業規範

## 共通規則

- 依使用者要求決定任務範圍；只要求解釋或設計時，不修改程式。
- 修改前讀取相關程式、指引、`git status --short` 與目前分支，保留既有修改。
- 延續目前 feature branch，不直接在 `main` 或 `dev` 開發。需要建立或切換分支時依使用者授權處理。
- 未經明確授權，不 commit、push、開 PR、merge、rebase、reset、刪除分支或部署；同一任務的既有授權不必重問。
- CI/CD 改動必須屬於任務範圍，不因驗證觸發正式環境操作。
- 不輸出或提交 `.env`、token、金鑰、登入狀態及個人機器路徑。
- 用繁體中文說明，先講結果，再補必要證據。公開 README 維持產品導向，開發操作文件放 `docs/`。

## 實作與驗證

- Go/Gin 後端位於 `backend/`，Go module 在 repo 根目錄；Vue/TypeScript 前端在 `frontend/`；migration 在 `supabase/migrations/`。
- 先追蹤呼叫端、資料流與測試，只改任務需要的範圍；維持 API 與資料相容性，除非任務要求改變。
- Go 改動先跑受影響套件測試，必要時於根目錄執行 `go test ./...`，並格式化修改的 Go 檔案。
- 前端改動於 `frontend/` 執行 `npm run build`；Windows PowerShell 可用 `npm.cmd run build`。UI 缺陷另驗證指定頁面與互動，build 成功不代表畫面正確。
- 資料庫工作讀取 `.agents/skills/gofin-db-migration/SKILL.md`；盤點授權不等於 reset 或遠端寫入授權。
- 純文件及 agent 設定檢查語法、連結與 diff，不必啟動應用或資料庫。
- 先查看 `.github/workflows/` 再判斷 CI 能力；keep-alive workflow 不代表測試已通過。
- 回報實際執行的檢查與限制，不把本機結果描述為遠端 CI 結果。

## 角色與交接

- 主 agent 負責任務邊界、分工、整合與最終驗證。小任務自行完成即可，不必固定啟動全部角色。
- 子 agent 的使用遵守使用者要求及執行環境的委派規則；建立角色檔案不代表每個任務都必須平行執行。
- 委派時提供目標、相關指引、可修改的檔案範圍、交付物與驗證責任，不假設子 agent 已取得全部背景。
- `gofin_explorer`：唯讀追蹤程式、資料流與測試，提出證據及未知事項。
- `gofin_implementer`：修改分配的範圍並執行相關本機驗證。
- `gofin_reviewer`：唯讀檢查正確性、回歸與驗證缺口，提出有位置和證據的問題。
- 不同 agent 不同時修改同一檔案。需要跨範圍修改時，先交由主 agent 重新分配，不覆寫他人工作。
- 子 agent 不自行操作分支、發布 Git 變更、部署或寫入遠端資料庫；交由主 agent 依既有授權處理。
- 交接列出發現或改動、檔案位置、已執行檢查與結果、待解問題；主 agent 檢查整合後 diff。

## Skills 與跨裝置

- 本檔及 repo 內的 skills、角色設定是共同規範來源，不依賴特定電腦的個人 skill。
- 個人 `implement-code-safely` 若可用，可輔助通用實作；專案必要規則已列於本檔，不另複製同名 skill。
- `.agents/skills/speak-human-tw/SKILL.md` 用於使用者要求說人話、去 AI 味或潤稿的文字工作；不套用到程式碼、log 或設定檔，不自動接進 CI。使用時保留事實、數字及技術意義，遵守使用者已提供的授權。
- 雙裝置設定、Git 交接及第三方 skill 維護見 [開發環境設定](docs/development-setup.md)。
