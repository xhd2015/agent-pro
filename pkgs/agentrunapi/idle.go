package agentrunapi

import (
	"fmt"
	"strings"
	"time"
)

// DefaultIdleTimeout is used when ExitOnIdle is set and IdleTimeout is zero.
const DefaultIdleTimeout = 10 * time.Minute

// NormalizeIdle is the shared pure helper for CLI parse and BuildFollowUpCommand.
//
//	!exitOnIdle            → enabled=false (timeout ignored, no error)
//	exitOnIdle && timeout==0 → enabled=true, d=DefaultIdleTimeout
//	exitOnIdle && timeout>0  → enabled=true, d=timeout
//	exitOnIdle && timeout<0  → error (do not enable)
func NormalizeIdle(exitOnIdle bool, timeout time.Duration) (enabled bool, d time.Duration, err error) {
	if !exitOnIdle {
		return false, 0, nil
	}
	if timeout < 0 {
		return false, 0, fmt.Errorf("--idle-timeout must be a positive duration (got %s)", timeout)
	}
	if timeout == 0 {
		return true, DefaultIdleTimeout, nil
	}
	return true, timeout, nil
}

// FormatCompactDuration emits Go duration units without trailing zero parts
// (10m, 2m, 2s) instead of time.Duration.String() (10m0s, 2m0s).
func FormatCompactDuration(d time.Duration) string {
	if d == 0 {
		return "0s"
	}
	neg := d < 0
	if neg {
		d = -d
	}
	s := d.String()
	// "1h0m0s" / "10m0s" → drop trailing zero seconds, then zero minutes.
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
