package dnsdialer

import (
	"bufio"
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

// getSystemResolver attempts to read the system's DNS resolver from /etc/resolv.conf.
// Returns the first nameserver found, with :53 appended if no port is specified.
// Falls back to "8.8.8.8:53" if no resolver can be determined.
func getSystemResolver() string {
	file, err := os.Open("/etc/resolv.conf")
	if err != nil {
		return "8.8.8.8:53"
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		if strings.HasPrefix(line, "nameserver") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				nameserver := fields[1]
				if !strings.Contains(nameserver, ":") {
					nameserver = net.JoinHostPort(nameserver, "53")
				}
				return nameserver
			}
		}
	}

	return "8.8.8.8:53"
}

func BenchmarkStdLibResolver_LookupHost(b *testing.B) {
	ctx := context.Background()
	resolver := &net.Resolver{}

	b.ResetTimer()
	for b.Loop() {
		_, err := resolver.LookupHost(ctx, "www.google.com")
		if err != nil {
			b.Fatalf("lookup failed: %v", err)
		}
	}
}

func BenchmarkDNSDialer_LookupHost_NoCache_Single_Race(b *testing.B) {
	ctx := context.Background()
	dialer := New(
		WithResolvers(getSystemResolver()),
		WithStrategy(Race{}),
	)

	b.ResetTimer()
	for b.Loop() {
		_, err := dialer.lookupIPs(ctx, "www.google.com")
		if err != nil {
			b.Fatalf("lookup failed: %v", err)
		}
	}
}

func BenchmarkDNSDialer_LookupHost_NoCache_Race3(b *testing.B) {
	ctx := context.Background()
	dialer := New(
		WithResolvers("8.8.8.8:53", "1.1.1.1:53", "9.9.9.9:53"),
		WithStrategy(Race{}),
	)

	b.ResetTimer()
	for b.Loop() {
		_, err := dialer.lookupIPs(ctx, "www.google.com")
		if err != nil {
			b.Fatalf("lookup failed: %v", err)
		}
	}
}

func BenchmarkDNSDialer_LookupHost_NoCache_Fallback2(b *testing.B) {
	ctx := context.Background()
	dialer := New(
		WithResolvers("8.8.8.8:53", "1.1.1.1:53"),
		WithStrategy(Fallback{}),
	)

	b.ResetTimer()
	for b.Loop() {
		_, err := dialer.lookupIPs(ctx, "www.google.com")
		if err != nil {
			b.Fatalf("lookup failed: %v", err)
		}
	}
}

func BenchmarkDNSDialer_LookupHost_NoCache_Consensus2(b *testing.B) {
	ctx := context.Background()
	dialer := New(
		WithResolvers("8.8.8.8:53", "1.1.1.1:53"),
		WithStrategy(Consensus{
			MinAgreement: 2,
			IgnoreTTL:    true,
		}),
		WithTimeout(5*time.Second),
	)

	b.ResetTimer()
	for b.Loop() {
		_, err := dialer.lookupIPs(ctx, "www.google.com")
		if err != nil {
			b.Logf("lookup failed (consensus not reached): %v", err)
			continue
		}
	}
}

func BenchmarkDNSDialer_LookupHost_Cache_Single_Race(b *testing.B) {
	ctx := context.Background()
	dialer := New(
		WithResolvers(getSystemResolver()),
		WithStrategy(Race{}),
		WithCache(1000, 1*time.Second, 5*time.Minute),
	)

	b.ResetTimer()
	for b.Loop() {
		_, err := dialer.lookupIPs(ctx, "www.google.com")
		if err != nil {
			b.Fatalf("lookup failed: %v", err)
		}
	}
}

func BenchmarkDNSDialer_LookupHost_Cache_Race3(b *testing.B) {
	ctx := context.Background()
	dialer := New(
		WithResolvers("8.8.8.8:53", "1.1.1.1:53", "9.9.9.9:53"),
		WithStrategy(Race{}),
		WithCache(1000, 1*time.Second, 5*time.Minute),
	)

	b.ResetTimer()
	for b.Loop() {
		_, err := dialer.lookupIPs(ctx, "www.google.com")
		if err != nil {
			b.Fatalf("lookup failed: %v", err)
		}
	}
}

func BenchmarkDNSDialer_LookupHost_Cache_Fallback2(b *testing.B) {
	ctx := context.Background()
	dialer := New(
		WithResolvers("8.8.8.8:53", "1.1.1.1:53"),
		WithStrategy(Fallback{}),
		WithCache(1000, 1*time.Second, 5*time.Minute),
	)

	b.ResetTimer()
	for b.Loop() {
		_, err := dialer.lookupIPs(ctx, "www.google.com")
		if err != nil {
			b.Fatalf("lookup failed: %v", err)
		}
	}
}

func BenchmarkDNSDialer_LookupHost_Cache_Consensus2(b *testing.B) {
	ctx := context.Background()
	dialer := New(
		WithResolvers("8.8.8.8:53", "1.1.1.1:53"),
		WithStrategy(Consensus{
			MinAgreement: 2,
			IgnoreTTL:    true,
		}),
		WithCache(1000, 1*time.Second, 5*time.Minute),
	)

	b.ResetTimer()
	for b.Loop() {
		_, err := dialer.lookupIPs(ctx, "www.google.com")
		if err != nil {
			b.Fatalf("lookup failed: %v", err)
		}
	}
}

func BenchmarkStdLib_DialContext(b *testing.B) {
	ctx := context.Background()
	dialer := &net.Dialer{
		Timeout: 5 * time.Second,
	}

	b.ResetTimer()
	for b.Loop() {
		conn, err := dialer.DialContext(ctx, "tcp", "www.google.com:80")
		if err != nil {
			b.Fatalf("dial failed: %v", err)
		}
		conn.Close()
	}
}

func BenchmarkDNSDialer_DialContext_NoCache_Single_Race(b *testing.B) {
	ctx := context.Background()
	dialer := New(
		WithResolvers(getSystemResolver()),
		WithStrategy(Race{}),
	)

	b.ResetTimer()
	for b.Loop() {
		conn, err := dialer.DialContext(ctx, "tcp", "www.google.com:80")
		if err != nil {
			b.Fatalf("dial failed: %v", err)
		}
		conn.Close()
	}
}

func BenchmarkDNSDialer_DialContext_Cache_Single_Race(b *testing.B) {
	ctx := context.Background()
	dialer := New(
		WithResolvers(getSystemResolver()),
		WithStrategy(Race{}),
		WithCache(1000, 1*time.Second, 5*time.Minute),
	)

	b.ResetTimer()
	for b.Loop() {
		conn, err := dialer.DialContext(ctx, "tcp", "www.google.com:80")
		if err != nil {
			b.Fatalf("dial failed: %v", err)
		}
		conn.Close()
	}
}

func BenchmarkDNSDialer_DialContext_NoCache_Race3(b *testing.B) {
	ctx := context.Background()
	dialer := New(
		WithResolvers("8.8.8.8:53", "1.1.1.1:53", "9.9.9.9:53"),
		WithStrategy(Race{}),
	)

	for b.Loop() {
		conn, err := dialer.DialContext(ctx, "tcp", "www.google.com:80")
		if err != nil {
			b.Fatalf("dial failed: %v", err)
		}
		conn.Close()
	}
}

func BenchmarkDNSDialer_DialContext_Cache_Race3(b *testing.B) {
	ctx := context.Background()
	dialer := New(
		WithResolvers("8.8.8.8:53", "1.1.1.1:53", "9.9.9.9:53"),
		WithStrategy(Race{}),
		WithCache(1000, 1*time.Second, 5*time.Minute),
	)

	for b.Loop() {
		conn, err := dialer.DialContext(ctx, "tcp", "www.google.com:80")
		if err != nil {
			b.Fatalf("dial failed: %v", err)
		}
		conn.Close()
	}
}

func BenchmarkStdLib_HTTPGet(b *testing.B) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	for b.Loop() {
		resp, err := client.Get("http://www.google.com")
		if err != nil {
			b.Fatalf("HTTP GET failed: %v", err)
		}
		resp.Body.Close()
	}
}

func BenchmarkDNSDialer_HTTPGet_NoCache_Single_Race(b *testing.B) {
	dialer := New(
		WithResolvers(getSystemResolver()),
		WithStrategy(Race{}),
	)

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: dialer.DialContext,
		},
		Timeout: 10 * time.Second,
	}

	for b.Loop() {
		resp, err := client.Get("http://www.google.com")
		if err != nil {
			b.Fatalf("HTTP GET failed: %v", err)
		}
		resp.Body.Close()
	}
}

func BenchmarkDNSDialer_HTTPGet_Cache_Single_Race(b *testing.B) {
	dialer := New(
		WithResolvers(getSystemResolver()),
		WithStrategy(Race{}),
		WithCache(1000, 1*time.Second, 5*time.Minute),
	)

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: dialer.DialContext,
		},
		Timeout: 10 * time.Second,
	}

	for b.Loop() {
		resp, err := client.Get("http://www.google.com")
		if err != nil {
			b.Fatalf("HTTP GET failed: %v", err)
		}
		resp.Body.Close()
	}
}

func BenchmarkDNSDialer_HTTPGet_NoCache_Race3(b *testing.B) {
	dialer := New(
		WithResolvers("8.8.8.8:53", "1.1.1.1:53", "9.9.9.9:53"),
		WithStrategy(Race{}),
	)

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: dialer.DialContext,
		},
		Timeout: 10 * time.Second,
	}

	for b.Loop() {
		resp, err := client.Get("http://www.google.com")
		if err != nil {
			b.Fatalf("HTTP GET failed: %v", err)
		}
		resp.Body.Close()
	}
}

func BenchmarkDNSDialer_HTTPGet_Cache_Race3(b *testing.B) {
	dialer := New(
		WithResolvers("8.8.8.8:53", "1.1.1.1:53", "9.9.9.9:53"),
		WithStrategy(Race{}),
		WithCache(1000, 1*time.Second, 5*time.Minute),
	)

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: dialer.DialContext,
		},
		Timeout: 10 * time.Second,
	}

	for b.Loop() {
		resp, err := client.Get("http://www.google.com")
		if err != nil {
			b.Fatalf("HTTP GET failed: %v", err)
		}
		resp.Body.Close()
	}
}

func BenchmarkStdLib_HTTPGet_Local(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	b.ResetTimer()
	for b.Loop() {
		resp, err := client.Get(server.URL)
		if err != nil {
			b.Fatalf("HTTP GET failed: %v", err)
		}
		resp.Body.Close()
	}
}

func BenchmarkDNSDialer_HTTPGet_Local(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	dialer := New(
		WithResolvers("8.8.8.8:53", "1.1.1.1:53"),
		WithStrategy(Race{}),
		WithCache(1000, 1*time.Second, 5*time.Minute),
	)

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: dialer.DialContext,
		},
		Timeout: 5 * time.Second,
	}

	b.ResetTimer()
	for b.Loop() {
		resp, err := client.Get(server.URL)
		if err != nil {
			b.Fatalf("HTTP GET failed: %v", err)
		}
		resp.Body.Close()
	}
}

func BenchmarkStdLib_DialContext_IPLiteral(b *testing.B) {
	ctx := context.Background()
	dialer := &net.Dialer{
		Timeout: 5 * time.Second,
	}

	b.ResetTimer()
	for b.Loop() {
		conn, err := dialer.DialContext(ctx, "tcp", "8.8.8.8:53")
		if err != nil {
			b.Fatalf("dial failed: %v", err)
		}
		conn.Close()
	}
}

func BenchmarkDNSDialer_DialContext_IPLiteral(b *testing.B) {
	ctx := context.Background()
	dialer := New(
		WithResolvers("8.8.8.8:53", "1.1.1.1:53"),
		WithStrategy(Race{}),
		WithCache(1000, 1*time.Second, 5*time.Minute),
	)

	b.ResetTimer()
	for b.Loop() {
		conn, err := dialer.DialContext(ctx, "tcp", "8.8.8.8:53")
		if err != nil {
			b.Fatalf("dial failed: %v", err)
		}
		conn.Close()
	}
}
