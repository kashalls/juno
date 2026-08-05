// Package appicons resolves Discord application icons via Discord's public,
// unauthenticated GET /applications/{id}/rpc endpoint, and caches the
// result so repeated lookups for the same application don't re-hit
// Discord. Unlike the profile package, this is looked up on demand: the
// endpoint accepts any application ID from any caller, not just a single
// tracked user, so there's nothing to background-refresh.
package appicons

import (
	"container/list"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"
)

const (
	rpcEndpointFmt  = "https://discord.com/api/v10/applications/%s/rpc"
	cdnIconURLFmt   = "https://cdn.discordapp.com/app-icons/%s/%s.png"
	maxCacheEntries = 512
	requestTimeout  = 5 * time.Second
)

// ErrNoIcon indicates the application exists but has no icon set.
var ErrNoIcon = errors.New("application has no icon")

// Resolver looks up application icon CDN URLs, backed by a bounded
// least-recently-used cache keyed by application ID so an unbounded stream
// of distinct (or bogus) IDs from public callers can't grow memory forever.
type Resolver struct {
	client *http.Client

	mu    sync.Mutex
	cache map[string]*list.Element
	order *list.List
}

type cacheEntry struct {
	appID string
	url   string
}

func NewResolver() *Resolver {
	return &Resolver{
		client: &http.Client{Timeout: requestTimeout},
		cache:  make(map[string]*list.Element),
		order:  list.New(),
	}
}

// IconURL returns the CDN URL for an application's icon.
func (r *Resolver) IconURL(ctx context.Context, appID string) (string, error) {
	if url, ok := r.get(appID); ok {
		return url, nil
	}

	url, err := r.fetch(ctx, appID)
	if err != nil {
		return "", err
	}

	r.put(appID, url)
	return url, nil
}

func (r *Resolver) fetch(ctx context.Context, appID string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf(rpcEndpointFmt, appID), nil)
	if err != nil {
		return "", err
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return "", ErrNoIcon
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("discord rpc lookup for application %s failed: status %d", appID, resp.StatusCode)
	}

	var rpc struct {
		Icon string `json:"icon"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rpc); err != nil {
		return "", err
	}
	if rpc.Icon == "" {
		return "", ErrNoIcon
	}

	return fmt.Sprintf(cdnIconURLFmt, appID, rpc.Icon), nil
}

func (r *Resolver) get(appID string) (string, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	el, ok := r.cache[appID]
	if !ok {
		return "", false
	}
	r.order.MoveToFront(el)
	return el.Value.(*cacheEntry).url, true
}

func (r *Resolver) put(appID, url string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if el, ok := r.cache[appID]; ok {
		el.Value.(*cacheEntry).url = url
		r.order.MoveToFront(el)
		return
	}

	r.cache[appID] = r.order.PushFront(&cacheEntry{appID: appID, url: url})

	if r.order.Len() > maxCacheEntries {
		oldest := r.order.Back()
		r.order.Remove(oldest)
		delete(r.cache, oldest.Value.(*cacheEntry).appID)
	}
}
