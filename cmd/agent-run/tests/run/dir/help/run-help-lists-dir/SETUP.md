# Scenario

**Feature**: `agent-run run --help` lists `--dir`

```
agent-run run --help → stdout contains --dir; ends with newline
```

## Steps

1. Run `agent-run run --help` (args set by parent help grouping).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if len(req.Args) < 2 || req.Args[0] != "run" || req.Args[1] != "--help" {
		req.Args = []string{"run", "--help"}
	}
	return nil
}
```
