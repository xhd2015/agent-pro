package agentstorage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DefaultIdleTimeout is the compact 10m default when exit-on-idle is on
// and the stored duration is zero. Matches agentrunapi.DefaultIdleTimeout.
const DefaultIdleTimeout = 10 * time.Minute

// IdlePolicy is the session-dir idle-exit file (not meta.json).
type IdlePolicy struct {
	ExitOnIdle  bool          `json:"exit_on_idle"`
	IdleTimeout time.Duration `json:"-"`
}

type idlePolicyWire struct {
	ExitOnIdle  bool   `json:"exit_on_idle"`
	IdleTimeout string `json:"idle_timeout"`
}

// IdlePolicyPath is $home/sessions/<sessionID>/idle-policy.json.
func IdlePolicyPath(home, sessionID string) string {
	return filepath.Join(home, "sessions", sessionID, "idle-policy.json")
}

// WriteIdlePolicy writes compact JSON. Zero timeout + ExitOnIdle → 10m.
func WriteIdlePolicy(home, sessionID string, p IdlePolicy) error {
	timeout := p.IdleTimeout
	if p.ExitOnIdle && timeout == 0 {
		timeout = DefaultIdleTimeout
	}
	path := IdlePolicyPath(home, sessionID)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("idle-policy dir: %w", err)
	}
	body, err := json.Marshal(idlePolicyWire{
		ExitOnIdle:  p.ExitOnIdle,
		IdleTimeout: formatCompactDuration(timeout),
	})
	if err != nil {
		return err
	}
	body = append(body, '\n')
	if err := os.WriteFile(path, body, 0o644); err != nil {
		return fmt.Errorf("write idle-policy: %w", err)
	}
	return nil
}

// ReadIdlePolicy loads idle-policy.json. Missing file → found=false, err=nil.
func ReadIdlePolicy(home, sessionID string) (p IdlePolicy, found bool, err error) {
	path := IdlePolicyPath(home, sessionID)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return IdlePolicy{}, false, nil
		}
		return IdlePolicy{}, false, err
	}
	var wire idlePolicyWire
	if err := json.Unmarshal(data, &wire); err != nil {
		return IdlePolicy{}, false, fmt.Errorf("idle-policy.json: %w", err)
	}
	d, err := time.ParseDuration(wire.IdleTimeout)
	if err != nil {
		return IdlePolicy{}, false, fmt.Errorf("idle-policy.json idle_timeout: %w", err)
	}
	return IdlePolicy{ExitOnIdle: wire.ExitOnIdle, IdleTimeout: d}, true, nil
}

func formatCompactDuration(d time.Duration) string {
	if d == 0 {
		return "0s"
	}
	neg := d < 0
	if neg {
		d = -d
	}
	s := d.String()
	if strings.HasSuffix(s, "m0s") {
		s = strings.TrimSuffix(s, "0s")
	}
	if strings.HasSuffix(s, "h0m") {
		s = strings.TrimSuffix(s, "0m")
	}
	if neg {
		return "-" + s
	}
	return s
}
