# Scenario

**Feature**: consolidated trace parsing under agent/event

```
# adapters register on import; line parser and aggregator consume JSONL
trace JSONL line -> adapter registry -> AgentTraceParsedEvent

# multi-line traces merge tool lifecycle
trace lines[] + created_at -> message aggregator -> timeline messages[]
```

## Preconditions
- Target packages `agent/event/traceparse`, `traceview`, and `summary` exist.
- Blank import of `agent/event/traceparse` registers provider adapters.

## Steps
1. Ensure doctest harness helpers are available for JSON assertions.

```go
import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = json.Unmarshal
	_ = strings.Contains
	return nil
}

func assertContains(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("missing %q in:\n%s", want, got)
	}
}

func assertNotContains(t *testing.T, got, want string) {
	t.Helper()
	if strings.Contains(got, want) {
		t.Fatalf("unexpected %q in:\n%s", want, got)
	}
}
```
