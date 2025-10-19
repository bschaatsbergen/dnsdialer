// Copyright 2025 Bruno Schaatsbergen. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dnsdialer

import (
	"context"
)

func (s Fallback) ResolveType(ctx context.Context, host string, qtype RecordType, resolvers []resolver, logger Logger) ([]Record, error) {
	var lastErr error

	// Try each resolver in order until one succeeds. This provides ordered failover,
	// useful when you have a preferred resolver (e.g., internal DNS) but want to fall
	// back to alternatives (e.g., public DNS) if it's unavailable.
	//
	// Unlike Race, this minimizes network traffic by only querying one resolver at a time.
	// The trade-off is higher latency if early resolvers in the list are slow or down.
	for _, res := range resolvers {
		records, err := res.ResolveType(ctx, host, qtype)
		if err == nil {
			logger.Debug("resolver succeeded",
				Field{"resolver", res.Name()},
				Field{"type", qtype.String()})
			return records, nil
		}
		// Keep trying the remaining resolvers. The error might be temporary like a timeout
		// or network issue, or permanent like domain doesn't exist. We can't really distinguish,
		// so we just try all resolvers before giving up.
		lastErr = err
		logger.Debug("resolver failed, trying next",
			Field{"resolver", res.Name()},
			Field{"type", qtype.String()},
			Field{"error", err.Error()})
	}

	// All resolvers failed. Return the last error, which may not be the most informative
	// one, but it's from the last fallback option we tried. Consider logging all errors
	// if you're debugging why all resolvers failed.
	return nil, lastErr
}
