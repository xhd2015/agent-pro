# Scenario

**Feature**: `agent-run run --help` lists `--no-submit`

```
agent-run run --help → stdout contains --no-submit; ends with newline
```

## Steps

1. Run `agent-run run --help` (args set by grouping).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if len(req.Args) < 2 || req.Args[0] != "run" || req.Args[1] != "--help" {
		req.Args = []string{"run", "--help"}
	}
	return nil
}
```
