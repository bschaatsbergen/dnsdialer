# dnsdialer

[![Go Reference](https://pkg.go.dev/badge/github.com/bschaatsbergen/dnsdialer.svg)](https://pkg.go.dev/github.com/bschaatsbergen/dnsdialer)
[![Go Report Card](https://goreportcard.com/badge/github.com/bschaatsbergen/dnsdialer)](https://goreportcard.com/report/bschaatsbergen/dnsdialer)

This package allows you to take control of DNS resolution behavior through configurable multi-resolver strategies.

Why you'd want multiple resolvers: Redundancy (primary resolver failure doesn't cascade into total DNS outage). Performance (concurrent queries across resolvers, returning fastest response). Security (consensus validation across independent resolvers mitigates poisoning and MITM attacks). Integrity (cross-resolver validation detects poisoning, cache corruption, and configuration drift before propagation).

Most OS-level DNS stacks already support multiple resolvers, but they don't use them in parallel, they typically try the first, then fail over in sequence (which can be slow if the first resolver hangs). In high-throughput systems where single-digit millisecond DNS latency affects tail latencies and resolver failures propagate into cascading outages, you need deterministic multi-resolver behavior.

While OS-level DNS caching (mDNSResponder on macOS, systemd-resolved on Linux) provides sub-millisecond lookups, this package bypasses it by default to ensure fresh results for redundancy and consensus validation. Optional LRU caching with TTL-aware expiration is available via `WithCache()` to reduce latency on repeated lookups while maintaining explicit control over cache size and TTL bounds.

This package provides a `DialContext` implementation that plugs directly into HTTP transports, gRPC clients, or any custom connection pools expecting [net.Dialer](https://pkg.go.dev/net#Dialer).

## How it works

Built on [miekg/dns](https://pkg.go.dev/github.com/miekg/dns), dnsdialer implements the same `DialContext` signature as [net.Dialer](https://pkg.go.dev/net#Dialer), making it a drop-in replacement for any Go code that accepts a custom dialer (HTTP clients, gRPC, etc.).

The only difference: instead of using your system DNS resolver, it queries multiple DNS servers using your chosen strategy.

## Performance

The standard library's [net.Dialer](https://pkg.go.dev/net#Dialer) relies on OS-level DNS caching (mDNSResponder on macOS, systemd-resolved on Linux), which provides sub-millisecond lookups once cached. dnsdialer has its own in-process LRU cache to avoid shared global state and maintain explicit control over TTL bounds. By caching parsed [net.IP](https://pkg.go.dev/net#IP) slices instead of raw DNS strings, you get similar dial latency with reduced per-lookup allocations.

```console
go test -bench='^BenchmarkStdLib_DialContext$|^BenchmarkDNSDialer_DialContext_Cache_Single_Race$' -run=^$ -benchtime=5s -benchmem
goos: darwin
goarch: arm64
pkg: github.com/bschaatsbergen/dnsdialer
cpu: Apple M4
BenchmarkStdLib_DialContext-10                               360          16735385 ns/op            3549 B/op         57 allocs/op
BenchmarkDNSDialer_DialContext_Cache_Single_Race-10          354          16519391 ns/op             936 B/op         21 allocs/op
PASS
ok      github.com/bschaatsbergen/dnsdialer     12.114s
```

The standard library's DNS resolver implementation varies by CGO status: with CGO enabled (default), it uses [getaddrinfo()](https://man7.org/linux/man-pages/man3/getaddrinfo.3.html) which requires a system call and inter-process communication to the OS DNS cache (mDNSResponder on macOS, systemd-resolved on Linux) for every lookup. With CGO disabled, it uses a [pure Go](https://github.com/golang/go/blob/master/src/net/lookup_unix.go#L58) DNS implementation that sends queries directly to DNS servers for every lookup. dnsdialer maintains deterministic in-memory caching regardless of build configuration, providing much faster lookups by eliminating external communication overhead (system calls, inter-process communication, or network round-trips):

```console
CGO_ENABLED=0 go test -bench=CGO -run=^ -benchtime=5s -benchmem
goos: darwin
goarch: arm64
pkg: github.com/bschaatsbergen/dnsdialer
cpu: Apple M4
BenchmarkCGO_StdLib_DialContext-10                   363          16641334 ns/op
BenchmarkCGO_DNSDialer_DialContext-10                363          16627420 ns/op
PASS
ok      github.com/bschaatsbergen/dnsdialer     12.343s
```

## Usage

### HTTP Client

```go
dialer := dnsdialer.New(
    dnsdialer.WithResolvers("8.8.8.8:53", "1.1.1.1:53"),
    dnsdialer.WithStrategy(dnsdialer.Race{}),
    dnsdialer.WithCache(1000, 1*time.Second, 5*time.Minute),
)

client := &http.Client{
    Transport: &http.Transport{
        DialContext: dialer.DialContext,
    },
}

resp, err := client.Get("https://api.github.com")
```

### gRPC

```go
dialer := dnsdialer.New(
    dnsdialer.WithResolvers("8.8.8.8:53", "1.1.1.1:53"),
    dnsdialer.WithStrategy(dnsdialer.Race{}),
    dnsdialer.WithCache(1000, 1*time.Second, 5*time.Minute),
)

conn, err := grpc.Dial(
    "api.example.com:443",
    grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
        return dialer.DialContext(ctx, "tcp", addr)
    }),
)
```

## Strategies

### Race

Queries all servers simultaneously and returns the first successful response.
Minimizes latency by leveraging the fastest available server.

```go
dialer := dnsdialer.New(
    dnsdialer.WithResolvers("8.8.8.8:53", "1.1.1.1:53"),
    dnsdialer.WithStrategy(dnsdialer.Race{}),
)
```

### Fallback

Queries servers sequentially in order, providing reliability through ordered failover.

```go
dialer := dnsdialer.New(
    dnsdialer.WithResolvers("primary.dns:53", "backup.dns:53"),
    dnsdialer.WithStrategy(dnsdialer.Fallback{}),
)
```

### Consensus

Requires a minimum number of servers to agree on the response.
Improves security by detecting inconsistencies or DNS poisoning.

```go
dialer := dnsdialer.New(
    dnsdialer.WithResolvers("8.8.8.8:53", "1.1.1.1:53", "9.9.9.9:53"),
    dnsdialer.WithStrategy(dnsdialer.Consensus{
        MinAgreement: 2,    // Require 2 servers to agree
        IgnoreTTL:    true, // Ignore TTL differences when comparing
    }),
)
```

### Compare

Queries all servers and detects discrepancies, calling a user-provided callback when differences are found.
Useful for monitoring DNS server integrity.

```go
dialer := dnsdialer.New(
    dnsdialer.WithResolvers("8.8.8.8:53", "1.1.1.1:53"),
    dnsdialer.WithStrategy(dnsdialer.Compare{
        OnDiscrepancy: func(host string, qtype dnsdialer.RecordType, results map[string][]dnsdialer.Record) {
            fmt.Printf("Discrepancy detected for %s (%s)\n", host, qtype)
        },
        IgnoreTTL: true,
    }),
)
```
