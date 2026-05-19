// Copyright 2025 Bruno Schaatsbergen. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dnsdialer

import (
	"context"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGooglePublicDNSv4_ValidAddresses(t *testing.T) {
	assert.Len(t, GooglePublicDNSv4, 2)
	assert.Equal(t, Resolver("8.8.8.8:53"), GooglePublicDNSv4[0])
	assert.Equal(t, Resolver("8.8.4.4:53"), GooglePublicDNSv4[1])
}

func TestGooglePublicDNSv6_ValidAddresses(t *testing.T) {
	assert.Len(t, GooglePublicDNSv6, 2)
	assert.Equal(t, Resolver("[2001:4860:4860::8888]:53"), GooglePublicDNSv6[0])
	assert.Equal(t, Resolver("[2001:4860:4860::8844]:53"), GooglePublicDNSv6[1])
}

func TestGooglePublicDNSv4_CanResolve(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resolver := newUDPResolver(string(GooglePublicDNSv4[0]), 5*time.Second, 1)
	records, err := resolver.ResolveType(ctx, "www.google.com", TypeA)
	assert.NoError(t, err)
	assert.NotEmpty(t, records)
}

func TestGooglePublicDNSv6_CanResolve(t *testing.T) {
	if os.Getenv("TEST_IPV6") == "" {
		t.Skip("Skipping IPv6 test (set TEST_IPV6=1 to enable)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resolver := newUDPResolver(string(GooglePublicDNSv6[0]), 5*time.Second, 1)
	records, err := resolver.ResolveType(ctx, "www.google.com", TypeAAAA)
	assert.NoError(t, err)
	assert.NotEmpty(t, records)
}

func TestCloudflareDNSv4_ValidAddresses(t *testing.T) {
	assert.Len(t, CloudflareDNSv4, 2)
	assert.Equal(t, Resolver("1.1.1.1:53"), CloudflareDNSv4[0])
	assert.Equal(t, Resolver("1.0.0.1:53"), CloudflareDNSv4[1])
}

func TestCloudflareDNSv6_ValidAddresses(t *testing.T) {
	assert.Len(t, CloudflareDNSv6, 2)
	assert.Equal(t, Resolver("[2606:4700:4700::1111]:53"), CloudflareDNSv6[0])
	assert.Equal(t, Resolver("[2606:4700:4700::1001]:53"), CloudflareDNSv6[1])
}

func TestCloudflareDNSv4_CanResolve(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resolver := newUDPResolver(string(CloudflareDNSv4[0]), 5*time.Second, 1)
	records, err := resolver.ResolveType(ctx, "www.google.com", TypeA)
	assert.NoError(t, err)
	assert.NotEmpty(t, records)
}

func TestCloudflareDNSv6_CanResolve(t *testing.T) {
	if os.Getenv("TEST_IPV6") == "" {
		t.Skip("Skipping IPv6 test (set TEST_IPV6=1 to enable)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resolver := newUDPResolver(string(CloudflareDNSv6[0]), 5*time.Second, 1)
	records, err := resolver.ResolveType(ctx, "www.google.com", TypeAAAA)
	assert.NoError(t, err)
	assert.NotEmpty(t, records)
}

func TestQuad9DNSv4_ValidAddresses(t *testing.T) {
	assert.Len(t, Quad9DNSv4, 2)
	assert.Equal(t, Resolver("9.9.9.9:53"), Quad9DNSv4[0])
	assert.Equal(t, Resolver("149.112.112.112:53"), Quad9DNSv4[1])
}

func TestQuad9DNSv6_ValidAddresses(t *testing.T) {
	assert.Len(t, Quad9DNSv6, 2)
	assert.Equal(t, Resolver("[2620:fe::fe]:53"), Quad9DNSv6[0])
	assert.Equal(t, Resolver("[2620:fe::9]:53"), Quad9DNSv6[1])
}

func TestQuad9DNSv4_CanResolve(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resolver := newUDPResolver(string(Quad9DNSv4[0]), 5*time.Second, 1)
	records, err := resolver.ResolveType(ctx, "www.google.com", TypeA)
	assert.NoError(t, err)
	assert.NotEmpty(t, records)
}

func TestQuad9DNSv6_CanResolve(t *testing.T) {
	if os.Getenv("TEST_IPV6") == "" {
		t.Skip("Skipping IPv6 test (set TEST_IPV6=1 to enable)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resolver := newUDPResolver(string(Quad9DNSv6[0]), 5*time.Second, 1)
	records, err := resolver.ResolveType(ctx, "www.google.com", TypeAAAA)
	assert.NoError(t, err)
	assert.NotEmpty(t, records)
}

func TestOpenDNSv4_ValidAddresses(t *testing.T) {
	assert.Len(t, OpenDNSv4, 2)
	assert.Equal(t, Resolver("208.67.222.222:53"), OpenDNSv4[0])
	assert.Equal(t, Resolver("208.67.220.220:53"), OpenDNSv4[1])
}

func TestOpenDNSv6_ValidAddresses(t *testing.T) {
	assert.Len(t, OpenDNSv6, 2)
	assert.Equal(t, Resolver("[2620:119:35::35]:53"), OpenDNSv6[0])
	assert.Equal(t, Resolver("[2620:119:53::53]:53"), OpenDNSv6[1])
}

func TestOpenDNSv4_CanResolve(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resolver := newUDPResolver(string(OpenDNSv4[0]), 5*time.Second, 1)
	records, err := resolver.ResolveType(ctx, "www.google.com", TypeA)
	assert.NoError(t, err)
	assert.NotEmpty(t, records)
}

func TestOpenDNSv6_CanResolve(t *testing.T) {
	if os.Getenv("TEST_IPV6") == "" {
		t.Skip("Skipping IPv6 test (set TEST_IPV6=1 to enable)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resolver := newUDPResolver(string(OpenDNSv6[0]), 5*time.Second, 1)
	records, err := resolver.ResolveType(ctx, "www.google.com", TypeAAAA)
	assert.NoError(t, err)
	assert.NotEmpty(t, records)
}

func TestLevel3DNSv4_ValidAddresses(t *testing.T) {
	assert.Len(t, Level3DNSv4, 2)
	assert.Equal(t, Resolver("4.2.2.1:53"), Level3DNSv4[0])
	assert.Equal(t, Resolver("4.2.2.2:53"), Level3DNSv4[1])
}

func TestLevel3DNSv4_CanResolve(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resolver := newUDPResolver(string(Level3DNSv4[0]), 5*time.Second, 1)
	records, err := resolver.ResolveType(ctx, "www.google.com", TypeA)
	assert.NoError(t, err)
	assert.NotEmpty(t, records)
}

func TestComodoSecureDNSv4_ValidAddresses(t *testing.T) {
	assert.Len(t, ComodoSecureDNSv4, 2)
	assert.Equal(t, Resolver("8.26.56.26:53"), ComodoSecureDNSv4[0])
	assert.Equal(t, Resolver("8.20.247.20:53"), ComodoSecureDNSv4[1])
}

func TestComodoSecureDNSv4_CanResolve(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resolver := newUDPResolver(string(ComodoSecureDNSv4[0]), 5*time.Second, 1)
	records, err := resolver.ResolveType(ctx, "www.google.com", TypeA)
	assert.NoError(t, err)
	assert.NotEmpty(t, records)
}

func TestVerisignDNSv4_ValidAddresses(t *testing.T) {
	assert.Len(t, VerisignDNSv4, 2)
	assert.Equal(t, Resolver("64.6.64.6:53"), VerisignDNSv4[0])
	assert.Equal(t, Resolver("64.6.65.6:53"), VerisignDNSv4[1])
}

func TestVerisignDNSv4_CanResolve(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resolver := newUDPResolver(string(VerisignDNSv4[0]), 5*time.Second, 1)
	records, err := resolver.ResolveType(ctx, "www.google.com", TypeA)
	assert.NoError(t, err)
	assert.NotEmpty(t, records)
}

func TestDynOracleDNSv4_ValidAddresses(t *testing.T) {
	assert.Len(t, DynOracleDNSv4, 2)
	assert.Equal(t, Resolver("216.146.35.35:53"), DynOracleDNSv4[0])
	assert.Equal(t, Resolver("216.146.36.36:53"), DynOracleDNSv4[1])
}

func TestDynOracleDNSv4_CanResolve(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resolver := newUDPResolver(string(DynOracleDNSv4[0]), 5*time.Second, 1)
	records, err := resolver.ResolveType(ctx, "www.google.com", TypeA)
	assert.NoError(t, err)
	assert.NotEmpty(t, records)
}

func TestAliDNSv4_ValidAddresses(t *testing.T) {
	assert.Len(t, AliDNSv4, 2)
	assert.Equal(t, Resolver("223.5.5.5:53"), AliDNSv4[0])
	assert.Equal(t, Resolver("223.6.6.6:53"), AliDNSv4[1])
}

func TestAliDNSv4_CanResolve(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resolver := newUDPResolver(string(AliDNSv4[0]), 5*time.Second, 1)
	records, err := resolver.ResolveType(ctx, "www.google.com", TypeA)
	assert.NoError(t, err)
	assert.NotEmpty(t, records)
}

func TestNTTDNSv4_ValidAddresses(t *testing.T) {
	assert.Len(t, NTTDNSv4, 2)
	assert.Equal(t, Resolver("129.250.35.250:53"), NTTDNSv4[0])
	assert.Equal(t, Resolver("129.250.35.251:53"), NTTDNSv4[1])
}

func TestNTTDNSv6_ValidAddresses(t *testing.T) {
	assert.Len(t, NTTDNSv6, 2)
	assert.Equal(t, Resolver("[2001:418:3ff::53]:53"), NTTDNSv6[0])
	assert.Equal(t, Resolver("[2001:418:3ff::1:53]:53"), NTTDNSv6[1])
}

func TestNTTDNSv4_CanResolve(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resolver := newUDPResolver(string(NTTDNSv4[0]), 5*time.Second, 1)
	records, err := resolver.ResolveType(ctx, "www.google.com", TypeA)
	assert.NoError(t, err)
	assert.NotEmpty(t, records)
}

func TestNTTDNSv6_CanResolve(t *testing.T) {
	if os.Getenv("TEST_IPV6") == "" {
		t.Skip("Skipping IPv6 test (set TEST_IPV6=1 to enable)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resolver := newUDPResolver(string(NTTDNSv6[0]), 5*time.Second, 1)
	records, err := resolver.ResolveType(ctx, "www.google.com", TypeAAAA)
	assert.NoError(t, err)
	assert.NotEmpty(t, records)
}

func TestCleanBrowsingDNSv4_ValidAddresses(t *testing.T) {
	assert.Len(t, CleanBrowsingDNSv4, 2)
	assert.Equal(t, Resolver("185.228.168.10:53"), CleanBrowsingDNSv4[0])
	assert.Equal(t, Resolver("185.228.169.11:53"), CleanBrowsingDNSv4[1])
}

func TestCleanBrowsingDNSv6_ValidAddresses(t *testing.T) {
	assert.Len(t, CleanBrowsingDNSv6, 2)
	assert.Equal(t, Resolver("[2a0d:2a00:1::1]:53"), CleanBrowsingDNSv6[0])
	assert.Equal(t, Resolver("[2a0d:2a00:2::1]:53"), CleanBrowsingDNSv6[1])
}

func TestCleanBrowsingDNSv4_CanResolve(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resolver := newUDPResolver(string(CleanBrowsingDNSv4[0]), 5*time.Second, 1)
	records, err := resolver.ResolveType(ctx, "www.google.com", TypeA)
	assert.NoError(t, err)
	assert.NotEmpty(t, records)
}

func TestCleanBrowsingDNSv6_CanResolve(t *testing.T) {
	if os.Getenv("TEST_IPV6") == "" {
		t.Skip("Skipping IPv6 test (set TEST_IPV6=1 to enable)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resolver := newUDPResolver(string(CleanBrowsingDNSv6[0]), 5*time.Second, 1)
	records, err := resolver.ResolveType(ctx, "www.google.com", TypeAAAA)
	assert.NoError(t, err)
	assert.NotEmpty(t, records)
}

// TestAllResolvers_HavePort ensures all resolver addresses include :53 port.
func TestAllResolvers_HavePort(t *testing.T) {
	allResolvers := [][]Resolver{
		GooglePublicDNSv4, GooglePublicDNSv6,
		CloudflareDNSv4, CloudflareDNSv6,
		Quad9DNSv4, Quad9DNSv6,
		OpenDNSv4, OpenDNSv6,
		Level3DNSv4,
		ComodoSecureDNSv4,
		VerisignDNSv4,
		DynOracleDNSv4,
		AliDNSv4,
		NTTDNSv4, NTTDNSv6,
		CleanBrowsingDNSv4, CleanBrowsingDNSv6,
	}

	for _, resolvers := range allResolvers {
		for _, resolver := range resolvers {
			assert.True(t, strings.HasSuffix(string(resolver), ":53"),
				"resolver %s should have :53 port", resolver)
		}
	}
}

// TestAllResolvers_ParseableAddresses ensures all resolver addresses can be
// parsed by net.SplitHostPort.
func TestAllResolvers_ParseableAddresses(t *testing.T) {
	allResolvers := [][]Resolver{
		GooglePublicDNSv4, GooglePublicDNSv6,
		CloudflareDNSv4, CloudflareDNSv6,
		Quad9DNSv4, Quad9DNSv6,
		OpenDNSv4, OpenDNSv6,
		Level3DNSv4,
		ComodoSecureDNSv4,
		VerisignDNSv4,
		DynOracleDNSv4,
		AliDNSv4,
		NTTDNSv4, NTTDNSv6,
		CleanBrowsingDNSv4, CleanBrowsingDNSv6,
	}

	for _, resolvers := range allResolvers {
		for _, resolver := range resolvers {
			host, port, err := net.SplitHostPort(string(resolver))
			assert.NoError(t, err, "resolver %s should be parseable", resolver)
			assert.NotEmpty(t, host, "resolver %s should have host", resolver)
			assert.Equal(t, "53", port, "resolver %s should have port 53", resolver)
		}
	}
}

// TestIPv4Resolvers_ValidIPAddresses ensures all IPv4 resolver addresses are
// valid IP addresses (not hostnames).
func TestIPv4Resolvers_ValidIPAddresses(t *testing.T) {
	ipv4Resolvers := [][]Resolver{
		GooglePublicDNSv4,
		CloudflareDNSv4,
		Quad9DNSv4,
		OpenDNSv4,
		Level3DNSv4,
		ComodoSecureDNSv4,
		VerisignDNSv4,
		CleanBrowsingDNSv4,
	}

	for _, resolvers := range ipv4Resolvers {
		for _, resolver := range resolvers {
			host, _, err := net.SplitHostPort(string(resolver))
			assert.NoError(t, err)

			ip := net.ParseIP(host)
			assert.NotNil(t, ip, "resolver %s should be valid IP", resolver)
			assert.NotNil(t, ip.To4(), "resolver %s should be IPv4", resolver)
		}
	}
}

// TestIPv6Resolvers_ValidIPAddresses ensures all IPv6 resolver addresses are
// valid IPv6 addresses (not hostnames) and properly bracketed.
func TestIPv6Resolvers_ValidIPAddresses(t *testing.T) {
	ipv6Resolvers := [][]Resolver{
		GooglePublicDNSv6,
		CloudflareDNSv6,
		Quad9DNSv6,
		OpenDNSv6,
		NTTDNSv6,
		CleanBrowsingDNSv6,
	}

	for _, resolvers := range ipv6Resolvers {
		for _, resolver := range resolvers {
			// IPv6 addresses must be bracketed in host:port format
			assert.True(t, strings.HasPrefix(string(resolver), "["),
				"resolver %s should start with [", resolver)

			host, _, err := net.SplitHostPort(string(resolver))
			assert.NoError(t, err)

			ip := net.ParseIP(host)
			assert.NotNil(t, ip, "resolver %s should be valid IP", resolver)
			assert.Nil(t, ip.To4(), "resolver %s should be IPv6", resolver)
			assert.NotNil(t, ip.To16(), "resolver %s should be IPv6", resolver)
		}
	}
}
