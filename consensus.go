package dnsdialer

import (
	"context"
	"fmt"
)

func (s Consensus) ResolveType(ctx context.Context, host string, qtype RecordType, resolvers []resolver, logger Logger) ([]Record, error) {
	// Default to simple majority if not specified. For 3 resolvers, need 2 to agree.
	// For 4 resolvers, need 3. This provides Byzantine fault tolerance assuming
	// at most (n-1)/2 resolvers are compromised or malfunctioning.
	if s.MinAgreement <= 0 {
		s.MinAgreement = (len(resolvers) / 2) + 1
	}

	type resultGroup struct {
		records []Record
		count   int
	}

	var groups []resultGroup

	// Query all resolvers and group responses by equality. We don't race or short-circuit
	// because we need to collect enough responses to reach consensus. This is inherently
	// slower than Race but provides security against DNS poisoning or compromised resolvers.
	for _, res := range resolvers {
		records, err := res.ResolveType(ctx, host, qtype)
		if err != nil {
			// Skip failed queries. Note that if too many fail, we won't reach consensus.
			// For example, with 3 resolvers and MinAgreement=2, if one fails we can still
			// succeed if the other 2 agree. But if 2 fail, we'll always fail to reach consensus.
			continue
		}

		// Check if these records match any existing group. Records are considered equal
		// if they contain the same values (and optionally same TTLs, depending on IgnoreTTL).
		matched := false
		for i := range groups {
			if recordsEqual(groups[i].records, records, s.IgnoreTTL) {
				groups[i].count++
				matched = true
				break
			}
		}

		// No matching group found, create a new one. This happens when a resolver
		// returns different data (could indicate DNS poisoning, misconfiguration,
		// or normal DNS propagation delay).
		if !matched {
			groups = append(groups, resultGroup{
				records: records,
				count:   1,
			})
		}
	}

	// Return the first group that has sufficient agreement. If multiple groups
	// reach the threshold (shouldn't happen with proper MinAgreement setting),
	// we return the first one encountered.
	for _, group := range groups {
		if group.count >= s.MinAgreement {
			logger.Debug("consensus reached",
				Field{"agreements", group.count},
				Field{"required", s.MinAgreement},
				Field{"type", qtype.String()})
			return group.records, nil
		}
	}

	// No consensus reached. This could mean:
	// 1. Too many resolvers failed to respond
	// 2. Resolvers returned different data and no group reached MinAgreement
	// 3. Active DNS poisoning attack with responses split across multiple values
	return nil, fmt.Errorf("consensus not reached: required %d agreements", s.MinAgreement)
}
