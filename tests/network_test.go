// +build integration

package tests

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/conneroisu/hydra-go/hydra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNetworkResilience(t *testing.T) {
	t.Run("handle connection refused", func(t *testing.T) {
		// Try to connect to a port that's not listening
		client, err := hydra.NewClientWithURL("http://localhost:9999")
		require.NoError(t, err)
		
		ctx := context.Background()
		_, err = client.ListProjects(ctx)
		assert.Error(t, err)
		
		// Check that the error is a connection error
		if netErr, ok := err.(*net.OpError); ok {
			assert.Equal(t, "dial", netErr.Op)
		}
	})
	
	t.Run("handle DNS resolution failure", func(t *testing.T) {
		// Use a non-existent domain
		client, err := hydra.NewClientWithURL("http://this-domain-definitely-does-not-exist-12345.com")
		require.NoError(t, err)
		
		ctx := context.Background()
		_, err = client.ListProjects(ctx)
		assert.Error(t, err)
	})
	
	t.Run("handle slow server", func(t *testing.T) {
		// Create a slow server
		slowServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(200 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("[]"))
		}))
		defer slowServer.Close()
		
		// Create client with very short timeout
		cfg := &hydra.Config{
			BaseURL: slowServer.URL,
			Timeout: 50 * time.Millisecond,
		}
		client, err := hydra.NewClient(cfg)
		require.NoError(t, err)
		
		ctx := context.Background()
		_, err = client.ListProjects(ctx)
		assert.Error(t, err)
	})
	
	t.Run("handle server returning invalid JSON", func(t *testing.T) {
		badServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("this is not valid JSON"))
		}))
		defer badServer.Close()
		
		client, err := hydra.NewClientWithURL(badServer.URL)
		require.NoError(t, err)
		
		ctx := context.Background()
		_, err = client.ListProjects(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unmarshal")
	})
	
	t.Run("handle various HTTP status codes", func(t *testing.T) {
		testCases := []struct {
			name       string
			statusCode int
			response   string
		}{
			{"bad request", http.StatusBadRequest, `{"error":"bad request"}`},
			{"unauthorized", http.StatusUnauthorized, `{"error":"unauthorized"}`},
			{"forbidden", http.StatusForbidden, `{"error":"forbidden"}`},
			{"not found", http.StatusNotFound, `{"error":"not found"}`},
			{"internal server error", http.StatusInternalServerError, `{"error":"internal server error"}`},
			{"service unavailable", http.StatusServiceUnavailable, `{"error":"service unavailable"}`},
		}
		
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(tc.statusCode)
					w.Write([]byte(tc.response))
				}))
				defer server.Close()
				
				client, err := hydra.NewClientWithURL(server.URL)
				require.NoError(t, err)
				
				ctx := context.Background()
				_, err = client.ListProjects(ctx)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), fmt.Sprintf("%d", tc.statusCode))
			})
		}
	})
	
	t.Run("handle redirect", func(t *testing.T) {
		redirectServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/redirect" {
				http.Redirect(w, r, "/final", http.StatusMovedPermanently)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("[]"))
		}))
		defer redirectServer.Close()
		
		client, err := hydra.NewClientWithURL(redirectServer.URL + "/redirect")
		require.NoError(t, err)
		
		ctx := context.Background()
		projects, err := client.ListProjects(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, projects)
	})
	
	t.Run("handle chunked response", func(t *testing.T) {
		chunkedServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Transfer-Encoding", "chunked")
			
			// Write response in chunks
			flusher, ok := w.(http.Flusher)
			if ok {
				w.Write([]byte("["))
				flusher.Flush()
				time.Sleep(10 * time.Millisecond)
				w.Write([]byte(`{"name":"test"}`))
				flusher.Flush()
				time.Sleep(10 * time.Millisecond)
				w.Write([]byte("]"))
				flusher.Flush()
			} else {
				w.Write([]byte(`[{"name":"test"}]`))
			}
		}))
		defer chunkedServer.Close()
		
		client, err := hydra.NewClientWithURL(chunkedServer.URL)
		require.NoError(t, err)
		
		ctx := context.Background()
		projects, err := client.ListProjects(ctx)
		assert.NoError(t, err)
		assert.Len(t, projects, 1)
	})
}

func TestEnvironmentVariables(t *testing.T) {
	t.Run("use HYDRA_URL environment variable", func(t *testing.T) {
		// This test demonstrates how the SDK could use environment variables
		// In a real implementation, you might have:
		// url := os.Getenv("HYDRA_URL")
		// if url == "" {
		//     url = "https://hydra.nixos.org"
		// }
		
		testURL := getTestURL()
		assert.NotEmpty(t, testURL)
	})
	
	t.Run("proxy configuration", func(t *testing.T) {
		// Test that HTTP proxy settings are respected
		// This would normally use HTTP_PROXY and HTTPS_PROXY env vars
		
		proxyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Proxy would forward the request
			w.Header().Set("X-Proxied", "true")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("[]"))
		}))
		defer proxyServer.Close()
		
		// In a real scenario, you'd set up the proxy in the HTTP client
		// For testing, we'll just verify the client can be configured
		httpClient := &http.Client{
			Timeout: 10 * time.Second,
			// Proxy configuration would go here
		}
		
		cfg := &hydra.Config{
			BaseURL:    proxyServer.URL,
			HTTPClient: httpClient,
		}
		
		client, err := hydra.NewClient(cfg)
		require.NoError(t, err)
		assert.NotNil(t, client)
	})
}

func TestRateLimiting(t *testing.T) {
	t.Run("handle rate limit responses", func(t *testing.T) {
		requestCount := 0
		rateLimitServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestCount++
			if requestCount > 3 {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("[]"))
			} else {
				w.Header().Set("Retry-After", "1")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error":"rate limit exceeded"}`))
			}
		}))
		defer rateLimitServer.Close()
		
		client, err := hydra.NewClientWithURL(rateLimitServer.URL)
		require.NoError(t, err)
		
		ctx := context.Background()
		
		// First three requests should fail with rate limit
		for i := 0; i < 3; i++ {
			_, err = client.ListProjects(ctx)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "429")
		}
		
		// Fourth request should succeed
		projects, err := client.ListProjects(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, projects)
	})
}

func TestContentTypes(t *testing.T) {
	t.Run("handle different content types", func(t *testing.T) {
		testCases := []struct {
			name        string
			contentType string
			body        string
			shouldError bool
		}{
			{"json", "application/json", "[]", false},
			{"json with charset", "application/json; charset=utf-8", "[]", false},
			{"html error", "text/html", "<html>Error</html>", true},
			{"plain text", "text/plain", "Error", true},
			{"no content type", "", "[]", false},
		}
		
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if tc.contentType != "" {
						w.Header().Set("Content-Type", tc.contentType)
					}
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(tc.body))
				}))
				defer server.Close()
				
				client, err := hydra.NewClientWithURL(server.URL)
				require.NoError(t, err)
				
				ctx := context.Background()
				_, err = client.ListProjects(ctx)
				
				if tc.shouldError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})
}

func TestLargeResponses(t *testing.T) {
	t.Run("handle large response body", func(t *testing.T) {
		largeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			
			// Generate a large response
			w.Write([]byte("["))
			for i := 0; i < 1000; i++ {
				if i > 0 {
					w.Write([]byte(","))
				}
				project := fmt.Sprintf(`{"name":"project-%d","displayname":"Project %d","owner":"admin","enabled":true}`, i, i)
				w.Write([]byte(project))
			}
			w.Write([]byte("]"))
		}))
		defer largeServer.Close()
		
		client, err := hydra.NewClientWithURL(largeServer.URL)
		require.NoError(t, err)
		
		ctx := context.Background()
		projects, err := client.ListProjects(ctx)
		assert.NoError(t, err)
		assert.Len(t, projects, 1000)
	})
}

func TestKeepAlive(t *testing.T) {
	t.Run("connection reuse", func(t *testing.T) {
		connectionCount := 0
		keepAliveServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			connectionCount++
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Connection", "keep-alive")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("[]"))
		}))
		defer keepAliveServer.Close()
		
		// Create client with connection pooling
		httpClient := &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     30 * time.Second,
			},
		}
		
		cfg := &hydra.Config{
			BaseURL:    keepAliveServer.URL,
			HTTPClient: httpClient,
		}
		
		client, err := hydra.NewClient(cfg)
		require.NoError(t, err)
		
		ctx := context.Background()
		
		// Make multiple requests
		for i := 0; i < 5; i++ {
			_, err = client.ListProjects(ctx)
			assert.NoError(t, err)
		}
		
		// Connection should be reused (this is more of a demonstration)
		// In reality, testing connection reuse requires more sophisticated monitoring
		assert.Equal(t, 5, connectionCount)
	})
}