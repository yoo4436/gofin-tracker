package newsfeed

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

type countingProvider struct {
	calls atomic.Int32
}

func (p *countingProvider) Fetch(_ context.Context, query string) (Feed, error) {
	p.calls.Add(1)
	return Feed{Query: query, Articles: []Article{}}, nil
}

func TestCachedProviderCachesLatestButNotSearch(t *testing.T) {
	provider := &countingProvider{}
	cached := NewCachedProvider(provider, time.Minute)

	if _, err := cached.Fetch(context.Background(), ""); err != nil {
		t.Fatal(err)
	}
	if _, err := cached.Fetch(context.Background(), ""); err != nil {
		t.Fatal(err)
	}
	if got := provider.calls.Load(); got != 1 {
		t.Fatalf("latest feed calls = %d, want 1", got)
	}

	if _, err := cached.Fetch(context.Background(), "AI"); err != nil {
		t.Fatal(err)
	}
	if _, err := cached.Fetch(context.Background(), "AI"); err != nil {
		t.Fatal(err)
	}
	if got := provider.calls.Load(); got != 3 {
		t.Fatalf("all provider calls = %d, want 3", got)
	}
}
