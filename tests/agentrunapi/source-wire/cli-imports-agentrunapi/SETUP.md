# Scenario

**Feature**: agent-run main package imports agentrunapi

```
parse cmd/agent-run/*.go imports
  -> pkgs/agentrunapi present
```

## Preconditions

- Module layout: DOCTEST_ROOT = tests/agentrunapi → `../../cmd/agent-run`.

## Steps

1. Mode already `source_wire` from parent.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	// Parent set source_wire; nothing else.
	if req.Mode != "source_wire" {
		req.Mode = "source_wire"
	}
	return nil
}
```
