package dnsdialer

import (
	"context"
)

func (s Compare) ResolveType(ctx context.Context, host string, qtype RecordType, resolvers []resolver, logger Logger) ([]Record, error) {
	results := make(map[string][]Record)

	// Query all resolvers and collect successful responses. Unlike Consensus,
	// we don't group by equality yet - we keep track of which resolver returned what
	// so the OnDiscrepancy callback can identify misbehaving resolvers.
	for _, res := range resolvers {
		records, err := res.ResolveType(ctx, host, qtype)
		if err == nil {
			results[res.Name()] = records
		}
		// Ignore errors - Compare is typically used for integrity checking,
		// so we work with whatever responses we get.
	}

	// Check if all successful responses agree. Use the first result as the baseline
	// and compare all others against it.
	var first []Record
	allMatch := true

	for _, records := range results {
		if first == nil {
			first = records
		} else if !recordsEqual(first, records, s.IgnoreTTL) {
			allMatch = false
			break
		}
	}

	if !allMatch {
		// Discrepancy detected. This could indicate:
		// 1. DNS propagation delay (records were recently updated)
		// 2. Geo-based DNS with different responses per region
		// 3. Compromised resolver returning malicious data
		// 4. Misconfigured resolver with stale cache
		//
		// The callback receives all results so the caller can analyze patterns,
		// log details, alert on suspicious activity, etc.
		logger.Info("discrepancy detected in record type query",
			Field{"host", host},
			Field{"type", qtype.String()})
		if s.OnDiscrepancy != nil {
			s.OnDiscrepancy(host, qtype, results)
		}
	}

	// Always return a result (the first one) even if there's a discrepancy.
	// Compare is about detecting differences, not blocking on them. If you need
	// to block on discrepancies, use Consensus instead.
	if first == nil {
		return nil, nil
	}
	return first, nil
}
