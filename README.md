# GoFin-Tracker 🚀

一個基於 **Go** 與 **Vue 3 (TypeScript)** 構建的高性能「即時金融監控與指標分析中台」。本系統旨在幫助投資者整合多市場（加密貨幣、股票）數據，透過後端自動化指標計算（MACD, Bollinger Bands, RSI）提供宏觀的決策分析輔助。

🔗 **線上展示連結**：[前往前端 Vercel 部署網頁](https://gofin-tracker-flame.vercel.app/)  
*(提示：後端 API 部署於 Render 免費方案，首次載入若需等待 30-50 秒為 Server 喚醒之冷啟動現象。)*

---

## 💡 核心動機 (Why this project?)

本專案起源於對金融投資的濃厚興趣，以及想親自動手實現一套完整交易系統的好奇心。在實際使用市面上的看盤工具時，我發現數據抓取、指標運算與前端視覺化往往流於碎片化。

同時，這也是我為了深入掌握 **Go 語言** 高併發特性與記憶體管理所打造的實作練習。因量化交易極度看重系統的吞吐量與啟動速度，Go 語言精簡的語法與強大的協程模型（Goroutine）非常符合這個系統的相性。

---

## 🛠️ 技術棧 (Tech Stack)

### 後端 (Backend) - Go Ecosystem
* **核心語言**: Go
* **Web 框架**: Gin (高效率路由、自動 JSON 綁定)
* **資料庫驅動**: pgx (高效能原生 PostgreSQL 驅動)
* **環境管理**: godotenv (機密資訊與連線字串抽離，確保資安)

### 前端 (Frontend) - Vue 3 流派
* **核心框架**: Vue 3 (Composition API) + Vite
* **開發語言**: TypeScript (強型別嚴格規範，大幅減少前後端對接 Bug)
* **圖表視覺化**: TradingView Lightweight Charts (高性能 Canvas 繪圖)

### 基礎設施與自動化 (DevOps & Infrastructure)
* **Database**: Supabase (PostgreSQL)
* **Deployment**: Vercel (Frontend靜態託管) + Render (Backend常駐服務)
* **CI/CD & 保活機制**: GitHub Actions Workflow (實作雲端自動化排程與 PING 保活機制) (但因Actions會有延遲關係，日後需要再另行解決方案)

---

## 🧠 架構決策與設計理念 (Architectural Decisions)

在開發過程中，為了解決雲端資源限制與系統效能，本專案採取了以下核心架構設計：

### 1. 冪等性與容量極致優化 (UPSERT 機制)
為了解決 Supabase 免費版的容量限制，並確保背景排程抓取資料時的穩定性，系統在 `klines` 價格表設定了 `(time, exchange_symbol_id, interval)` 的**複合主鍵**。
在寫入端全面採用 PostgreSQL 的 `ON CONFLICT DO UPDATE SET` (UPSERT) 機制。配合 `GREATEST` 與 `LEAST` 函數，常駐程式在輪詢時僅更新當天日線的一列數據。這不僅實現了極高的冪等性（程式即使斷線重啟也不會產生髒數據），也確保了 500MB 的免費空間可穩定常駐運行數十年。

### 2. 記憶體內高效計算 (In-Memory Processing)
為了維持極致的系統效能，所有的技術指標（MACD、MA 均線、布林帶三軌、RSI）**皆不在資料庫開表儲存**。當前端發起請求時，Go 後端利用極快的浮點數迴圈運算，在伺服器記憶體內即時計算完成並封裝 JSON。這避免了頻繁的資料庫 I/O，實現了近乎零延遲的指標呈現。

### 3. 多對多市場抽象化設計 (Market Abstraction)
考量到未來需整合台股或期貨市場，資料庫設計導入了 `exchange_symbols` 作為中介對照表，將「商品定義」與「價格數據」完美解耦。後端的資料抓取器（Collector）與 API 伺服器（API Server）也遵循 Monorepo 架構拆分為雙核心執行檔 (`cmd/collector` 與 `cmd/api`)，確保職責分離。

---

## ⚙️ 如何在本機運行 (Quick Start)

### 1. 複製本專案
```bash
git clone [https://github.com/yoo4436/gofin-tracker.git] 
cd gofin-tracker
```

### 2. 環境變數設定
```bash
請參考專案根目錄下的 .env.example，在根目錄建立一個 .env 檔案，並填入你的 Supabase 連線字串（請換上你當初設定的正確密碼）：
DATABASE_URL="postgresql://postgres.[YOUR_PROJECT_ID]:[YOUR_PASSWORD]@aws-0-xxx.pooler.supabase.com:6543/postgres"
```

### 3. 啟動後端服務 (API Server)
確保你的終端機路徑停留在專案根目錄（gofin-tracker/），執行以下指令來下載 Go 套件並啟動服務：
```bash
# 下載後端所需的依賴套件 (如 Gin, pgx, godotenv)
go mod tidy

# 啟動 API 伺服器 (預設會動態綁定或監聽 Port 8080)
go run backend/cmd/api/main.go
```

若需要手動執行資料抓取器更新最新 K 線數據，可另外開啟終端機執行：
```bash
go run backend/cmd/collector/main.go
```

### 4. 啟動前端環境
開啟另一個新的終端機視窗，切換到前端資料夾，下載套件並啟動 Vite 測試伺服器：
```bash
cd frontend
npm install
npm run dev
 ```

---
