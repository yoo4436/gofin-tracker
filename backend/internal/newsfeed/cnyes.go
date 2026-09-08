package newsfeed

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	defaultCNYESBaseURL = "https://api.cnyes.com/media/api/v1"
	// maxResponseBytes 限制從 CNYES API 讀取的最大字節數，以避免過大的響應導致內存問題
	maxResponseBytes   = 4 << 20
	categoryFetchLimit = 30
)

type Category string

const (
	CategoryTaiwan        Category = "taiwan"
	CategoryInternational Category = "international"
	CategoryUS            Category = "us"
)

// 對外統一使用的新聞格式，不讓前端直接依賴鉅亨原始 JSON
type Article struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	ImageURL    string    `json:"image_url"`
	Source      string    `json:"source"`
	Category    Category  `json:"category"`
	PublishedAt time.Time `json:"published_at"`
}

// Feed 包含了新聞查詢結果的完整資訊，包括查詢字串、抓取時間以及文章列表
type Feed struct {
	Query     string    `json:"query"`
	FetchedAt time.Time `json:"fetched_at"`
	Articles  []Article `json:"articles"`
}

// Provider 定義了新聞來源的接口，允許不同的實現來抓取新聞資料
type Provider interface {
	Fetch(context.Context, string) (Feed, error)
}

type Config struct {
	BaseURL    string
	HTTPClient *http.Client
}

type CNYESClient struct {
	baseURL    string
	httpClient *http.Client
}

type cnyesResponse struct {
	Items struct {
		Data []cnyesArticle `json:"data"`
	} `json:"items"`
}

type cnyesArticle struct {
	NewsID       int64           `json:"newsId"`
	Title        string          `json:"title"`
	PublishAt    int64           `json:"publishAt"`
	Source       string          `json:"source"`
	Signature    string          `json:"signature"`
	CategoryName string          `json:"categoryName"`
	Categories   []cnyesCategory `json:"category"`
	CoverSrc     cnyesCover      `json:"coverSrc"`
}

type cnyesCategory struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type cnyesCover struct {
	Large  cnyesImage `json:"l"`
	Medium cnyesImage `json:"m"`
}

type cnyesImage struct {
	Src string `json:"src"`
}

func NewCNYESClient(config Config) *CNYESClient {
	// 確保 baseURL 不以斜線結尾，並去除多餘的空白
	baseURL := strings.TrimRight(strings.TrimSpace(config.BaseURL), "/")
	if baseURL == "" {
		baseURL = defaultCNYESBaseURL
	}
	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 12 * time.Second}
	}
	return &CNYESClient{baseURL: baseURL, httpClient: httpClient}
}

func (c *CNYESClient) Fetch(ctx context.Context, query string) (Feed, error) {
	query = strings.TrimSpace(query)
	var articles []Article // nil slice
	var err error          // 非 nil 的空 slice
	if query == "" {
		articles, err = c.fetchLatest(ctx)
	} else {
		articles, err = c.fetchSearch(ctx, query)
	}
	if err != nil {
		return Feed{}, err
	}
	if articles == nil {
		articles = []Article{}
	}
	sort.SliceStable(articles, func(i, j int) bool {
		return articles[i].PublishedAt.After(articles[j].PublishedAt)
	})
	return Feed{Query: query, FetchedAt: time.Now().UTC(), Articles: articles}, nil
}

func (c *CNYESClient) fetchLatest(ctx context.Context) ([]Article, error) {
	type result struct {
		articles []cnyesArticle
		err      error
	}
	// 使用 channel 和 WaitGroup 來並行抓取不同分類的新聞，並收集結果
	results := make(chan result, 2)
	var waitGroup sync.WaitGroup
	for _, slug := range []string{"tw_stock", "wd_stock"} {
		waitGroup.Add(1)
		go func(categorySlug string) {
			// 確保在 goroutine 結束時調用 Done，避免 WaitGroup 永遠等待
			defer waitGroup.Done()
			items, err := c.request(ctx, "/newslist/category/"+categorySlug, url.Values{
				"limit": []string{fmt.Sprint(categoryFetchLimit)},
			})
			results <- result{articles: items, err: err}
		}(slug)
	}
	// 等待所有 goroutine 完成後關閉 channel，避免死鎖
	waitGroup.Wait()
	close(results)

	var all []Article
	var failures []error
	for response := range results {
		if response.err != nil {
			failures = append(failures, response.err)
			continue
		}
		//normalize 先轉換資料。 後面的 ... 是把 slice 拆開逐筆加入
		all = append(all, normalize(response.articles)...)
	}
	if len(failures) == 2 {
		return nil, fmt.Errorf("fetch CNYES news categories: %w", errors.Join(failures...))
	}
	return deduplicateAndLimit(all, 10), nil
}

func (c *CNYESClient) fetchSearch(ctx context.Context, query string) ([]Article, error) {
	items, err := c.request(ctx, "/search/news", url.Values{
		"q":    []string{query},
		"page": []string{"1"},
	})
	if err != nil {
		return nil, fmt.Errorf("search CNYES news: %w", err)
	}
	return deduplicateAndLimit(normalize(items), 10), nil
}

func (c *CNYESClient) request(ctx context.Context, path string, query url.Values) ([]cnyesArticle, error) {
	endpoint, err := url.Parse(c.baseURL + path)
	if err != nil {
		return nil, fmt.Errorf("parse CNYES URL: %w", err)
	}
	endpoint.RawQuery = query.Encode()
	// 使用 http.NewRequestWithContext 來創建請求，確保可以在 context 被取消時中止請求
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create CNYES request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "GoFin-Tracker/1.0")

	response, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call CNYES API: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 1024))
		return nil, fmt.Errorf("CNYES API returned %d", response.StatusCode)
	}

	var payload cnyesResponse
	decoder := json.NewDecoder(io.LimitReader(response.Body, maxResponseBytes))
	if err := decoder.Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode CNYES response: %w", err)
	}
	return payload.Items.Data, nil
}

func normalize(items []cnyesArticle) []Article {
	articles := make([]Article, 0, len(items))
	for _, item := range items {
		category, ok := classify(item)
		title := stripMarkup(item.Title)
		if !ok || item.NewsID <= 0 || title == "" || item.PublishAt <= 0 {
			continue
		}
		source := strings.TrimSpace(item.Source)
		if source == "" {
			source = strings.TrimSpace(item.Signature)
		}
		if source == "" {
			source = "鉅亨網"
		}
		imageURL := strings.TrimSpace(item.CoverSrc.Large.Src)
		if imageURL == "" {
			imageURL = strings.TrimSpace(item.CoverSrc.Medium.Src)
		}
		//PublishedAt 使用 UTC 時間，前端再用使用者當地時區顯示。
		articles = append(articles, Article{
			ID:          item.NewsID,
			Title:       title,
			URL:         fmt.Sprintf("https://news.cnyes.com/news/id/%d", item.NewsID),
			ImageURL:    imageURL,
			Source:      source,
			Category:    category,
			PublishedAt: time.Unix(item.PublishAt, 0).UTC(),
		})
	}
	return articles
}

func classify(item cnyesArticle) (Category, bool) {
	names := []string{item.CategoryName}
	slugs := make([]string, 0, len(item.Categories))
	for _, category := range item.Categories {
		names = append(names, category.Name)
		slugs = append(slugs, category.Slug)
	}
	joinedNames := strings.Join(names, " ")
	joinedSlugs := strings.Join(slugs, " ")
	switch {
	case strings.Contains(joinedNames, "台股") || strings.Contains(joinedSlugs, "tw_"):
		return CategoryTaiwan, true
	case strings.Contains(joinedNames, "美股") || strings.Contains(joinedSlugs, "us_stock"):
		return CategoryUS, true
	case strings.Contains(joinedNames, "國際") || strings.Contains(joinedNames, "歐亞") ||
		strings.Contains(joinedNames, "大陸政經") || strings.Contains(joinedNames, "陸港股") ||
		strings.Contains(joinedSlugs, "cn_stock") || strings.Contains(joinedSlugs, "cn_macro"):
		return CategoryInternational, true
	default:
		return "", false
	}
}

func stripMarkup(value string) string {
	value = html.UnescapeString(value)
	value = strings.ReplaceAll(value, "<mark>", "")
	value = strings.ReplaceAll(value, "</mark>", "")
	return strings.TrimSpace(value)
}

func deduplicateAndLimit(items []Article, limitPerCategory int) []Article {
	seen := make(map[int64]struct{})
	counts := make(map[Category]int)
	result := make([]Article, 0, len(items))
	for _, item := range items {
		if _, exists := seen[item.ID]; exists || counts[item.Category] >= limitPerCategory {
			continue
		}
		seen[item.ID] = struct{}{}
		counts[item.Category]++
		result = append(result, item)
	}
	return result
}
