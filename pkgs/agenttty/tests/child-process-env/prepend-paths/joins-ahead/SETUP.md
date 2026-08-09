# Scenario

**Feature**: S6 — prependPaths → PATH starts with joined prefixes

```
# S6
prependPaths=["/opt/agent-bin-a", "/opt/agent-bin-b"]
  -> Set PATH value starts with /opt/agent-bin-a + sep + /opt/agent-bin-b
```

## Steps

1. Set PrependPaths to two absolute-style prefixes.
2. Assert PATH key present and value has expected prefix.

## Context

- Empty/whitespace-only path parts should be skipped by production (not asserted here).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.PrependPaths = []string{"/opt/agent-bin-a", "/opt/agent-bin-b"}
	return nil
}
```
