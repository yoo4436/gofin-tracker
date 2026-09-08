package newsfeed

import (
	"context"
	"strings"
	"sync"
	"time"
)

type CachedProvider struct {
	provider  Provider
	ttl       time.Duration
	mu        sync.RWMutex
	latest    Feed
	expiresAt time.Time
}

func NewCachedProvider(provider Provider, ttl time.Duration) *CachedProvider {
	return &CachedProvider{provider: provider, ttl: ttl}
}

func (p *CachedProvider) Fetch(ctx context.Context, query string) (Feed, error) {
	if strings.TrimSpace(query) != "" || p.ttl <= 0 {
		return p.provider.Fetch(ctx, query)
	}

	now := time.Now()
	p.mu.RLock()
	if now.Before(p.expiresAt) {
		feed := p.latest
		p.mu.RUnlock()
		return feed, nil
	}
	p.mu.RUnlock()

	feed, err := p.provider.Fetch(ctx, "")
	if err != nil {
		return Feed{}, err
	}
	p.mu.Lock()
	p.latest = feed
	p.expiresAt = now.Add(p.ttl)
	p.mu.Unlock()
	return feed, nil
}
