// Copyright 2025 Bruno Schaatsbergen. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dnsdialer

import (
	"context"
	"time"
)

func (s Race) ResolveType(ctx context.Context, host string, qtype RecordType, resolvers []resolver, logger Logger) ([]Record, error) {
	// Create a cancellable context so we can stop in-flight queries once we get
	// a successful response. This prevents unnecessary network traffic and reduces
	// load on DNS servers.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	type result struct {
		records  []Record
		err      error
		resolver string
		latency  time.Duration
	}

	// Buffered channel sized to number of resolvers ensures no goroutine blocks
	// when sending results, even if we've already returned to the caller.
	results := make(chan result, len(resolvers))

	// Launch all queries simultaneously. The idea here is to leverage whichever
	// resolver is fastest, whoever responds first wins. This minimizes latency
	// at the cost of increased network traffic, all servers get queried even though
	// we only use one response.
	for _, res := range resolvers {
		go func(r resolver) {
			start := time.Now()
			records, err := r.ResolveType(ctx, host, qtype)
			results <- result{
				records:  records,
				err:      err,
				resolver: r.Name(),
				latency:  time.Since(start),
			}
		}(res)
	}

	// Return the first successful response. We have to wait for all resolvers to
	// either succeed or fail before giving up, since early failures from fast-but-broken
	// resolvers shouldn't prevent us from getting results from slower-but-working ones.
	var lastErr error
	for i := 0; i < len(resolvers); i++ {
		r := <-results
		if r.err == nil {
			logger.Debug("resolver won race",
				Field{"resolver", r.resolver},
				Field{"latency", r.latency},
				Field{"type", qtype.String()})
			// Cancel outstanding queries to avoid wasting resources. Note: UDP queries
			// may have already been sent, but this at least prevents us from waiting
			// for responses we don't need anymore.
			cancel()
			return r.records, nil
		}
		lastErr = r.err
	}

	// All resolvers failed. Return the last error we encountered. In a production
	// system, you might want to aggregate all errors to help diagnose whether one
	// resolver is down vs. the domain doesn't exist.
	return nil, lastErr
}
