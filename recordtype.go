// Copyright 2025 Bruno Schaatsbergen. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dnsdialer

import (
	"fmt"

	"github.com/miekg/dns"
)

// RecordType represents a DNS record type
type RecordType uint16

const (
	// TypeA represents an IPv4 address record
	TypeA RecordType = RecordType(dns.TypeA)
	// TypeAAAA represents an IPv6 address record
	TypeAAAA RecordType = RecordType(dns.TypeAAAA)
	// TypeCNAME represents a canonical name record
	TypeCNAME RecordType = RecordType(dns.TypeCNAME)
	// TypeMX represents a mail exchange record
	TypeMX RecordType = RecordType(dns.TypeMX)
	// TypeNS represents a name server record
	TypeNS RecordType = RecordType(dns.TypeNS)
	// TypeTXT represents a text record
	TypeTXT RecordType = RecordType(dns.TypeTXT)
	// TypeSOA represents a start of authority record
	TypeSOA RecordType = RecordType(dns.TypeSOA)
	// TypePTR represents a pointer record
	TypePTR RecordType = RecordType(dns.TypePTR)
	// TypeSRV represents a service record
	TypeSRV RecordType = RecordType(dns.TypeSRV)
)

// String returns the string representation of the record type
func (rt RecordType) String() string {
	return dns.TypeToString[uint16(rt)]
}

// Record represents a DNS record with its value
type Record struct {
	Type  RecordType
	Value string
	TTL   uint32
}

// String returns a string representation of the record
func (r Record) String() string {
	return fmt.Sprintf("%s: %s (TTL: %d)", r.Type.String(), r.Value, r.TTL)
}
