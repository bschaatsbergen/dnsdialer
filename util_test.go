// Copyright 2025 Bruno Schaatsbergen. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dnsdialer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRecordsEqual_SameRecords(t *testing.T) {
	records1 := []Record{
		{Value: "192.168.1.1", TTL: 300},
		{Value: "192.168.1.2", TTL: 300},
	}
	records2 := []Record{
		{Value: "192.168.1.1", TTL: 300},
		{Value: "192.168.1.2", TTL: 300},
	}

	assert.True(t, recordsEqual(records1, records2, false))
}

func TestRecordsEqual_DifferentOrder(t *testing.T) {
	records1 := []Record{
		{Value: "192.168.1.1", TTL: 300},
		{Value: "192.168.1.2", TTL: 300},
	}
	records2 := []Record{
		{Value: "192.168.1.2", TTL: 300},
		{Value: "192.168.1.1", TTL: 300},
	}

	assert.True(t, recordsEqual(records1, records2, false))
}

func TestRecordsEqual_DifferentValues(t *testing.T) {
	records1 := []Record{
		{Value: "192.168.1.1", TTL: 300},
	}
	records2 := []Record{
		{Value: "192.168.1.2", TTL: 300},
	}

	assert.False(t, recordsEqual(records1, records2, false))
}

func TestRecordsEqual_DifferentTTL(t *testing.T) {
	records1 := []Record{
		{Value: "192.168.1.1", TTL: 300},
	}
	records2 := []Record{
		{Value: "192.168.1.1", TTL: 600},
	}

	assert.False(t, recordsEqual(records1, records2, false))
}

func TestRecordsEqual_IgnoreTTL(t *testing.T) {
	records1 := []Record{
		{Value: "192.168.1.1", TTL: 300},
	}
	records2 := []Record{
		{Value: "192.168.1.1", TTL: 600},
	}

	assert.True(t, recordsEqual(records1, records2, true))
}

func TestRecordsEqual_DuplicateRecords(t *testing.T) {
	records1 := []Record{
		{Value: "192.168.1.1", TTL: 300},
		{Value: "192.168.1.1", TTL: 300},
		{Value: "192.168.1.2", TTL: 300},
	}
	records2 := []Record{
		{Value: "192.168.1.2", TTL: 300},
		{Value: "192.168.1.1", TTL: 300},
		{Value: "192.168.1.1", TTL: 300},
	}

	assert.True(t, recordsEqual(records1, records2, false))
}

func TestRecordsEqual_DifferentDuplicateCount(t *testing.T) {
	records1 := []Record{
		{Value: "192.168.1.1", TTL: 300},
		{Value: "192.168.1.1", TTL: 300},
	}
	records2 := []Record{
		{Value: "192.168.1.1", TTL: 300},
	}

	assert.False(t, recordsEqual(records1, records2, false))
}

func TestRecordsEqual_EmptySlices(t *testing.T) {
	var records1 []Record
	var records2 []Record

	assert.True(t, recordsEqual(records1, records2, false))
}

func TestRecordsEqual_OneEmpty(t *testing.T) {
	records1 := []Record{
		{Value: "192.168.1.1", TTL: 300},
	}
	var records2 []Record

	assert.False(t, recordsEqual(records1, records2, false))
}

func TestRecordsEqual_DifferentLengths(t *testing.T) {
	records1 := []Record{
		{Value: "192.168.1.1", TTL: 300},
		{Value: "192.168.1.2", TTL: 300},
	}
	records2 := []Record{
		{Value: "192.168.1.1", TTL: 300},
	}

	assert.False(t, recordsEqual(records1, records2, false))
}
