# Scenario

**Feature**: explain list filters and highlights sessions via --grep

```
# seed sessions under isolated config home
caller: explain list --grep P... [--or|--and] [--limit N] [--color]
  -> sort newest-first -> filter message bodies -> limit
  -> matching cards (optional bold-red spans) | empty/no-match msg | Error
doctest <- stdout / stderr / exit code
```

## Preconditions

- Root harness builds `explain`, isolates ConfigHome, points EXPLAIN_AGENT_PATH
  at the failing fake agent.
- Grep is case-insensitive **literal** substring match on Q/A message bodies
  only (not agent_runner, model, or dirname).
- Multiple greps default to OR; `--and` is session-level (patterns may hit
  different messages). `--or` is explicit OR.
- Empty pattern / conflicting mode flags / mode without greps → hard error.

## Steps

1. Groupings and leaves seed distinctive sessions and set Args.
2. Assert keep-set, title match totals, empty/no-match wording, highlight SGR,
   or stderr `Error:` as appropriate.

## Context

- Pipeline: sort → filter → total = match count → apply limit.
- Title shape: `Recent explain sessions (%d shown of %d, limit %d)`.
- No storage format change; full bodies still printed.

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if req.Bin == "" {
		t.Fatalf("grep setup: explain binary not built")
	}
	return nil
}

// assertGrepError asserts non-zero exit, stderr has "Error:", and each want
// substring appears (case-insensitive).
func assertGrepError(t *testing.T, resp *Response, wants ...string) {
	t.Helper()
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit for grep error, got 0\nstdout:\n%s\nstderr:\n%s",
			resp.Stdout, resp.Stderr)
	}
	if !strings.Contains(resp.Stderr, "Error:") {
		t.Fatalf("stderr must contain \"Error:\":\n%s", resp.Stderr)
	}
	lower := strings.ToLower(resp.Stderr)
	for _, w := range wants {
		if !strings.Contains(lower, strings.ToLower(w)) {
			t.Fatalf("stderr missing %q:\n%s", w, resp.Stderr)
		}
	}
}
```
