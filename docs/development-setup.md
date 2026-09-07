# 開發環境設定與雙裝置維護

## 共用檔案

| 位置 | 用途 |
| --- | --- |
| `AGENTS.md` | 專案規則、驗證與交接責任 |
| `.agents/skills/gofin-db-migration/` | Supabase schema 工作流程 |
| `.agents/skills/speak-human-tw/` | 第三方繁體中文潤稿 skill |
| `.codex/agents/` | 探索、實作、審查角色 |

這些檔案跟程式碼一起由 Git 管理，不需要另一個 repo。兩台裝置取得相同 commit，即取得相同的專案規範；個人設定或 CLI 版本仍可能影響實際行為。

## 每台裝置首次設定

1. 各自 clone 此 repo，不用雲端同步資料夾同步整個工作目錄或 `.git`。
2. 安裝相容的 Codex CLI、Go、Node/npm；需要本機資料庫才另裝 Supabase CLI 與 Docker。Go 版本以 `go.mod` 為準，前端以 lockfile 安裝依賴。
3. 各自設定登入、模型、權限及環境變數。依 `.env.example` 配置秘密，不把 token、登入檔或含個人路徑的全域設定提交到 Git。
4. 執行 `codex --version`，盡量讓兩台使用相同版本。PowerShell 若遇到 `.ps1` 執行限制，可用 `codex.cmd`、`npm.cmd`。
5. 從 repo 根目錄開啟新的 Codex 工作階段，請它列出載入的指引；用 `/skills` 確認兩個專案 skills。
6. 確認 CLI 支援 `.codex/agents/*.toml`，再明確要求小型唯讀任務測試角色是否載入。TOML 語法通過不代表 CLI 已載入角色。

本專案不指定模型或 reasoning effort，角色沿用父工作階段設定。唯讀角色的 sandbox 預設仍受父工作階段及實際執行環境影響。個人 `implement-code-safely` 可以繼續使用，另一台未安裝也不影響本專案必要規範。

## 換裝置繼續工作

下列是使用者手動交接流程，不代表 agent 已取得 commit、push 或切分支授權。

1. 裝置 A 查看分支、`git status` 與 diff，完成需要的驗證。
2. 將要交接的變更 commit 並 push 到目前 feature branch，記下分支、commit SHA、已跑檢查與待辦。未 push 的 commit、未提交修改及 stash 不會出現在 B。
3. 裝置 B 先保存自己的修改，再 fetch、切到交接分支並 `git pull --ff-only`。若分歧或衝突，先釐清兩端內容，不用 reset 或 force-push 蓋掉工作。
4. 比對 commit SHA、準備依賴，再開新 Codex 工作階段。Git 不同步舊對話及授權，需提供必要背景與待辦。
5. 同一任務盡量交接後才繼續；若兩台同時開發，使用不同 feature branches 並分配不同範圍，再依既有 review 流程整合。

兩台本機 Supabase 資料庫各自獨立；拉取 migration 不會同步資料或自動套用 SQL。重建前先確認目標、資料是否可丟棄及 reset 授權。

## 說人話 skill

- 上游：[Raymondhou0917/speak-human-tw](https://github.com/Raymondhou0917/speak-human-tw)，MIT 授權；保留原始 LICENSE 及 skill 所需資源。`assets/readme/` 只是上游展示圖片，由根目錄 `.gitignore` 排除，不隨 Git 同步；需查看完整圖文介紹時使用上游連結。
- 本次以官方 skill-installer 從 `master` 下載，SKILL.md 標示版本 `1.4.0`。這是 repo 內的副本，不會自動跟隨上游更新；確切導入內容以本專案提交記錄為準。
- 下一輪可使用 `$speak-human-tw`；若未顯示，重開 Codex 並檢查 `/skills`。尚未在兩台裝置上驗證載入。
- 上游預設先列出潤稿建議再確認；使用者可明確要求「這次直接套用，不用先問」。本次安裝不包含自動化 pipeline 整合。
- 更新時先在 feature branch 比較上游差異、授權與指引，再替換並驗證，commit/push 後讓另一台拉取。不要在兩台各自更新副本，也不要直接執行未檢查的安裝腳本。

## 維護規範

每次工作都適用的規則放 `AGENTS.md`；特定任務步驟放 skills；角色責任放 agent TOML。修改後檢查 YAML/TOML、引用檔案與 diff，再以相符任務確認行為。若未來多個 repo 都要共用同一組 skills，再評估獨立儲存庫或 plugin。

官方文件：[AGENTS.md](https://learn.chatgpt.com/docs/agent-configuration/agents-md)、[Skills](https://learn.chatgpt.com/docs/build-skills)、[Subagents](https://learn.chatgpt.com/docs/agent-configuration/subagents)。本次本機版本為 `codex-cli 0.153.4`，並非最低版本宣告；自訂角色尚待實際 CLI 載入驗證。
