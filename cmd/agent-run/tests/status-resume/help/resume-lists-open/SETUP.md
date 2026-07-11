# Scenario

**Feature**: `resume --help` lists `--open` and session-id / followup

```
agent-run resume --help -> --open, session-id
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
