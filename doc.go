// Copyright 2025 Bruno Schaatsbergen. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package dnsdialer provides a drop-in replacement for net.Dialer with configurable
// DNS resolution strategies.
//
// Instead of using your system's DNS resolver, dnsdialer queries multiple DNS servers
// using your chosen strategy (Race, Fallback, Consensus, or Compare).
//
// # Usage
//
// Replace net.Dialer.DialContext in any HTTP client, gRPC connection, or custom dialer:
//
//	resolver := dnsdialer.New(
//	    dnsdialer.WithResolvers("8.8.8.8:53", "1.1.1.1:53"),
//	    dnsdialer.WithStrategy(dnsdialer.Race{}),
//	)
//
//	client := &http.Client{
//	    Transport: &http.Transport{
//	        DialContext: resolver.DialContext,
//	    },
//	}
package dnsdialer
