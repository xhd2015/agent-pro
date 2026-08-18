package agentruncli

import (
	"fmt"
	"time"

	"github.com/xhd2015/agent-pro/pkgs/agentrunapi"
)

// ParseRunIdle interprets launch-time idle flags after the CLI reads
// --exit-on-idle and the raw --idle-timeout string (empty = omitted).
// Invalid raw → error before NormalizeIdle.
// !exitOnIdle → enabled=false, no error even if timeout parses (including 2s).
// Does not start a TTY.
func ParseRunIdle(exitOnIdle bool, timeoutRaw string) (enabled bool, d time.Duration, err error) {
	if timeoutRaw == "" {
		return agentrunapi.NormalizeIdle(exitOnIdle, 0)
	}
	parsed, err := time.ParseDuration(timeoutRaw)
	if err != nil {
		return false, 0, fmt.Errorf("invalid value for --idle-timeout: %s", timeoutRaw)
	}
	return agentrunapi.NormalizeIdle(exitOnIdle, parsed)
}
