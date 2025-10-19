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

// This is just a convenience bench test to report CGO status in benchmarks.
// If the Go compiler is invoked with CGO_ENABLED=1, cgoStatus will be
// "WithCGO". if CGO_ENABLED=0, it will be "NoCGO".
//
// This helps interpret benchmark results when ran.
func BenchmarkCGOStatus_Info(b *testing.B) {
	b.Logf("CGO Status: %s", cgoStatus)
}

func BenchmarkCGO_StdLib_LookupHost(b *testing.B) {
	resolver := &net.Resolver{}

	// To avoid measuring cold-start effects, we do a warmup lookup first.
	// This ensures any internal caches are primed before the benchmark runs.
	//
	// When CGO is enabled, this primes the OS DNS cache (mDNSResponder on
	// macOS, systemd-resolved on Linux), providing a fairer comparison to
	// dnsdialer's internal caching.
	//
	// When CGO is disabled, there isn't any internal caching in the standard
	// library resolver, it always does fresh DNS queries to the configured
	// DNS servers, so at best this primes any DNS server caches upstream.
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

func BenchmarkCGO_DNSDialer_LookupHost(b *testing.B) {
	ctx := context.Background()
	dialer := New(
		WithResolvers(getSystemResolver()),
		WithStrategy(Race{}),
		WithCache(1000, 1*time.Second, 5*time.Minute),
	)

	// Similar to the standard library benchmark, we do a warmup lookup
	// first. This primes dnsdialer's internal cache before the benchmark
	// runs.
	//
	// This ensures we're measuring cached lookups in the benchmark,
	// providing a fair comparison to the standard library benchmark when CGO
	// is enabled (which benefits from OS-level caching).
	//
	// When CGO is disabled, the standard library doesn't cache lookups,
	// which means this benchmark should always be much faster than the
	// standard library benchmark with CGO disabled.
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
