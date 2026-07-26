# Scenario

**Feature**: codex-show-status surfaces failures on stderr with non-zero exit

```
# failure condition -> non-zero exit, stderr explains reason, stdout empty or partial
codex-show-status -> error path -> exit != 0
doctest <- stderr: codex / timeout / parse message
```

## Preconditions

- Each error leaf isolates one failure mode: missing codex, timeout, or parse error.
- Success stdout (three clean lines) must not be produced.

## Steps

1. Leaf `Setup` configures PATH, fake TUI variant, or timeout override.
2. Run CLI with isolated env.
3. Assert non-zero exit and stderr mentions the expected failure keyword.

## Context

- Error leaves use `assertError` helper (case-insensitive stderr substring match).
- `codex-not-found` uses minimal PATH and omits `CODEX_SHOW_STATUS_COMMAND`.

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if req.Bin == "" {
		t.Fatalf("errors setup: codex-show-status binary not built (root Setup skipped?)")
	}
	return nil
}

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

func assertStdoutNotSuccessLines(t *testing.T, resp *Response) {
	t.Helper()
	trimmed := strings.TrimSpace(resp.Stdout)
	if strings.Contains(trimmed, "Monthly usage:") &&
		strings.Contains(trimmed, "Credits used:") &&
		strings.Contains(trimmed, "Next reset:") {
		t.Fatalf("expected incomplete/failed output, got success-like stdout:\n%s", resp.Stdout)
	}
}
```