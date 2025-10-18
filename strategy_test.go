package dnsdialer

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// mockResolver implements the resolver interface for testing
type mockResolver struct {
	name     string
	response []Record
	err      error
	delay    time.Duration
}

func (m *mockResolver) Name() string {
	return m.name
}

func (m *mockResolver) ResolveType(ctx context.Context, host string, qtype RecordType) ([]Record, error) {
	if m.delay > 0 {
		select {
		case <-time.After(m.delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	return m.response, m.err
}

func (m *mockResolver) String() string {
	return m.name
}

// mockLogger for testing
type mockLogger struct {
	logs []string
}

func (m *mockLogger) Debug(msg string, fields ...Field) {
	m.logs = append(m.logs, "DEBUG: "+msg)
}

func (m *mockLogger) Info(msg string, fields ...Field) {
	m.logs = append(m.logs, "INFO: "+msg)
}

func (m *mockLogger) Error(msg string, err error, fields ...Field) {
	logMsg := "ERROR: " + msg
	if err != nil {
		logMsg += " (" + err.Error() + ")"
	}
	m.logs = append(m.logs, logMsg)
}

func TestRace_FirstSuccessfulResponse(t *testing.T) {
	ctx := context.Background()
	logger := &mockLogger{}

	resolvers := []resolver{
		&mockResolver{name: "slow", response: []Record{{Value: "1.1.1.1", TTL: 300}}, delay: 100 * time.Millisecond},
		&mockResolver{name: "fast", response: []Record{{Value: "2.2.2.2", TTL: 300}}, delay: 10 * time.Millisecond},
	}

	strategy := Race{}
	records, err := strategy.ResolveType(ctx, "example.com", TypeA, resolvers, logger)

	assert.NoError(t, err)
	assert.Len(t, records, 1)
	assert.Equal(t, "2.2.2.2", records[0].Value)
}

func TestRace_AllResolversFail(t *testing.T) {
	ctx := context.Background()
	logger := &mockLogger{}

	resolvers := []resolver{
		&mockResolver{name: "resolver1", err: errors.New("timeout")},
		&mockResolver{name: "resolver2", err: errors.New("server failure")},
	}

	strategy := Race{}
	records, err := strategy.ResolveType(ctx, "example.com", TypeA, resolvers, logger)

	assert.Error(t, err)
	assert.Nil(t, records)
}

func TestConsensus_MinorityAgreement_Fails(t *testing.T) {
	ctx := context.Background()
	logger := &mockLogger{}

	resolvers := []resolver{
		&mockResolver{name: "resolver1", response: []Record{{Value: "1.1.1.1", TTL: 300}}},
		&mockResolver{name: "resolver2", response: []Record{{Value: "2.2.2.2", TTL: 300}}},
		&mockResolver{name: "resolver3", response: []Record{{Value: "3.3.3.3", TTL: 300}}},
	}

	strategy := Consensus{MinAgreement: 2}
	records, err := strategy.ResolveType(ctx, "example.com", TypeA, resolvers, logger)

	assert.Error(t, err)
	assert.Nil(t, records)
}

func TestConsensus_MajorityAgreement_Succeeds(t *testing.T) {
	ctx := context.Background()
	logger := &mockLogger{}

	resolvers := []resolver{
		&mockResolver{name: "resolver1", response: []Record{{Value: "1.1.1.1", TTL: 300}}},
		&mockResolver{name: "resolver2", response: []Record{{Value: "1.1.1.1", TTL: 300}}},
		&mockResolver{name: "resolver3", response: []Record{{Value: "2.2.2.2", TTL: 300}}},
	}

	strategy := Consensus{MinAgreement: 2}
	records, err := strategy.ResolveType(ctx, "example.com", TypeA, resolvers, logger)

	assert.NoError(t, err)
	assert.Len(t, records, 1)
	assert.Equal(t, "1.1.1.1", records[0].Value)
}

func TestConsensus_IgnoreTTL(t *testing.T) {
	ctx := context.Background()
	logger := &mockLogger{}

	resolvers := []resolver{
		&mockResolver{name: "resolver1", response: []Record{{Value: "1.1.1.1", TTL: 300}}},
		&mockResolver{name: "resolver2", response: []Record{{Value: "1.1.1.1", TTL: 600}}},
	}

	strategy := Consensus{MinAgreement: 2, IgnoreTTL: true}
	records, err := strategy.ResolveType(ctx, "example.com", TypeA, resolvers, logger)

	assert.NoError(t, err)
	assert.Len(t, records, 1)
	assert.Equal(t, "1.1.1.1", records[0].Value)
}

func TestConsensus_DefaultMinAgreement(t *testing.T) {
	ctx := context.Background()
	logger := &mockLogger{}

	// 3 resolvers, default should be (3/2)+1 = 2
	resolvers := []resolver{
		&mockResolver{name: "resolver1", response: []Record{{Value: "1.1.1.1", TTL: 300}}},
		&mockResolver{name: "resolver2", response: []Record{{Value: "1.1.1.1", TTL: 300}}},
		&mockResolver{name: "resolver3", response: []Record{{Value: "2.2.2.2", TTL: 300}}},
	}

	strategy := Consensus{} // MinAgreement: 0 should default to majority
	records, err := strategy.ResolveType(ctx, "example.com", TypeA, resolvers, logger)

	assert.NoError(t, err)
	assert.Len(t, records, 1)
	assert.Equal(t, "1.1.1.1", records[0].Value)
}

func TestFallback_FirstResolverSucceeds(t *testing.T) {
	ctx := context.Background()
	logger := &mockLogger{}

	resolvers := []resolver{
		&mockResolver{name: "resolver1", response: []Record{{Value: "1.1.1.1", TTL: 300}}},
		&mockResolver{name: "resolver2", response: []Record{{Value: "2.2.2.2", TTL: 300}}},
	}

	strategy := Fallback{}
	records, err := strategy.ResolveType(ctx, "example.com", TypeA, resolvers, logger)

	assert.NoError(t, err)
	assert.Len(t, records, 1)
	assert.Equal(t, "1.1.1.1", records[0].Value)
}

func TestFallback_FirstResolverFails(t *testing.T) {
	ctx := context.Background()
	logger := &mockLogger{}

	resolvers := []resolver{
		&mockResolver{name: "resolver1", err: errors.New("timeout")},
		&mockResolver{name: "resolver2", response: []Record{{Value: "2.2.2.2", TTL: 300}}},
	}

	strategy := Fallback{}
	records, err := strategy.ResolveType(ctx, "example.com", TypeA, resolvers, logger)

	assert.NoError(t, err)
	assert.Len(t, records, 1)
	assert.Equal(t, "2.2.2.2", records[0].Value)
}

func TestFallback_AllResolversFail(t *testing.T) {
	ctx := context.Background()
	logger := &mockLogger{}

	resolvers := []resolver{
		&mockResolver{name: "resolver1", err: errors.New("timeout")},
		&mockResolver{name: "resolver2", err: errors.New("server failure")},
	}

	strategy := Fallback{}
	records, err := strategy.ResolveType(ctx, "example.com", TypeA, resolvers, logger)

	assert.Error(t, err)
	assert.Nil(t, records)
}

func TestCompare_NoDiscrepancy(t *testing.T) {
	ctx := context.Background()
	logger := &mockLogger{}
	var discrepancyCalled bool

	resolvers := []resolver{
		&mockResolver{name: "resolver1", response: []Record{{Value: "1.1.1.1", TTL: 300}}},
		&mockResolver{name: "resolver2", response: []Record{{Value: "1.1.1.1", TTL: 300}}},
	}

	strategy := Compare{
		OnDiscrepancy: func(host string, qtype RecordType, results map[string][]Record) {
			discrepancyCalled = true
		},
	}

	records, err := strategy.ResolveType(ctx, "example.com", TypeA, resolvers, logger)

	assert.NoError(t, err)
	assert.Len(t, records, 1)
	assert.Equal(t, "1.1.1.1", records[0].Value)
	assert.False(t, discrepancyCalled)
}

func TestCompare_WithDiscrepancy(t *testing.T) {
	ctx := context.Background()
	logger := &mockLogger{}
	var discrepancyCalled bool
	var discrepancyHost string

	resolvers := []resolver{
		&mockResolver{name: "resolver1", response: []Record{{Value: "1.1.1.1", TTL: 300}}},
		&mockResolver{name: "resolver2", response: []Record{{Value: "2.2.2.2", TTL: 300}}},
	}

	strategy := Compare{
		OnDiscrepancy: func(host string, qtype RecordType, results map[string][]Record) {
			discrepancyCalled = true
			discrepancyHost = host
		},
	}

	records, err := strategy.ResolveType(ctx, "example.com", TypeA, resolvers, logger)

	assert.NoError(t, err)
	assert.Len(t, records, 1) // Returns first successful result
	assert.True(t, discrepancyCalled)
	assert.Equal(t, "example.com", discrepancyHost)
}

func TestCompare_IgnoreTTLDifferences(t *testing.T) {
	ctx := context.Background()
	logger := &mockLogger{}
	var discrepancyCalled bool

	resolvers := []resolver{
		&mockResolver{name: "resolver1", response: []Record{{Value: "1.1.1.1", TTL: 300}}},
		&mockResolver{name: "resolver2", response: []Record{{Value: "1.1.1.1", TTL: 600}}},
	}

	strategy := Compare{
		IgnoreTTL: true,
		OnDiscrepancy: func(host string, qtype RecordType, results map[string][]Record) {
			discrepancyCalled = true
		},
	}

	records, err := strategy.ResolveType(ctx, "example.com", TypeA, resolvers, logger)

	assert.NoError(t, err)
	assert.Len(t, records, 1)
	assert.Equal(t, "1.1.1.1", records[0].Value)
	assert.False(t, discrepancyCalled) // No discrepancy when ignoring TTL
}

func TestCompare_NoOnDiscrepancyCallback(t *testing.T) {
	ctx := context.Background()
	logger := &mockLogger{}

	resolvers := []resolver{
		&mockResolver{name: "resolver1", response: []Record{{Value: "1.1.1.1", TTL: 300}}},
		&mockResolver{name: "resolver2", response: []Record{{Value: "2.2.2.2", TTL: 300}}},
	}

	strategy := Compare{} // No OnDiscrepancy callback

	// Should not panic when callback is nil
	records, err := strategy.ResolveType(ctx, "example.com", TypeA, resolvers, logger)

	assert.NoError(t, err)
	assert.Len(t, records, 1)
}
