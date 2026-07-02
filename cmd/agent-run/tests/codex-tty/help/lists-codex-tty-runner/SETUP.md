# Scenario

**Feature**: runner help lists `codex-tty` as a supported backend

```
agent-run run --help → stdout contains codex-tty
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
