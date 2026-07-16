# Scenario

**Feature**: `agent-run run --help` lists `--detach`

```
agent-run run --help → stdout contains --detach; ends with newline
```

## Steps

1. Run `agent-run run --help`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"run", "--help"}
	return nil
}
```
