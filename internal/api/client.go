package api

import (
	"bytes"
	"io"
	"net/http"
)

const defaultBaseURL = "https://www.cursor.com"

// Client makes authenticated requests to the Cursor dashboard API.
type Client struct {
	baseURL      string
	sessionToken string
	httpClient   *http.Client
}

// Option configures a Client.
type Option func(*Client)

// WithBaseURL overrides the default base URL (useful for testing).
func WithBaseURL(url string) Option {
	return func(c *Client) {
		c.baseURL = url
	}
}

// WithHTTPClient overrides the default http.Client.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) {
		c.httpClient = hc
	}
}

// NewClient creates a Client with the given session token and options.
func NewClient(sessionToken string, opts ...Option) *Client {
	c := &Client{
		baseURL:      defaultBaseURL,
		sessionToken: sessionToken,
		httpClient:   http.DefaultClient,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// BaseURL returns the client's base URL.
func (c *Client) BaseURL() string {
	return c.baseURL
}

// Do executes an HTTP request with the session cookie attached.
// For POST/PUT/PATCH requests, it also sets the Origin header (CSRF protection).
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	req.AddCookie(&http.Cookie{
		Name:  "WorkosCursorSessionToken",
		Value: c.sessionToken,
	})

	switch req.Method {
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		req.Header.Set("Origin", c.baseURL)
	}

	return c.httpClient.Do(req)
}

// CheckResponse reads the response and returns an error for non-2xx status codes.
// On success, it returns an io.Reader containing the response body.
// The caller does not need to close the original response body; it is fully read.
func (c *Client) CheckResponse(resp *http.Response) (io.Reader, error) {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return bytes.NewReader(body), nil
	}

	return nil, &APIError{
		StatusCode: resp.StatusCode,
		Body:       string(body),
	}
}
