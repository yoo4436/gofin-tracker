package main

import (
	"database/sql"
	"log"
	"math"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

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
		v1.GET("/klines", func(c *gin.Context) {
			// 💡 EMA/RSI 需要熱身期，這次撈 200 根算完再切 100 根給前端
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

			// 🧮 記憶體一擊流：一次算好所有技術指標
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
// 🧮 技術指標核心演算法 (純 Go 實作)
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
