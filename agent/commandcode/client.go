package commandcode

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// maxBodyBytes caps a response body read; usage payloads are a few KB.
const maxBodyBytes = 8 << 20

// APIError is a non-2xx response from the Command Code API.
type APIError struct {
	Method  string
	Path    string
	Status  int
	Code    string
	Message string
	Docs    string
	rawBody string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("%s: %s", strings.TrimSpace(e.Method+" "+e.Path), e.describe())
}

// describe renders the API's own message with its code and status, without the
// method and path that Error prefixes.
func (e *APIError) describe() string {
	msg := strings.TrimSpace(e.Message)
	if msg == "" {
		msg = strings.TrimSpace(e.rawBody)
	}
	if msg == "" {
		msg = http.StatusText(e.Status)
	}
	if e.Code != "" {
		return fmt.Sprintf("%s (%s %d)", msg, e.Code, e.Status)
	}
	return fmt.Sprintf("%s (%d)", msg, e.Status)
}

// Unauthorized reports whether the credential itself was rejected. A 403 is
// not an authentication failure: the credential is valid but lacks permission
// for the requested scope, so the caller reports the API's own message.
func (e *APIError) Unauthorized() bool {
	return e.Status == http.StatusUnauthorized
}

// NetworkError is a transport failure or an unreachable API.
type NetworkError struct {
	URL string
	Err error
}

func (e *NetworkError) Error() string {
	return fmt.Sprintf("network error: unable to reach %s: %v", e.URL, e.Err)
}

func (e *NetworkError) Unwrap() error { return e.Err }

// Client calls the Command Code HTTP API for one home.
type Client struct {
	BaseURL   string
	Home      string
	Auth      *Auth
	HTTP      *http.Client
	UserAgent string
}

// NewClient reads credentials for home and returns a client for baseURL.
// baseURL is resolved with ResolveBaseURL when empty.
func NewClient(home, baseURL string) (*Client, error) {
	resolvedHome, err := ResolveHome(home)
	if err != nil {
		return nil, err
	}
	auth, err := ReadAuth(resolvedHome)
	if err != nil {
		return nil, err
	}
	return &Client{
		BaseURL:   ResolveBaseURL(baseURL),
		Home:      resolvedHome,
		Auth:      auth,
		UserAgent: UserAgent,
	}, nil
}

func (c *Client) httpClient() *http.Client {
	if c.HTTP != nil {
		return c.HTTP
	}
	return &http.Client{Timeout: 30 * time.Second}
}

// Endpoint builds the request URL, omitting empty parameters the way the cmd
// CLI does (an empty orgId= is rejected with 400).
func (c *Client) Endpoint(path string, params map[string]string) string {
	base := strings.TrimRight(c.BaseURL, "/")
	query := url.Values{}
	for key, value := range params {
		if strings.TrimSpace(value) == "" {
			continue
		}
		query.Set(key, value)
	}
	if encoded := query.Encode(); encoded != "" {
		return base + path + "?" + encoded
	}
	return base + path
}

// Get performs a GET and decodes the JSON body into out. A nil out keeps the
// body undecoded.
func (c *Client) Get(ctx context.Context, path string, params map[string]string, out any) error {
	requestURL := c.Endpoint(path, params)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return fmt.Errorf("build request %s: %w", path, err)
	}
	req.Header.Set("Authorization", "Bearer "+c.Auth.APIKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return &NetworkError{URL: requestURL, Err: err}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes))
	if err != nil {
		return &NetworkError{URL: requestURL, Err: err}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return parseAPIError(http.MethodGet, path, resp.StatusCode, body)
	}
	if out == nil {
		return nil
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("%s: decode response: %w", path, err)
	}
	return nil
}

// parseAPIError maps an error body onto an APIError, falling back to the raw
// body when the payload is not the documented error envelope.
func parseAPIError(method, path string, status int, body []byte) error {
	apiErr := &APIError{Method: method, Path: path, Status: status}
	var envelope struct {
		Success bool `json:"success"`
		Error   *struct {
			Code    string `json:"code"`
			Status  int    `json:"status"`
			Message string `json:"message"`
			Docs    string `json:"docs"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &envelope); err == nil && envelope.Error != nil {
		apiErr.Code = envelope.Error.Code
		apiErr.Message = envelope.Error.Message
		apiErr.Docs = envelope.Error.Docs
		return apiErr
	}
	apiErr.rawBody = truncate(strings.TrimSpace(string(body)), 200)
	return apiErr
}

func truncate(s string, limit int) string {
	if len(s) <= limit {
		return s
	}
	return s[:limit] + "…"
}
