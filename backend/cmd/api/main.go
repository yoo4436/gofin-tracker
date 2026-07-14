package main

import (
	"database/sql"
	"fmt" // 新增：用於格式化字串拼接 SQL
	"log"
	"math"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// SymbolResponse 定義商品清單的回傳格式
type SymbolResponse struct {
	ID           int    `json:"id"`
	SymbolCode   string `json:"symbol_code"`
	Name         string `json:"name"`
	MarketType   string `json:"market_type"`
	ExchangeName string `json:"exchange_name"` // 這是 JOIN 拿到的交易所名稱
}

// 升級回應結構體：包進所有全新指標
type KlineResponse struct {
	Time  int64   `json:"time"`
	Open  float64 `json:"open"`
	High  float64 `json:"high"`
	Low   float64 `json:"low"`
	Close float64 `json:"close"`
	// MACD
	Dif  float64 `json:"dif"`
	Dea  float64 `json:"dea"`
	Hist float64 `json:"hist"`
	// MA 線
	Ma7  float64 `json:"ma7"`
	Ma25 float64 `json:"ma25"`
	// 布林帶
	BbiUpper  float64 `json:"bbiUpper"`
	BbiMiddle float64 `json:"bbiMiddle"`
	BbiLower  float64 `json:"bbiLower"`
	// RSI
	Rsi float64 `json:"rsi"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("本地未找到 .env 檔案，將直接讀取系統環境變數")
	}

	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		log.Fatal("找不到 DATABASE_URL 環境變數")
	}

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

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"message": "指數追蹤系統後端 API 正常運行中！",
		})
	})

	v1 := r.Group("/api/v1")
	{
		// 獲取商品清單與動態搜尋 API
		v1.GET("/symbols", func(c *gin.Context) {
			marketType := c.Query("market_type")
			searchQuery := c.Query("q") // 接收前端傳來名叫 'q' 的搜尋字串

			// 修改 SQL 查詢：加入 COALESCE 防止 NULL 炸掉程式
			baseQuery := `
				SELECT 
					s.id, 
					s.symbol_code, 
					COALESCE(s.name, '') AS name, 
					COALESCE(s.market_type, '') AS market_type, 
					e.name AS exchange_name 
				FROM symbol s
				INNER JOIN exchange_symbol es ON s.id = es.symbol_id
				INNER JOIN exchange e ON es.exchange_id = e.id
			`

			// 用來存放要帶入 SQL 的動態參數 (防止 SQL Injection)
			var args []interface{}
			paramIndex := 1

			// 判斷是否需要過濾市場類型
			if marketType != "" {
				baseQuery += fmt.Sprintf(` AND s.market_type = $%d`, paramIndex)
				args = append(args, marketType)
				paramIndex++
			}

			// 判斷是否有搜尋關鍵字 (使用 ILIKE 達成不分大小寫的模糊搜尋)
			if searchQuery != "" {
				baseQuery += fmt.Sprintf(` AND (s.symbol_code ILIKE $%d OR s.name ILIKE $%d)`, paramIndex, paramIndex)
				// 例如搜尋 "btc"，就會變成找包含 btc 的代碼或中文名稱
				searchTerm := "%" + searchQuery + "%"
				args = append(args, searchTerm)
			}

			baseQuery += ` ORDER BY s.id ASC;`

			// 執行查詢，將 args 展開帶入
			rows, err := db.Query(baseQuery, args...)
			if err != nil {
				log.Println("資料庫查詢失敗:", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "查詢商品清單失敗"})
				return
			}
			defer rows.Close()

			var symbols []SymbolResponse
			for rows.Next() {
				var s SymbolResponse
				// 把錯誤印出來：這樣萬一還是錯，我們看終端機就知道兇手是哪個欄位！
				if err := rows.Scan(&s.ID, &s.SymbolCode, &s.Name, &s.MarketType, &s.ExchangeName); err != nil {
					log.Println("⚠️ 資料解析錯誤 (請檢查欄位名稱或型別):", err)
					continue
				}
				symbols = append(symbols, s)
			}

			// 確保如果沒有資料時，回傳的是 [] 而不是 null，避免前端炸掉
			if symbols == nil {
				symbols = []SymbolResponse{}
			}

			c.JSON(http.StatusOK, symbols)
		})

		// 獲取 K 線與技術指標 API
		v1.GET("/klines", func(c *gin.Context) {
			// EMA/RSI 需要熱身期，這次撈 200 根算完再切 100 根給前端
			query := `SELECT time, open_price, high_price, low_price, close_price 
					  FROM klines 
					  WHERE exchange_symbol_id = 1 AND interval = '1d'
					  ORDER BY time ASC;`

			rows, err := db.Query(query)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			defer rows.Close()

			var klines []KlineResponse
			var closePrices []float64

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

			if len(klines) == 0 {
				c.JSON(http.StatusOK, klines)
				return
			}

			// 記憶體一擊流：一次算好所有技術指標
			dif, dea, hist := calculateMACD(closePrices)
			ma7 := calculateSMA(closePrices, 7)
			ma25 := calculateSMA(closePrices, 25)
			bbUpper, bbMiddle, bbLower := calculateBollingerBands(closePrices, 20, 2.0)
			rsi := calculateRSI(closePrices, 14)

			// 將算好的指標數據塞回原本的 klines 陣列中
			for i := 0; i < len(klines); i++ {
				klines[i].Dif = dif[i]
				klines[i].Dea = dea[i]
				klines[i].Hist = hist[i]
				klines[i].Ma7 = ma7[i]
				klines[i].Ma25 = ma25[i]
				klines[i].BbiUpper = bbUpper[i]
				klines[i].BbiMiddle = bbMiddle[i]
				klines[i].BbiLower = bbLower[i]
				klines[i].Rsi = rsi[i]
			}

			// 切出最後 100 根最精準、熱身好的資料給前端
			if len(klines) > 30 {
				klines = klines[30:]
			} else {
				// 如果資料不足 100 根，則返回所有可用資料
			}

			c.JSON(http.StatusOK, klines)
		})
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}

// =========================================================================
// 技術指標核心演算法 (純 Go 實作)
// =========================================================================

// 1. 計算簡單移動平均線 (SMA)
func calculateSMA(prices []float64, period int) []float64 {
	sma := make([]float64, len(prices))
	if len(prices) < period {
		return sma
	}
	var sum float64
	for i := 0; i < period; i++ {
		sum += prices[i]
		sma[i] = sum / float64(i+1) // 前期未滿週期時的平滑處理
	}
	sma[period-1] = sum / float64(period)

	for i := period; i < len(prices); i++ {
		sum = sum - prices[i-period] + prices[i]
		sma[i] = sum / float64(period)
	}
	return sma
}

// 2. 計算指數移動平均線 (EMA)
func calculateEMA(prices []float64, period int) []float64 {
	ema := make([]float64, len(prices))
	if len(prices) == 0 {
		return ema
	}
	alpha := 2.0 / (float64(period) + 1.0)
	ema[0] = prices[0]
	for i := 1; i < len(prices); i++ {
		ema[i] = prices[i]*alpha + ema[i-1]*(1.0-alpha)
	}
	return ema
}

// 3. 計算 MACD
func calculateMACD(prices []float64) ([]float64, []float64, []float64) {
	length := len(prices)
	dif := make([]float64, length)
	hist := make([]float64, length)
	if length == 0 {
		return dif, dif, dif
	}
	ema12 := calculateEMA(prices, 12)
	ema26 := calculateEMA(prices, 26)
	for i := 0; i < length; i++ {
		dif[i] = ema12[i] - ema26[i]
	}
	dea := calculateEMA(dif, 9)
	for i := 0; i < length; i++ {
		hist[i] = dif[i] - dea[i]
	}
	return dif, dea, hist
}

// 4. 計算布林帶 (Bollinger Bands)
func calculateBollingerBands(prices []float64, period int, k float64) ([]float64, []float64, []float64) {
	length := len(prices)
	upper := make([]float64, length)
	middle := calculateSMA(prices, period)
	lower := make([]float64, length)

	for i := 0; i < length; i++ {
		if i < period-1 {
			upper[i] = prices[i]
			lower[i] = prices[i]
			continue
		}
		// 計算標準差
		var variance float64
		for j := i - period + 1; j <= i; j++ {
			variance += math.Pow(prices[j]-middle[i], 2)
		}
		stdDev := math.Sqrt(variance / float64(period))
		upper[i] = middle[i] + k*stdDev
		lower[i] = middle[i] - k*stdDev
	}
	return upper, middle, lower
}

// 5. 計算 RSI (14) - Wilder's Smoothing 標準算法
func calculateRSI(prices []float64, period int) []float64 {
	length := len(prices)
	rsi := make([]float64, length)
	if length <= period {
		return rsi
	}

	gains := make([]float64, length)
	losses := make([]float64, length)

	for i := 1; i < length; i++ {
		change := prices[i] - prices[i-1]
		if change > 0 {
			gains[i] = change
		} else {
			losses[i] = -change
		}
	}

	var avgGain, avgLoss float64
	for i := 1; i <= period; i++ {
		avgGain += gains[i]
		avgLoss += losses[i]
	}
	avgGain /= float64(period)
	avgLoss /= float64(period)

	if avgLoss == 0 {
		rsi[period] = 100
	} else {
		rsi[period] = 100 - (100 / (1 + avgGain/avgLoss))
	}

	for i := period + 1; i < length; i++ {
		avgGain = (avgGain*float64(period-1) + gains[i]) / float64(period)
		avgLoss = (avgLoss*float64(period-1) + losses[i]) / float64(period)

		if avgLoss == 0 {
			rsi[i] = 100
		} else {
			rsi[i] = 100 - (100 / (1 + avgGain/avgLoss))
		}
	}
	return rsi
}
