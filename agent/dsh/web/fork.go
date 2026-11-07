package web

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
)

// ForkOptions describes a child conversation seeded from another session.
// The Host copies the source transcript up to AtSeq and starts the child in Cwd.
type ForkOptions struct {
	// SessionID is the fork source and is required.
	SessionID string
	// AtSeq is the inclusive cut; 0 lets the Host keep the latest completed turn.
	AtSeq int64
	// Cwd is the child checkout. It must be fully qualified: the Host matches it
	// against its workspace registry and silently keeps the source for an
	// unknown directory, so callers verify the returned child.
	Cwd string
}

// Fork creates a child session and returns its ID. It neither seeds a draft nor
// opens a browser; agent/dsh/fork composes both so the child can be reviewed
// before its first turn.
func (c Client) Fork(ctx context.Context, opts ForkOptions) (string, error) {
	source := strings.TrimSpace(opts.SessionID)
	if source == "" {
		return "", fmt.Errorf("DSH fork requires a source session ID")
	}
	if opts.AtSeq < 0 {
		return "", fmt.Errorf("DSH fork cut must not be negative")
	}
	request := map[string]any{"sessionId": source}
	if opts.AtSeq > 0 {
		request["atSeq"] = opts.AtSeq
	}
	if checkout := strings.TrimSpace(opts.Cwd); checkout != "" {
		if !filepath.IsAbs(checkout) {
			return "", fmt.Errorf("fork cwd must be a fully qualified directory")
		}
		request["cwd"] = checkout
	}
	var value struct {
		SessionID string `json:"sessionId"`
	}
	if err := c.rpc(ctx, "session/fork", map[string]any{"request": request}, &value); err != nil {
		return "", err
	}
	if strings.TrimSpace(value.SessionID) == "" {
		return "", fmt.Errorf("DSH session/fork returned no session ID")
	}
	return value.SessionID, nil
}
