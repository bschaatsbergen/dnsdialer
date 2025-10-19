// Copyright 2025 Bruno Schaatsbergen. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dnsdialer

import (
	"context"
	"net"
	"testing"
	"time"
)

// BenchmarkCGOStatus_Info displays the current CGO configuration.
func BenchmarkCGOStatus_Info(b *testing.B) {
	b.Logf("CGO Status: %s", cgoStatus)
}

// BenchmarkCGO_StdLib_LookupHost benchmarks the standard library's DNS lookup.
// With CGO enabled (default), uses getaddrinfo() and OS caching (mDNSResponder/systemd-resolved).
// With CGO disabled, uses pure Go DNS resolver.
func BenchmarkCGO_StdLib_LookupHost(b *testing.B) {
	resolver := &net.Resolver{}

	// Warmup to prime cache (OS cache with CGO, or DNS server cache without CGO)
	if _, err := resolver.LookupHost(context.Background(), "www.google.com"); err != nil {
		b.Fatalf("warmup failed: %v", err)
	}

	b.ResetTimer()
	for b.Loop() {
		if _, err := resolver.LookupHost(context.Background(), "www.google.com"); err != nil {
			b.Fatalf("lookup failed: %v", err)
		}
	}
}

// BenchmarkCGO_DNSDialer_LookupHost benchmarks dnsdialer's DNS lookup with in-memory cache.
// Cache behavior is identical regardless of CGO status.
func BenchmarkCGO_DNSDialer_LookupHost(b *testing.B) {
	ctx := context.Background()
	dialer := New(
		WithResolvers(getSystemResolver()),
		WithStrategy(Race{}),
		WithCache(1000, 1*time.Second, 5*time.Minute),
	)

	// Warmup to prime in-memory LRU cache
	if _, err := dialer.lookupIPs(ctx, "www.google.com"); err != nil {
		b.Fatalf("warmup failed: %v", err)
	}

	b.ResetTimer()
	for b.Loop() {
		if _, err := dialer.lookupIPs(ctx, "www.google.com"); err != nil {
			b.Fatalf("lookup failed: %v", err)
		}
	}
}
