# dnsdialer

[![Go Reference](https://pkg.go.dev/badge/github.com/bschaatsbergen/dnsdialer.svg)](https://pkg.go.dev/github.com/bschaatsbergen/dnsdialer)
[![Go Report Card](https://goreportcard.com/badge/github.com/bschaatsbergen/dnsdialer)](https://goreportcard.com/report/bschaatsbergen/dnsdialer)

A `net.Dialer.DialContext` replacement with deterministic DNS resolution strategies.

Implement concurrent races for minimum latency, consensus validation for poisoning detection, or ordered fallback for resolver diversity. Built on [miekg/dns](https://pkg.go.dev/github.com/miekg/dns) with sub-resolver control over timeout behavior, retry logic, and response validation.

Useful for systems where DNS resolution latency impacts P99 response times and DNS failures cascade into service outages. Drop into any `DialContext` call in HTTP transports, gRPC clients, or custom connection pools.

## How it works

dnsdialer implements the same `DialContext` signature as `net.Dialer`, making it a drop-in replacement for any Go code that accepts a custom dialer (HTTP clients, gRPC, etc.).

The only difference: instead of using your system DNS resolver, it queries multiple DNS servers using your chosen strategy.

## Usage

### HTTP Client

```go
dialer := dnsdialer.New(
    dnsdialer.WithResolvers("8.8.8.8:53", "1.1.1.1:53"),
    dnsdialer.WithStrategy(dnsdialer.Race{}),
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
    dnsdialer.WithStrategy(dnsdialer.Consensus{
        MinAgreement: 2,
        IgnoreTTL:    true,
    }),
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
