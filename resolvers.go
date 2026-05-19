// Copyright 2025 Bruno Schaatsbergen. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dnsdialer

// Resolver represents a DNS resolver address in host:port format.
type Resolver string

// Predefined public DNS resolvers.
//
// Most providers give multiple addresses for redundancy and IPv6 support.
// These slices can be passed directly to dnsdialer strategies like Race, Fallback, or Consensus.
//
// Example of combining multiple IPv4 resolvers inline:
//
//	  dialer := New(
//			WithResolvers(string(GooglePublicDNSv4[0]), string(CloudflareDNSv4[1])),
//	     WithStrategy(Race{}),
//	  )
var (
	// Google Public DNS
	GooglePublicDNSv4 = []Resolver{
		"8.8.8.8:53",
		"8.8.4.4:53",
	}
	GooglePublicDNSv6 = []Resolver{
		"[2001:4860:4860::8888]:53",
		"[2001:4860:4860::8844]:53",
	}

	// Cloudflare Public DNS
	CloudflareDNSv4 = []Resolver{
		"1.1.1.1:53",
		"1.0.0.1:53",
	}
	CloudflareDNSv6 = []Resolver{
		"[2606:4700:4700::1111]:53",
		"[2606:4700:4700::1001]:53",
	}

	// Quad9 Public DNS
	Quad9DNSv4 = []Resolver{
		"9.9.9.9:53",
		"149.112.112.112:53",
	}
	Quad9DNSv6 = []Resolver{
		"[2620:fe::fe]:53",
		"[2620:fe::9]:53",
	}

	// OpenDNS (Cisco)
	OpenDNSv4 = []Resolver{
		"208.67.222.222:53",
		"208.67.220.220:53",
	}
	OpenDNSv6 = []Resolver{
		"[2620:119:35::35]:53",
		"[2620:119:53::53]:53",
	}

	// Level 3 / CenturyLink
	Level3DNSv4 = []Resolver{
		"4.2.2.1:53",
		"4.2.2.2:53",
	}

	// Comodo Secure DNS
	ComodoSecureDNSv4 = []Resolver{
		"8.26.56.26:53",
		"8.20.247.20:53",
	}

	// Verisign Public DNS
	VerisignDNSv4 = []Resolver{
		"64.6.64.6:53",
		"64.6.65.6:53",
	}

	// Dyn / Oracle Public DNS
	DynOracleDNSv4 = []Resolver{
		"216.146.35.35:53",
		"216.146.36.36:53",
	}

	// Alibaba Public DNS
	AliDNSv4 = []Resolver{
		"223.5.5.5:53",
		"223.6.6.6:53",
	}

	// NTT Public DNS
	NTTDNSv4 = []Resolver{
		"129.250.35.250:53",
		"129.250.35.251:53",
	}
	NTTDNSv6 = []Resolver{
		"[2001:418:3ff::53]:53",
		"[2001:418:3ff::1:53]:53",
	}

	// CleanBrowsing Family-safe DNS
	CleanBrowsingDNSv4 = []Resolver{
		"185.228.168.10:53",
		"185.228.169.11:53",
	}
	CleanBrowsingDNSv6 = []Resolver{
		"[2a0d:2a00:1::1]:53",
		"[2a0d:2a00:2::1]:53",
	}
)
