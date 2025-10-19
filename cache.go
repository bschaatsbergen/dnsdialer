// Copyright 2025 Bruno Schaatsbergen. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dnsdialer

import (
	"net"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru/v2/expirable"
)

// ipCacheEntry holds cached IP addresses with their expiration time. We cache the
// already-parsed net.IP values here so we can skip the net.ParseIP overhead on every
// cache hit, which saves a surprising amount of time.
type ipCacheEntry struct {
	ips       []net.IP
	expiresAt time.Time
}

// isExpired checks if the IP cache entry has expired based on DNS TTL.
func (e *ipCacheEntry) isExpired() bool {
	return time.Now().After(e.expiresAt)
}

// dnsCache wraps an LRU cache with TTL-aware expiration for IP addresses. It mimics
// OS-level DNS caching behavior (mDNSResponder, systemd-resolved) while providing
// explicit control over cache size, TTL bounds, and invalidation.
type dnsCache struct {
	ipCache *lru.LRU[string, *ipCacheEntry]
	mu      sync.RWMutex
	enabled bool

	// minTTL prevents caching entries with very short TTLs that would just thrash the cache.
	// For example, setting this to 1s means we won't bother caching a record with TTL=0.
	minTTL time.Duration

	// maxTTL caps how long we'll cache an entry, regardless of what the DNS server tells us.
	// This ensures we periodically re-validate even if the server sends a very high TTL.
	maxTTL time.Duration
}

// newDNSCache creates a new DNS cache with the specified size and TTL bounds.
// Size controls the maximum number of hostnames to cache (LRU eviction when full).
// minTTL and maxTTL clamp DNS response TTLs to prevent both cache thrashing from
// very short TTLs and indefinite caching from very long TTLs.
func newDNSCache(size int, minTTL, maxTTL time.Duration) *dnsCache {
	if size <= 0 {
		return &dnsCache{enabled: false}
	}

	// Create LRU cache for IP addresses. The golang-lru library handles eviction
	// and basic TTL tracking for us, but we also check expiration manually in getIPs()
	// since we want to respect DNS TTLs from individual records.
	ipCache := lru.NewLRU[string, *ipCacheEntry](size, nil, maxTTL)

	return &dnsCache{
		ipCache: ipCache,
		enabled: true,
		minTTL:  minTTL,
		maxTTL:  maxTTL,
	}
}

// getIPs retrieves cached IP addresses for a hostname if they exist and haven't expired.
// This is the fast path for lookupIPs() and is crucial for performance. By caching
// parsed net.IP values instead of DNS records, we avoid calling net.ParseIP on every
// cache hit.
func (c *dnsCache) getIPs(host string) []net.IP {
	if !c.enabled {
		return nil
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.ipCache.Get(host)
	if !ok {
		return nil
	}

	// Same expiration logic as the record cache, don't bother removing it, just return
	// nil to signal a cache miss. The LRU will evict it eventually.
	if entry.isExpired() {
		return nil
	}

	// Return a copy to prevent the caller from modifying our cached data. net.IP is a
	// slice, so we need to copy the slice itself, not just the individual IP values.
	ips := make([]net.IP, len(entry.ips))
	copy(ips, entry.ips)
	return ips
}

// setIPs stores already-parsed IP addresses in the cache with TTL-based expiration.
// The TTL is passed in from the caller who has already figured out the minimum TTL
// from the DNS response records. We just need to clamp it to our configured bounds.
func (c *dnsCache) setIPs(host string, ips []net.IP, ttl time.Duration) {
	if !c.enabled || len(ips) == 0 {
		return
	}

	// Clamp TTL to our configured bounds, don't trust DNS servers too much.
	if ttl < c.minTTL {
		ttl = c.minTTL
	}
	if ttl > c.maxTTL {
		ttl = c.maxTTL
	}

	entry := &ipCacheEntry{
		ips:       ips,
		expiresAt: time.Now().Add(ttl),
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.ipCache.Add(host, entry)
}
