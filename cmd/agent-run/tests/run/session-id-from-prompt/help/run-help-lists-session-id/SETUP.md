# Scenario

**Feature**: `agent-run run --help` documents `--session-id` alias

```
agent-run run --help -> mentions --session-id (alias of --session)
```

## Steps

1. Invoke `run --help`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"run", "--help"}
	return nil
}
```
