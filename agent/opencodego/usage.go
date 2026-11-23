// Package opencodego reads OpenCode Go subscription usage over HTTP.
package opencodego

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	// DefaultBaseURL is the Go inference root. Usage is GET {base}/v1/usage.
	DefaultBaseURL = "https://opencode.ai/inference/go"
	// PathUsage is the usage meter path on DefaultBaseURL.
	PathUsage = "/v1/usage"
	// DefaultTokenFile is the plaintext API key llm-proxy also reads.
	DefaultTokenFile = "~/tmp/llm-token-opencode-go.txt"
	// UserAgent identifies this client; OpenCode rejects generic SDK names.
	UserAgent = "agent-pro-opencode-go/1.0"
	maxBody   = 1 << 20
)

// FetchOpts configures one usage GET.
type FetchOpts struct {
	// TokenFile is a plaintext file holding the Go API key. Empty uses
	// DefaultTokenFile. A nonempty Token skips the file.
	TokenFile string
	// Token is the Go API key. When set, TokenFile is not read.
	Token string
	// BaseURL overrides DefaultBaseURL (tests). Trailing slashes are stripped.
	BaseURL string
	// HTTP is the client; nil uses a 30s timeout client.
	HTTP *http.Client
}

// Meter is one Go limit window.
type Meter struct {
	Status   string `json:"status"`
	Percent  int    `json:"percent"`
	ResetsAt string `json:"resetsAt"`
}

// ResetTime parses ResetsAt. Zero means the API omitted a usable instant.
func (m Meter) ResetTime() time.Time {
	raw := strings.TrimSpace(m.ResetsAt)
	if raw == "" {
		return time.Time{}
	}
	if at, err := time.Parse(time.RFC3339Nano, raw); err == nil {
		return at
	}
	if at, err := time.Parse(time.RFC3339, raw); err == nil {
		return at
	}
	return time.Time{}
}

// Usage is the shared Go allowance as three meters. Paid models share this
// pool; the API does not report remaining usage per model.
type Usage struct {
	Rolling Meter `json:"rolling"`
	Weekly  Meter `json:"weekly"`
	Monthly Meter `json:"monthly"`
}

type usageEnvelope struct {
	Usage Usage `json:"usage"`
}

type apiErrorEnvelope struct {
	Type  string `json:"type"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

// Fetch GETs the Go usage meters with the given key.
func Fetch(ctx context.Context, opts FetchOpts) (*Usage, error) {
	token := strings.TrimSpace(opts.Token)
	if token == "" {
		path, err := ResolveTokenFile(opts.TokenFile)
		if err != nil {
			return nil, err
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read OpenCode token file %s: %w", path, err)
		}
		token = strings.TrimSpace(string(raw))
		if token == "" {
			return nil, fmt.Errorf("OpenCode token file %s is empty", path)
		}
	}
	base := strings.TrimRight(strings.TrimSpace(opts.BaseURL), "/")
	if base == "" {
		base = DefaultBaseURL
	}
	reqURL := base + PathUsage
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build OpenCode usage request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", UserAgent)

	client := opts.HTTP
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("OpenCode usage: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBody))
	if err != nil {
		return nil, fmt.Errorf("read OpenCode usage: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, parseAPIError(resp.StatusCode, body)
	}
	var envelope usageEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, fmt.Errorf("decode OpenCode usage: %w", err)
	}
	return &envelope.Usage, nil
}

func parseAPIError(status int, body []byte) error {
	var envelope apiErrorEnvelope
	if err := json.Unmarshal(body, &envelope); err == nil && envelope.Error != nil && strings.TrimSpace(envelope.Error.Message) != "" {
		kind := strings.TrimSpace(envelope.Error.Type)
		if kind == "" {
			kind = http.StatusText(status)
		}
		return fmt.Errorf("OpenCode usage: %s (%s %d)", envelope.Error.Message, kind, status)
	}
	msg := strings.TrimSpace(string(body))
	if msg == "" {
		msg = http.StatusText(status)
	}
	if len(msg) > 200 {
		msg = msg[:200]
	}
	return fmt.Errorf("OpenCode usage: %s (%d)", msg, status)
}

// ResolveTokenFile expands ~ and empty to DefaultTokenFile.
func ResolveTokenFile(path string) (string, error) {
	raw := strings.TrimSpace(path)
	if raw == "" {
		raw = DefaultTokenFile
	}
	if raw == "~" || strings.HasPrefix(raw, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("expand OpenCode token file: %w", err)
		}
		if raw == "~" {
			return home, nil
		}
		return filepath.Join(home, strings.TrimPrefix(raw, "~/")), nil
	}
	return raw, nil
}
