package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const (
	binanceKlinesURL = "https://api.binance.com/api/v3/klines"
	klineInterval    = "1d"
	klineLimit       = 500
)

type exchangeSymbol struct {
	ID              int
	TradingPairCode string
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env not found; using environment variables")
	}
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		log.Fatal("DATABASE_URL is required")
	}
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalf("connect database: %v", err)
	}

	products, err := loadBinanceSymbols(db)
	if err != nil {
		log.Fatalf("load exchange symbols: %v", err)
	}
	if len(products) == 0 {
		log.Println("no active Binance exchange symbols with a trading_pair_code")
		return
	}

	client := &http.Client{Timeout: 30 * time.Second}
	failed := 0
	for _, product := range products {
		if err := collectKlines(client, db, product); err != nil {
			failed++
			log.Printf("collect %s (exchange_symbol_id=%d): %v", product.TradingPairCode, product.ID, err)
		}
	}
	log.Printf("collector finished: total=%d succeeded=%d failed=%d", len(products), len(products)-failed, failed)
}

func loadBinanceSymbols(db *sql.DB) ([]exchangeSymbol, error) {
	rows, err := db.Query(`
		SELECT es.id, es.trading_pair_code
		FROM exchange_symbol es
		JOIN exchange e ON e.id = es.exchange_id
		WHERE LOWER(e.name) = 'binance'
		  AND NULLIF(TRIM(es.trading_pair_code), '') IS NOT NULL
		ORDER BY es.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []exchangeSymbol
	for rows.Next() {
		var product exchangeSymbol
		if err := rows.Scan(&product.ID, &product.TradingPairCode); err != nil {
			return nil, err
		}
		products = append(products, product)
	}
	return products, rows.Err()
}

func collectKlines(client *http.Client, db *sql.DB, product exchangeSymbol) error {
	endpoint, err := url.Parse(binanceKlinesURL)
	if err != nil {
		return err
	}
	query := endpoint.Query()
	query.Set("symbol", product.TradingPairCode)
	query.Set("interval", klineInterval)
	query.Set("limit", strconv.Itoa(klineLimit))
	endpoint.RawQuery = query.Encode()

	resp, err := client.Get(endpoint.String())
	if err != nil {
		return fmt.Errorf("call Binance API: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Binance API returned %s", resp.Status)
	}

	var rawData [][]json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&rawData); err != nil {
		return fmt.Errorf("decode Binance response: %w", err)
	}
	for index, item := range rawData {
		if len(item) < 6 {
			return fmt.Errorf("kline %d has %d fields", index, len(item))
		}
		var openTimeMs int64
		var values [5]string
		if err := json.Unmarshal(item[0], &openTimeMs); err != nil {
			return fmt.Errorf("decode kline %d open time: %w", index, err)
		}
		for i := range values {
			if err := json.Unmarshal(item[i+1], &values[i]); err != nil {
				return fmt.Errorf("decode kline %d field %d: %w", index, i+1, err)
			}
		}

		_, err := db.Exec(`
			INSERT INTO klines (time, exchange_symbol_id, interval, open_price, high_price, low_price, close_price, volume)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (time, exchange_symbol_id, interval)
			DO UPDATE SET
				high_price = GREATEST(klines.high_price, EXCLUDED.high_price),
				low_price = LEAST(klines.low_price, EXCLUDED.low_price),
				close_price = EXCLUDED.close_price,
				volume = EXCLUDED.volume`,
			time.UnixMilli(openTimeMs), product.ID, klineInterval,
			values[0], values[1], values[2], values[3], values[4],
		)
		if err != nil {
			return fmt.Errorf("upsert kline %d: %w", index, err)
		}
	}
	log.Printf("stored %d klines for %s (exchange_symbol_id=%d)", len(rawData), product.TradingPairCode, product.ID)
	return nil
}
