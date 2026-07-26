# Scenario

**Feature**: list card content — Q/A full bodies, multi-line indent, corrupt skip

```
# seed session.data with messages / corrupt dirs
explain list -> formatted cards; full message bodies; indent multi-line; skip invalid dirs
```

## Preconditions

- Valid fixtures use agent_runner + model + alternating user/assistant messages.
- Corrupt fixtures use harness convention: empty AgentRunner/Model and nil Messages
  → directory created without session.data.

## Steps

1. Leaves seed content-specific fixtures.
2. Assert Q/A labels, time formatting, turn counts, full body, multi-line indent, skip rules.

## Context

- Turn count = number of user messages.
- Message roles map: user → `Q`, assistant → `A`.
- Full message body (no soft-truncate, no `…`); multi-line: first line after label,
  non-empty continuations indented 6 spaces; blank segments pure `\n`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if req.Bin == "" {
		t.Fatalf("content setup: explain binary not built")
	}
	return nil
}
```
