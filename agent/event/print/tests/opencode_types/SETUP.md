# Scenario

**Bug**: native opencode tool input details are dropped from compact trace output

```
# opencode emits a native tool_use JSONL event with structured tool input
opencode tool_use line -> opencode trace adapter -> AgentTraceActivity

# the compact printer should render both the header and useful input details
AgentTraceActivity -> compact trace printer -> terminal text
```

## Preconditions
- Native opencode trace lines are parsed through the opencode trace adapter and
  formatted by `print.FormatTraceLine`.

## Steps
1. Leaf tests provide a native opencode JSONL line in `req.Line`.
2. The shared print test harness formats that line.

```go
import (
	"strings"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Line = strings.TrimSpace(req.Line)
	_ = assertContains
	return nil
}

func assertContains(t *testing.T, got string, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("missing %q in:\n%s", want, got)
	}
}
```
