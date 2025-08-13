package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/conneroisu/hydra-go/hydra/models"
)

// Client represents a Hydra API client.
type Client struct {
	baseURL    string
	httpClient *http.Client
	userAgent  string

	// Session management
	Username string
	cookie   *http.Cookie
}

// Option represents a client configuration option.
type Option func(*Client)

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithUserAgent sets a custom user agent.
func WithUserAgent(userAgent string) Option {
	return func(c *Client) {
		c.userAgent = userAgent
	}
}

// WithTimeout sets the HTTP client timeout.
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

// NewClient creates a new Hydra API client.
func NewClient(baseURL string, opts ...Option) (*Client, error) {
	if baseURL == "" {
		return nil, errors.New("base URL cannot be empty")
	}

	// Ensure baseURL doesn't have trailing slash
	baseURL = strings.TrimSuffix(baseURL, "/")

	// Parse URL to validate it
	if _, err := url.Parse(baseURL); err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	// Create cookie jar for session management
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create cookie jar: %w", err)
	}

	c := &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Jar:     jar,
		},
		userAgent: "hydra-go-client/1.0.0",
	}

	// Apply options
	for _, opt := range opts {
		opt(c)
	}

	return c, nil
}

// DoRequest performs an HTTP request.
func (c *Client) DoRequest(ctx context.Context, method, path string, body interface{}, v interface{}) error {
	// Build URL
	u := c.baseURL + path

	// Prepare body
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, u, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("User-Agent", c.userAgent)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	// Add session cookie if available
	if c.cookie != nil {
		req.AddCookie(c.cookie)
	}

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Handle errors
	if resp.StatusCode >= 400 {
		var apiErr models.Error
		if err := json.Unmarshal(respBody, &apiErr); err != nil {
			return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
		}

		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    apiErr.Error,
		}
	}

	// Parse response if needed
	if v != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, v); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	// Store session cookie if this was a login request
	if strings.HasSuffix(path, "/login") && resp.StatusCode == http.StatusOK {
		for _, cookie := range resp.Cookies() {
			if cookie.Name == "hydra_session" {
				c.cookie = cookie

				break
			}
		}
	}

	return nil
}

// APIError represents an API error.
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error (status %d): %s", e.StatusCode, e.Message)
}

// IsAuthenticated returns true if the client has an active session.
func (c *Client) IsAuthenticated() bool {
	return c.cookie != nil
}

// GetUsername returns the username of the authenticated user.
func (c *Client) GetUsername() string {
	return c.Username
}

// Logout clears the session.
func (c *Client) Logout() {
	c.cookie = nil
	c.Username = ""
	// Clear cookies from jar
	if c.httpClient.Jar != nil {
		u, _ := url.Parse(c.baseURL)
		c.httpClient.Jar.SetCookies(u, []*http.Cookie{})
	}
}

// BaseURL returns the base URL of the client.
func (c *Client) BaseURL() string {
	return c.baseURL
}

// SetBaseURL updates the base URL of the client.
func (c *Client) SetBaseURL(baseURL string) error {
	baseURL = strings.TrimSuffix(baseURL, "/")
	if _, err := url.Parse(baseURL); err != nil {
		return fmt.Errorf("invalid base URL: %w", err)
	}
	c.baseURL = baseURL

	return nil
}
