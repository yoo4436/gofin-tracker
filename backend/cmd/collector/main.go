package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	// 💡 記得引入 PostgreSQL 驅動（必須執行 go get github.com/lib/pq）
	_ "github.com/lib/pq"
)

// 這裡定義一個與你資料庫欄位對應的 Go 結構體
type Candlestick struct {
	Time       time.Time
	OpenPrice  float64
	HighPrice  float64
	LowPrice   float64
	ClosePrice float64
	Volume     float64
}

func main() {
	// 1. 載入根目錄的 .env 檔案
	err := godotenv.Load()
	if err != nil {
		log.Fatal("載入 .env 檔案失敗")
	}

	// 2. 讀取環境變數
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		log.Fatal("找不到 DATABASE_URL 環境變數")
	}

	// 3. 連線資料庫
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("資料庫連線失敗: %v", err)
	}
	defer db.Close()

	// 實際戳戳看資料庫，確保網路有通
	err = db.Ping()
	if err != nil {
		log.Fatalf("雲端資料庫連線失敗 (請檢查帳密或防火牆): %v", err)
	}
	fmt.Println("🎉 雲端資料庫連線成功！")

	// ==========================================
	// 2. 戳 幣安 API 抓取 BTCUSDT 日線歷史資料
	// ==========================================
	symbol := "BTCUSDT"
	interval := "1d"
	limit := 50 // 先抓最近 50 根測試就好

	url := fmt.Sprintf("https://api.binance.com/api/v3/klines?symbol=%s&interval=%s&limit=%d", symbol, interval, limit)
	fmt.Printf("正在從幣安抓取資料... 網址: %s\n", url)

	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("發送網路請求失敗: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("幣安 API 回傳錯誤狀態碼: %d", resp.StatusCode)
	}

	// ==========================================
	// 3. 解析幣安神奇的「二維不規則陣列 JSON」
	// ==========================================
	// 幣安格式: [ [開放時間, "開盤價", "最高價", "最低價", "收盤價", "成交量", ...], [...] ]
	// 在 Go 裡面最快解析這種雜亂陣列的做法是解析成 [][]interface{}
	var rawData [][]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&rawData); err != nil {
		log.Fatalf("JSON 解析失敗: %v", err)
	}

	fmt.Printf("成功抓取！開始處理 %d 條 K 線數據...\n", len(rawData))

	// ==========================================
	// 4. 轉換型別並寫入資料庫
	// ==========================================
	exchangeSymbolID := 1 // 假設你在 exchange_symbols 表中，BTCUSDT 的 id 是 1

	for _, item := range rawData {
		// 幣安格式第 0 項是毫秒時間戳 (在 json 轉換中會先變成 float64)
		openTimeMs := int64(item[0].(float64))
		// 將毫秒轉成 Go 的 time.Time 格式
		klineTime := time.Unix(0, openTimeMs*int64(time.Millisecond))

		// 💡 經典坑：價格在 JSON 裡是字串型別 (例如 "64500.20")，必須手動轉成 float64
		openPrice, _ := strconv.ParseFloat(item[1].(string), 64)
		highPrice, _ := strconv.ParseFloat(item[2].(string), 64)
		lowPrice, _ := strconv.ParseFloat(item[3].(string), 64)
		closePrice, _ := strconv.ParseFloat(item[4].(string), 64)
		volume, _ := strconv.ParseFloat(item[5].(string), 64)

		// 準備寫入資料庫（使用 ON CONFLICT 防止重複主鍵噴錯）
		query := `
			INSERT INTO klines (time, exchange_symbol_id, interval, open_price, high_price, low_price, close_price, volume)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (time, exchange_symbol_id, interval) DO NOTHING;
		`

		_, err := db.Exec(query, klineTime, exchangeSymbolID, interval, openPrice, highPrice, lowPrice, closePrice, volume)
		if err != nil {
			fmt.Printf("⚠️ 寫入時間為 %s 的資料失敗: %v\n", klineTime, err)
			continue
		}
	}

	fmt.Println("🚀 所有測試資料處理完畢，請至雲端資料庫檢查成果！")
}
