package dnsdialer

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/miekg/dns"
)

// udpResolver implements the resolver interface using UDP transport.
//
// It uses connection pooling to reduce socket allocation overhead and supports
// context-based deadlines for timeout control.
type udpResolver struct {
	// addr is the DNS server address with port (e.g., "8.8.8.8:53")
	addr string

	// timeout is the default timeout we use if the context has no deadline set
	timeout time.Duration

	// client is currently unused, kept around for potential future use with miekg/dns Client API
	client *dns.Client

	// connPool is the connection pool for socket reuse, important for performance
	connPool *connPool
}

func newUDPResolver(addr string, timeout time.Duration, poolSize int) *udpResolver {
	// Ensure the address includes a port. DNS servers typically listen on port 53.
	// This lets users specify just "8.8.8.8" instead of requiring "8.8.8.8:53".
	if _, _, err := net.SplitHostPort(addr); err != nil {
		addr = net.JoinHostPort(addr, "53")
	}

	return &udpResolver{
		addr:     addr,
		timeout:  timeout,
		connPool: newConnPool(addr, timeout, poolSize),
		client: &dns.Client{
			Net:     "udp",
			Timeout: timeout,
			UDPSize: 4096, // EDNS0 UDP payload size, bigger than default 512 to support larger responses
		},
	}
}

func (r *udpResolver) ResolveType(ctx context.Context, host string, qtype RecordType) ([]Record, error) {
	// Construct the DNS query message
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(host), uint16(qtype)) // dns.Fqdn ensures trailing dot (e.g., "example.com.")
	msg.RecursionDesired = true                    // Ask the server to recursively resolve if it doesn't have the answer cached

	// Get a connection from the pool. This might return a reused connection or create a new
	// one if the pool is empty.
	conn, err := r.connPool.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}

	// Set the deadline on the connection to honor context cancellation and timeouts. We prefer
	// the context deadline if set, otherwise use the resolver's default timeout.
	// We ignore errors from SetDeadline because:
	// 1. Failure is rare, would indicate connection already closed
	// 2. The subsequent Read/Write will fail anyway if there's a problem
	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
	} else {
		_ = conn.SetDeadline(time.Now().Add(r.timeout))
	}

	// Wrap the UDP connection in miekg/dns.Conn for DNS wire protocol handling
	dnsConn := &dns.Conn{Conn: conn}

	// Send the query. If this fails, the connection is likely broken, so we close it
	// rather than returning it to the pool where it might cause future failures.
	if err := dnsConn.WriteMsg(msg); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("query failed: %w", err)
	}

	// Read the response. Same error handling as WriteMsg: close on error instead of returning to pool.
	response, err := dnsConn.ReadMsg()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("query failed: %w", err)
	}

	// Query succeeded, so return the connection to the pool for reuse. Do this before processing
	// the response so the connection becomes available ASAP for other queries.
	r.connPool.Put(conn)

	// Check DNS response code. RcodeSuccess (0) means the query succeeded. Other codes include
	// NXDomain (domain doesn't exist), ServFail (server error), etc.
	if response.Rcode != dns.RcodeSuccess {
		return nil, fmt.Errorf("dns error: %s", dns.RcodeToString[response.Rcode])
	}

	// Parse the answer section into our Record format. The DNS response contains raw resource
	// records that we need to convert into a more usable structure.
	var records []Record
	for _, ans := range response.Answer {
		record := Record{
			Type: RecordType(ans.Header().Rrtype),
			TTL:  ans.Header().Ttl,
		}

		// Extract the value based on record type. Each DNS record type has its own struct
		// in miekg/dns, so we use a type switch to handle them.
		switch a := ans.(type) {
		case *dns.A:
			// IPv4 address (e.g., "93.184.216.34")
			record.Value = a.A.String()
		case *dns.AAAA:
			// IPv6 address (e.g., "2606:2800:220:1:248:1893:25c8:1946")
			record.Value = a.AAAA.String()
		case *dns.CNAME:
			// Canonical name / alias (e.g., "www.example.com.")
			record.Value = a.Target
		case *dns.MX:
			// Mail exchange, includes priority and mailserver
			// Format: "priority mailserver" (e.g., "10 mail.example.com.")
			record.Value = fmt.Sprintf("%d %s", a.Preference, a.Mx)
		case *dns.NS:
			// Name server (e.g., "ns1.example.com.")
			record.Value = a.Ns
		case *dns.TXT:
			// Text record, can contain multiple strings, we format as a single string
			record.Value = fmt.Sprintf("%v", a.Txt)
		case *dns.SOA:
			// Start of Authority, contains zone metadata
			// Format: "ns mbox serial refresh retry expire minttl"
			record.Value = fmt.Sprintf("%s %s %d %d %d %d %d",
				a.Ns, a.Mbox, a.Serial, a.Refresh, a.Retry, a.Expire, a.Minttl)
		case *dns.PTR:
			// Pointer record, used for reverse DNS lookups
			record.Value = a.Ptr
		case *dns.SRV:
			// Service record, used for service discovery
			// Format: "priority weight port target"
			record.Value = fmt.Sprintf("%d %d %d %s",
				a.Priority, a.Weight, a.Port, a.Target)
		default:
			// For record types we don't explicitly handle, use the library's string representation.
			// This provides basic support for any record type without requiring explicit handling
			// for each one.
			record.Value = ans.String()
		}

		records = append(records, record)
	}

	// Some DNS servers return RcodeSuccess with an empty answer section when a record
	// exists but has no data (e.g., a domain with no A records). We treat this as an
	// error rather than returning an empty slice, to distinguish it from never calling
	// this function vs calling it and getting nothing.
	if len(records) == 0 {
		return nil, fmt.Errorf("no records found")
	}

	return records, nil
}

func (r *udpResolver) Name() string {
	return r.addr
}
