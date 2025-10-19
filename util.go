// Copyright 2025 Bruno Schaatsbergen. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dnsdialer

// recordKey is used as a map key for comparing DNS records.
// It combines value and TTL to enable multiset equality checking.
type recordKey struct {
	value string
	ttl   uint32
}

// recordsEqual checks if two slices of DNS records are equal, treating them as multisets.
//
// This means:
//   - Order doesn't matter: [A, B] equals [B, A]
//   - Duplicates matter: [A, A, B] does not equal [A, B]
//   - When ignoreTTL is true, records are compared only by value (useful since TTLs
//     naturally decay over time and can differ between resolvers even for the same data)
//
// The algorithm uses a frequency map to handle duplicates correctly. This is necessary
// because DNS responses can contain the same record multiple times (e.g., multiple A
// records for round-robin load balancing).
func recordsEqual(a, b []Record, ignoreTTL bool) bool {
	// Fast path: if lengths differ, they can't be equal
	if len(a) != len(b) {
		return false
	}

	// Build a frequency map for slice 'a'. This counts how many times each
	// unique record appears. For example, if 'a' contains [X, X, Y], the map
	// will be {X: 2, Y: 1}.
	aMap := make(map[recordKey]int)
	for _, r := range a {
		key := recordKey{
			value: r.Value,
			ttl:   r.TTL,
		}
		if ignoreTTL {
			// Normalize TTL to 0 when comparing. This treats records with different
			// TTLs but the same value as equal. Important because:
			// 1. TTLs count down independently at each resolver
			// 2. Resolvers may have cached the record at different times
			// 3. We care about "is this the same data" not "same data with exact same TTL"
			key.ttl = 0
		}
		aMap[key]++
	}

	// Check that slice 'b' has the exact same frequency of each record.
	// For each record in 'b', decrement its count in the map. If we encounter
	// a record that's not in the map or has count 0, the slices aren't equal.
	for _, r := range b {
		key := recordKey{
			value: r.Value,
			ttl:   r.TTL,
		}
		if ignoreTTL {
			key.ttl = 0
		}
		count, exists := aMap[key]
		if !exists || count == 0 {
			// Either this record isn't in 'a', or 'b' has more copies of it than 'a' does
			return false
		}
		aMap[key]--
	}

	// If we get here, both slices contain the same records with the same frequencies
	return true
}
