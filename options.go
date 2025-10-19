package dnsdialer

import "time"

// Option is a function that configures a Dialer.
//
// This package uses the functional options pattern, which provides:
// 1. Sensible defaults - you can create a Dialer with just New()
// 2. Flexible configuration - add only the options you need
// 3. Backward compatibility - new options don't break existing code
// 4. Clear intent - each option function name documents what it does
type Option func(*Dialer)

// WithResolvers sets the DNS servers to query.
//
// Each address can be:
// - IP with port: "8.8.8.8:53"
// - IP without port: "8.8.8.8" (port 53 is assumed)
// - Hostname with port: "dns.google:53"
//
// The order matters for the Fallback strategy (tries in order), but not for
// Race (queries all simultaneously) or Consensus (queries all and compares).
//
// Example:
//
//	dialer := New(
//	    WithResolvers("8.8.8.8", "1.1.1.1", "9.9.9.9"),
//	)
func WithResolvers(addrs ...string) Option {
	return func(r *Dialer) {
		for _, addr := range addrs {
			r.resolvers = append(r.resolvers, newUDPResolver(addr, r.timeout, r.poolSize))
		}
	}
}

// WithStrategy sets the resolution strategy.
//
// Available strategies:
//
//   - Race: Query all servers, return first successful response (minimize latency)
//   - Fallback: Try servers in order until one succeeds (ordered failover)
//   - Consensus: Require N servers to agree (detect poisoning/inconsistencies)
//   - Compare: Query all and detect discrepancies (detect poisoning/inconsistencies)
//
// Default is Race if not specified.
//
// Example:
//
//	dialer := New(
//	    WithResolvers("8.8.8.8", "1.1.1.1", "9.9.9.9"),
//	    WithStrategy(Consensus{MinAgreement: 2}),
//	)
func WithStrategy(s Strategy) Option {
	return func(r *Dialer) {
		r.strategy = s
	}
}

// WithTimeout sets the per-query timeout.
//
// This timeout applies to individual DNS queries, not the overall Lookup() call.
// For strategies that query multiple servers (Race, Consensus, Compare), each
// server query gets this timeout. For Fallback, each sequential attempt gets
// this timeout.
//
// Default is 2 seconds if not specified.
//
// Example:
//
//	dialer := New(
//	    WithResolvers("8.8.8.8"),
//	    WithTimeout(5 * time.Second),
//	)
func WithTimeout(d time.Duration) Option {
	return func(r *Dialer) {
		r.timeout = d
	}
}

// WithLogger sets a custom logger for debugging and monitoring.
//
// The logger receives structured log events about query attempts, failures,
// strategy decisions, and performance metrics. Useful for debugging resolution
// issues or monitoring DNS server health.
//
// Default is a no-op logger that discards all log messages.
//
// Example:
//
//	dialer := New(
//	    WithResolvers("8.8.8.8"),
//	    WithLogger(myLogger),
//	)
func WithLogger(l Logger) Option {
	return func(r *Dialer) {
		r.logger = l
	}
}

// WithConnPoolSize sets the maximum number of pooled connections per resolver.
//
// Connection pooling reduces socket creation/destruction overhead. Each DNS server
// gets its own pool. Higher values reduce the chance of creating new connections
// under load, but consume more file descriptors.
//
// Default is 4 connections per resolver if not specified.
//
// Example:
//
//	// High-throughput application
//	dialer := New(
//	    WithResolvers("8.8.8.8"),
//	    WithConnPoolSize(10),
//	)
func WithConnPoolSize(size int) Option {
	return func(r *Dialer) {
		if size > 0 {
			r.poolSize = size
		}
	}
}

// WithCache enables DNS response caching with TTL-aware expiration.
//
// Caching reduces query latency for repeated lookups and load on DNS servers.
// The cache respects DNS TTL values from responses, clamped between minTTL and maxTTL.
//
// Parameters:
//   - size: Maximum number of hostnames to cache (LRU eviction when full)
//   - minTTL: Minimum TTL for cache entries (prevents caching very short TTLs)
//   - maxTTL: Maximum TTL for cache entries (prevents indefinite caching)
//
// The cache mimics OS-level DNS caching behavior while providing explicit control
// over cache size, TTL bounds, and invalidation.
//
// Example:
//
//	// Cache up to 1000 hostnames, with TTL between 1s and 5 minutes
//	dialer := New(
//	    WithResolvers("8.8.8.8", "1.1.1.1"),
//	    WithCache(1000, 1*time.Second, 5*time.Minute),
//	)
func WithCache(size int, minTTL, maxTTL time.Duration) Option {
	return func(r *Dialer) {
		r.cache = newDNSCache(size, minTTL, maxTTL)
	}
}
