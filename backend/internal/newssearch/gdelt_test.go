package newssearch

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func TestSearchReturnsUniqueGDELTSourcesInsideRequestedWindow(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("startdatetime"); got != "20260809003000" {
			t.Fatalf("unexpected startdatetime: %s", got)
		}
		if got := r.URL.Query().Get("enddatetime"); got != "20260810003000" {
			t.Fatalf("unexpected enddatetime: %s", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"articles":[
			{"url":"https://one.example/a","title":"AI &amp; markets","seendate":"20260810T001500Z","domain":"one.example","language":"English","sourcecountry":"United States"},
			{"url":"https://two.example/b","title":"Taiwan technology","seendate":"20260809T231500Z","domain":"two.example"},
			{"url":"https://three.example/c","title":"Economic outlook","seendate":"20260809T221500Z","domain":"three.example"},
			{"url":"https://one.example/a","title":"Duplicate","seendate":"20260810T001500Z","domain":"one.example"}
		]}`))
	}))
	defer server.Close()

	client := NewGDELTClient(Config{BaseURL: server.URL, HTTPClient: server.Client()})
	location := time.FixedZone("Asia/Taipei", 8*60*60)
	windowEnd := time.Date(2026, 8, 10, 8, 30, 0, 0, location)
	research, metadata, err := client.Search(context.Background(), windowEnd.Add(-24*time.Hour), windowEnd)
	if err != nil {
		t.Fatalf("Search returned an error: %v", err)
	}
	if len(metadata.Sources) != 3 {
		t.Fatalf("expected 3 unique sources, got %d", len(metadata.Sources))
	}
	if metadata.Sources[0].Title != "AI & markets" || metadata.Sources[0].ObservedAt != "2026-08-10T00:15:00Z" {
		t.Fatalf("unexpected first source: %#v", metadata.Sources[0])
	}
	if research == "" {
		t.Fatal("research notes must not be empty")
	}
}

func TestSearchRejectsInsufficientUsableArticles(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"articles":[{"url":"https://one.example/a","title":"Only one"}]}`))
	}))
	defer server.Close()

	client := NewGDELTClient(Config{BaseURL: server.URL, HTTPClient: server.Client()})
	_, _, err := client.Search(context.Background(), time.Now().Add(-time.Hour), time.Now())
	if err == nil {
		t.Fatal("expected an error")
	}
}

func TestSearchRetriesRateLimit(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if calls.Add(1) == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"articles":[
			{"url":"https://one.example/a","title":"One"},
			{"url":"https://two.example/b","title":"Two"},
			{"url":"https://three.example/c","title":"Three"}
		]}`))
	}))
	defer server.Close()

	client := NewGDELTClient(Config{BaseURL: server.URL, HTTPClient: server.Client()})
	_, _, err := client.Search(context.Background(), time.Now().Add(-time.Hour), time.Now())
	if err != nil {
		t.Fatalf("Search returned an error after retry: %v", err)
	}
	if calls.Load() != 2 {
		t.Fatalf("expected 2 calls, got %d", calls.Load())
	}
}

func TestSearchRetriesTransientNetworkFailure(t *testing.T) {
	var calls atomic.Int32
	httpClient := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		if calls.Add(1) == 1 {
			return nil, errors.New("net/http: TLS handshake timeout")
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body: io.NopCloser(strings.NewReader(`{"articles":[
				{"url":"https://one.example/a","title":"One"},
				{"url":"https://two.example/b","title":"Two"},
				{"url":"https://three.example/c","title":"Three"}
			]}`)),
		}, nil
	})}

	client := NewGDELTClient(Config{BaseURL: "https://gdelt.example/news", HTTPClient: httpClient})
	_, _, err := client.Search(context.Background(), time.Now().Add(-time.Hour), time.Now())
	if err != nil {
		t.Fatalf("Search returned an error after network retry: %v", err)
	}
	if calls.Load() != 2 {
		t.Fatalf("expected 2 calls, got %d", calls.Load())
	}
}

func TestDefaultClientExtendsTLSHandshakeTimeout(t *testing.T) {
	client := NewGDELTClient(Config{})
	transport, ok := client.httpClient.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("unexpected transport type: %T", client.httpClient.Transport)
	}
	if transport.TLSHandshakeTimeout != 20*time.Second {
		t.Fatalf("unexpected TLS handshake timeout: %s", transport.TLSHandshakeTimeout)
	}
}
