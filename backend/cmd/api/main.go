package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	// 💡 新增 os 套件
	"github.com/gin-gonic/gin" // 💡 新增 godotenv 套件
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// 升級回應結構體：把 MACD 的數據一起包進去吐給前端
type KlineResponse struct {
	Time  int64   `json:"time"`
	Open  float64 `json:"open"`
	High  float64 `json:"high"`
	Low   float64 `json:"low"`
	Close float64 `json:"close"`
	// 💡 新增 MACD 欄位
	Dif  float64 `json:"dif"`
	Dea  float64 `json:"dea"`
	Hist float64 `json:"hist"`
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

	r := gin.Default()

	// CORS 跨域處理
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		c.Next()
	})

	v1 := r.Group("/api/v1")
	{
		v1.GET("/klines", func(c *gin.Context) {
			// 💡 實務小常識：EMA 需要一定的「熱身期（Warm-up）」計算才會精準。
			// 雖然前端可能只想看 100 根，但我們從資料庫撈出 200 根來計算，算完再切給前端。
			query := `SELECT time, open_price, high_price, low_price, close_price 
					  FROM klines 
					  WHERE exchange_symbol_id = 1 AND interval = '1d'
					  ORDER BY time ASC LIMIT 200;`

			rows, err := db.Query(query)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			defer rows.Close()

			var klines []KlineResponse
			var closePrices []float64 // 用來單獨存放收盤價，餵給 MACD 演算法

			for rows.Next() {
				var t time.Time
				var k KlineResponse
				err := rows.Scan(&t, &k.Open, &k.High, &k.Low, &k.Close)
				if err != nil {
					continue
				}
				k.Time = t.Unix()
				klines = append(klines, k)
				closePrices = append(closePrices, k.Close)
			}

			// 💡 核心：呼叫我們寫好的純 Go MACD 計算器
			dif, dea, hist := calculateMACD(closePrices)

			// 將算好的指標數據塞回原本的 klines 陣列中
			for i := 0; i < len(klines); i++ {
				klines[i].Dif = dif[i]
				klines[i].Dea = dea[i]
				klines[i].Hist = hist[i]
			}

			// 如果前端只需要 100 根，我們就把前面用來熱身的舊資料切掉，只給最後 100 根
			if len(klines) > 100 {
				klines = klines[len(klines)-100:]
			}

			c.JSON(http.StatusOK, klines)
		})
	}

	r.Run(":8080")
}

// =========================================================================
// 🧮 純 Go 語言技術指標演算法核心
// =========================================================================

// 計算指數移動平均線 (EMA)
func calculateEMA(prices []float64, period int) []float64 {
	ema := make([]float64, len(prices))
	if len(prices) == 0 {
		return ema
	}

	// 平滑係數 alpha 公式
	alpha := 2.0 / (float64(period) + 1.0)

	// 第一根 K 線的 EMA 預設等於它自己的收盤價
	ema[0] = prices[0]

	// 依序迭代計算後續的 EMA
	for i := 1; i < len(prices); i++ {
		ema[i] = prices[i]*alpha + ema[i-1]*(1.0-alpha)
	}
	return ema
}

// 計算 MACD (12, 26, 9)
func calculateMACD(prices []float64) ([]float64, []float64, []float64) {
	length := len(prices)
	dif := make([]float64, length)
	hist := make([]float64, length)

	if length == 0 {
		return dif, dif, dif
	}

	// 1. 算出 12 EMA 與 26 EMA
	ema12 := calculateEMA(prices, 12)
	ema26 := calculateEMA(prices, 26)

	// 2. 算出 DIF (快線) = 12EMA - 26EMA
	for i := 0; i < length; i++ {
		dif[i] = ema12[i] - ema26[i]
	}

	// 3. 算出 DEA (慢線) = DIF 的 9 EMA
	dea := calculateEMA(dif, 9)

	// 4. 算出 HIST (柱狀圖) = DIF - DEA
	for i := 0; i < length; i++ {
		hist[i] = dif[i] - dea[i]
	}

	return dif, dea, hist
}
