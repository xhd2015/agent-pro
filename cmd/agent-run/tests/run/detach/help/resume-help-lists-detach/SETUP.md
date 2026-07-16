# Scenario

**Feature**: `agent-run resume --help` lists `--detach`

```
agent-run resume --help → stdout contains --detach; ends with newline
```

## Steps

1. Run `agent-run resume --help`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"resume", "--help"}
	return nil
}
```
