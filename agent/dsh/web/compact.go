package web

import (
	"context"
	"fmt"
	"strings"
)

// CompactIntoNew resets one Session into a new one whose seed is its compacted
// context: the Host compacts the source, materializes the resulting surface
// (summary plus the retained tail) as the child's log, and returns the child
// identity. The child log starts at the compacted size instead of copying the
// source transcript, which is what Fork does.
func (c Client) CompactIntoNew(ctx context.Context, sessionID string) (string, error) {
	source := strings.TrimSpace(sessionID)
	if source == "" {
		return "", fmt.Errorf("DSH compacted reset requires a source session ID")
	}
	var value struct {
		SessionID string `json:"sessionId"`
	}
	if err := c.rpc(ctx, "session/compactIntoNew", map[string]any{"request": map[string]string{"sessionId": source}}, &value); err != nil {
		return "", err
	}
	if strings.TrimSpace(value.SessionID) == "" {
		return "", fmt.Errorf("DSH session/compactIntoNew returned no session ID")
	}
	return value.SessionID, nil
}
