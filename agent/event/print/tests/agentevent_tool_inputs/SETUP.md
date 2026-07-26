# Scenario

**Bug**: canonical AgentEvent search-like tool calls print only the `SEARCH` header

```
# knowledge-hub records canonical AgentEvent tool_call lines
AgentEvent{tool=glob, tool_input.pattern} -> compact trace printer

# compact output should include the pattern below the SEARCH header
compact trace printer -> SEARCH block with pattern detail
```

## Preconditions
- Canonical AgentEvent trace lines are formatted by `print.FormatTraceLine`.

## Steps
1. Leaf tests provide a canonical AgentEvent JSONL line in `req.Line`.
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
