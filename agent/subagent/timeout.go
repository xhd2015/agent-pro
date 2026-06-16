package subagent

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// ParseTimeoutDuration parses a duration string for the --timeout flag.
//   - Empty input defaults to 1 hour.
//   - Bare numbers (no suffix) are treated as seconds.
//   - Otherwise, parsed with time.ParseDuration.
//   - Whitespace is trimmed before parsing.
//   - Duration < 1 minute returns an error.
//   - 1m ≤ duration < 10m prints a warning to stderr.
func ParseTimeoutDuration(s string) (time.Duration, error) {
	if s == "" {
		return time.Hour, nil
	}

	s = strings.TrimSpace(s)

	hasLetters := false
	for _, ch := range s {
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') {
			hasLetters = true
			break
		}
	}

	var d time.Duration
	var err error
	if !hasLetters {
		d, err = time.ParseDuration(s + "s")
	} else {
		d, err = time.ParseDuration(s)
	}
	if err != nil {
		return 0, err
	}

	if d < time.Minute {
		return 0, fmt.Errorf("timeout must be at least 1m")
	}

	if d < 10*time.Minute {
		fmt.Fprintf(os.Stderr, "this sub-agent may take a while to finish, it is suggested to set to a longer timeout, e.g. --timeout=1h\n")
	}

	return d, nil
}

// TestExported_parseTimeoutDuration wraps ParseTimeoutDuration for doctest access.
func TestExported_parseTimeoutDuration(s string) (time.Duration, error) {
	return ParseTimeoutDuration(s)
}
