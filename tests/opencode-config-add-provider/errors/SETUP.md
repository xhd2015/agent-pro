# Scenario

**Feature**: add-provider rejects missing/invalid flags and duplicate ids

```
# invalid input -> non-zero exit, stderr mentions the offending flag/value, no file written
agent-pro opencode config add-provider <bad-args> -> error -> exit != 0
doctest <- config file absent (or unchanged for duplicate-id)
```

## Preconditions

- Each error leaf omits or invalidates exactly one mandatory input, or (for
  duplicate-provider-id) pre-seeds the target config with the same id.

## Steps

1. Set `req.Args` with the offending combination (and `req.PreConfig` for the
   duplicate-id case).
2. Run with isolated `HOME`.
3. Assert non-zero exit and that stderr mentions the relevant flag/value.
4. Assert no config file was written (or, for duplicate-id, that the file is
   byte-for-byte unchanged).

## Context

- Error leaves assert on `resp.Stderr` substrings (case-insensitive via
  `stderrContains`) and `resp.ExitCode != 0`. Exact wording is not asserted
  beyond presence of the named flag / valid values.

```go
import (
	"os"
	"strings"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	// Leaves under errors/ set req.Args (and req.PreConfig) in their own
	// Setup. Verify the shared harness is ready for this subtree: the binary
	// must have been built by the root Setup.
	if req.Bin == "" {
		t.Fatalf("errors setup: agent-pro binary not built (root Setup skipped?)")
	}
	return nil
}

// assertError asserts a non-zero exit and that stderr mentions each want
// substring (case-insensitive). Returns the lowercased stderr for further
// checks.
func assertError(t *testing.T, resp *Response, wants ...string) {
	t.Helper()
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit, got 0\nstdout:%s\nstderr:%s", resp.Stdout, resp.Stderr)
	}
	lower := strings.ToLower(resp.Stderr)
	for _, w := range wants {
		if !strings.Contains(lower, strings.ToLower(w)) {
			t.Fatalf("stderr missing %q:\n%s", w, resp.Stderr)
		}
	}
}

// assertNoConfigFile asserts the config file was not created.
func assertNoConfigFile(t *testing.T, resp *Response) {
	t.Helper()
	if _, err := os.Stat(resp.ConfigPath); err == nil {
		t.Fatalf("expected no config file at %s, but it exists", resp.ConfigPath)
	}
}
```
