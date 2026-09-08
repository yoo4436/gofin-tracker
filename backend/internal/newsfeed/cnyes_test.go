package newsfeed

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFetchLatestClassifiesAndNormalizesArticles(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/tw_stock"):
			_, _ = w.Write([]byte(`{"items":{"data":[{"newsId":1,"title":"台股新聞","publishAt":1788760000,"categoryName":"台股新聞","coverSrc":{"m":{"src":"https://images.example/tw.jpg"}}}]}}`))
		case strings.HasSuffix(r.URL.Path, "/wd_stock"):
			_, _ = w.Write([]byte(`{"items":{"data":[{"newsId":2,"title":"國際新聞","publishAt":1788760100,"categoryName":"國際政經"},{"newsId":3,"title":"美股新聞","publishAt":1788760200,"categoryName":"美股雷達"}]}}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	feed, err := NewCNYESClient(Config{BaseURL: server.URL, HTTPClient: server.Client()}).Fetch(context.Background(), "")
	if err != nil {
		t.Fatalf("Fetch returned an error: %v", err)
	}
	if len(feed.Articles) != 3 {
		t.Fatalf("got %d articles, want 3", len(feed.Articles))
	}
	if feed.Articles[0].Category != CategoryUS || feed.Articles[1].Category != CategoryInternational || feed.Articles[2].Category != CategoryTaiwan {
		t.Fatalf("unexpected category order: %#v", feed.Articles)
	}
	if feed.Articles[2].ImageURL != "https://images.example/tw.jpg" {
		t.Fatalf("unexpected image URL: %s", feed.Articles[2].ImageURL)
	}
}

func TestFetchSearchRemovesHighlightMarkup(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/search/news" || r.URL.Query().Get("q") != "台積電" {
			t.Fatalf("unexpected request: %s", r.URL.String())
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"items":{"data":[{"newsId":8,"title":"<mark>台積電</mark>上漲","publishAt":1788760000,"signature":"鉅亨網記者","category":[{"slug":"tw_quo","name":"台股盤勢"}]}]}}`))
	}))
	defer server.Close()

	feed, err := NewCNYESClient(Config{BaseURL: server.URL, HTTPClient: server.Client()}).Fetch(context.Background(), "台積電")
	if err != nil {
		t.Fatalf("Fetch returned an error: %v", err)
	}
	if len(feed.Articles) != 1 || feed.Articles[0].Title != "台積電上漲" {
		t.Fatalf("unexpected search result: %#v", feed.Articles)
	}
}

func TestFetchRejectsProviderFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	_, err := NewCNYESClient(Config{BaseURL: server.URL, HTTPClient: server.Client()}).Fetch(context.Background(), "AI")
	if err == nil || !strings.Contains(err.Error(), "429") {
		t.Fatalf("expected provider status error, got %v", err)
	}
}
