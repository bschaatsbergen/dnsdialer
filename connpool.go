// Copyright 2025 Bruno Schaatsbergen. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dnsdialer

import (
	"net"
	"sync"
	"time"
)

// connPool manages a pool of reusable UDP connections to a single DNS resolver.
//
// Connection pooling reduces the overhead of socket creation/destruction for
// high-throughput DNS resolution. Each pool maintains up to 'size' connections,
// creating them on demand and reusing them across queries.
//
// Concurrency: The pool is safe for concurrent access. Multiple goroutines can
// Get and Put connections simultaneously.
type connPool struct {
	// addr is the DNS resolver address we're pooling connections for (e.g., "8.8.8.8:53")
	addr string

	// timeout is the connection timeout for creating new connections
	timeout time.Duration

	// size is the maximum number of pooled connections we'll keep around
	size int

	// conns is a buffered channel acting as a LIFO queue of available connections
	conns chan *net.UDPConn

	// mu protects the 'closed' flag, we don't want races when closing
	mu sync.Mutex

	// closed is set to true when pool is closed, prevents new Gets from working
	closed bool

	// dialer is used for creating new connections, reuse it instead of allocating each time
	dialer *net.Dialer
}

func newConnPool(addr string, timeout time.Duration, size int) *connPool {
	if size <= 0 {
		size = 4 // default pool size, reasonable balance between connection reuse and resource usage
	}

	pool := &connPool{
		addr:    addr,
		timeout: timeout,
		size:    size,
		// Buffered channel of size 'size' acts as a queue. The channel buffer size is what
		// limits how many connections we'll keep idle. When the channel is full, Put() will
		// just close excess connections rather than blocking.
		conns: make(chan *net.UDPConn, size),
		dialer: &net.Dialer{
			Timeout: timeout,
		},
	}

	return pool
}

// Get retrieves a connection from the pool or creates a new one.
//
// This implements a lazy allocation strategy: connections are only created
// when needed, not pre-allocated. The pool will grow up to 'size' connections
// over time as they're Put() back.
func (p *connPool) Get() (*net.UDPConn, error) {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil, net.ErrClosed
	}
	p.mu.Unlock()

	// Try to get an idle connection from the pool. Using select with default makes this
	// a non-blocking receive: if a connection is available, grab it; otherwise fall through
	// to create a new one.
	select {
	case conn := <-p.conns:
		// Got a connection from the pool. In theory it should be valid, but we check
		// anyway in case something unexpected happened, shouldn't be nil in practice.
		if conn != nil {
			return conn, nil
		}
	default:
		// Pool is empty. This happens when:
		// 1. No connections have been created yet (cold start)
		// 2. All connections are currently in use
		// 3. Pool is under high load
		// Fall through to create a new connection.
	}

	// Create a new connection. Note that we don't enforce the pool size limit here,
	// we can temporarily have more than 'size' connections in flight. The limit is really
	// enforced by Put(), which will close connections when the pool is full.
	raddr, err := net.ResolveUDPAddr("udp", p.addr)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// Put returns a connection to the pool for reuse, or closes it if the pool is full.
//
// Always call Put() after you're done with a connection, even if an error occurred.
// This ensures proper resource cleanup and connection reuse.
func (p *connPool) Put(conn *net.UDPConn) {
	if conn == nil {
		return
	}

	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		// Pool is closed, so don't return the connection to it. Just close it immediately.
		// The blank identifier assignment silences linter warnings about unchecked errors,
		// we can't do anything meaningful with Close() errors here anyway.
		_ = conn.Close()
		return
	}
	p.mu.Unlock()

	// Try to return the connection to the pool. If successful, the connection becomes
	// available for the next Get() call.
	select {
	case p.conns <- conn:
		// Successfully queued the connection for reuse. The connection stays open
		// and will be returned by a future Get() call.
	default:
		// Pool is full. This happens when more than 'size' connections were created
		// during high load and are now being returned. Rather than blocking or growing
		// the pool unbounded, we just close excess connections.
		//
		// This is a key part of the pool's self-regulation: it can temporarily exceed
		// its size limit during load spikes, but will shrink back down as connections
		// get returned.
		_ = conn.Close()
	}
}

// Close shuts down the pool and closes all idle connections.
//
// After Close() is called, Get() will return net.ErrClosed and Put() will
// close connections immediately rather than pooling them.
//
// Note: This only closes idle connections currently in the pool. Connections
// that are checked out (via Get() but not yet Put() back) will not be closed.
// The caller is responsible for ensuring all in-flight connections are returned
// or closed before calling Close().
func (p *connPool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}

	p.closed = true
	// Closing the channel signals that no more connections will be added. This also
	// allows the range loop below to terminate once all queued connections have been
	// processed.
	close(p.conns)

	// Drain and close all idle connections in the pool. The range terminates when
	// the channel is both closed and empty.
	for conn := range p.conns {
		if conn != nil {
			_ = conn.Close()
		}
	}

	return nil
}
