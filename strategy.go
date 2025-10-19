// Copyright 2025 Bruno Schaatsbergen. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dnsdialer

import (
	"context"
)

// Strategy determines how to coordinate DNS queries across multiple resolvers.
//
// All strategies implement this interface, allowing them to be swapped without
// changing the calling code. The strategy receives all configured resolvers and
// decides how to query them (sequentially, concurrently, with consensus, etc.).
type Strategy interface {
	ResolveType(ctx context.Context, host string, qtype RecordType, resolvers []resolver, logger Logger) ([]Record, error)
}

// Race queries all resolvers simultaneously and returns the first successful response.
type Race struct{}

// Consensus requires a minimum number of resolvers to return identical results.
type Consensus struct {
	// MinAgreement is the minimum number of resolvers that must return identical results.
	// If 0, defaults to simple majority: (n/2)+1. For Byzantine fault tolerance,
	// ensure at most (MinAgreement-1) resolvers can be compromised.
	MinAgreement int

	// IgnoreTTL, when true, treats records as equal if values match regardless of TTL differences.
	// Useful because TTLs naturally decay at different rates across resolvers depending on
	// when they cached the record.
	IgnoreTTL bool
}

// Fallback tries resolvers sequentially in order until one succeeds.
type Fallback struct{}

// Compare queries all resolvers and detects discrepancies without failing on them.
type Compare struct {
	// OnDiscrepancy is an optional callback invoked when resolvers return different results.
	OnDiscrepancy func(host string, qtype RecordType, results map[string][]Record)

	// IgnoreTTL, when true, means only values are compared (TTL differences don't trigger discrepancy).
	IgnoreTTL bool
}
