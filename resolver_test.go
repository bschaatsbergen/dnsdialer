package dnsdialer

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDialer_DialContext(t *testing.T) {
	// Create local HTTP test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello from test server"))
	}))
	defer server.Close()

	dialer := New(
		WithResolvers("8.8.8.8:53"),
		WithStrategy(Race{}),
	)

	// Parse server URL to get host:port
	serverURL, _ := url.Parse(server.URL)

	ctx := context.Background()
	conn, err := dialer.DialContext(ctx, "tcp", serverURL.Host)

	assert.NoError(t, err)
	assert.NotNil(t, conn)
	assert.Equal(t, "tcp", conn.LocalAddr().Network())
	conn.Close()
}

func TestDialer_DialContext_HTTPClient(t *testing.T) {
	// Create local HTTP test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello from test server"))
	}))
	defer server.Close()

	dialer := New(
		WithResolvers("8.8.8.8:53"),
		WithStrategy(Race{}),
	)

	// Create HTTP client using our dialer
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: dialer.DialContext,
		},
	}

	resp, err := client.Get(server.URL)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
}
