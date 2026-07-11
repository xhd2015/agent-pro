# Scenario

**Feature**: `status --help` documents session-ref and multi-layer probe

```
agent-run status --help -> mentions session id / layers (or session-ref)
```

## Steps

1. Run `agent-run status --help`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"status", "--help"}
	return nil
}
```
