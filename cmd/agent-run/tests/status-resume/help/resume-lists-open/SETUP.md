# Scenario

**Feature**: `resume --help` lists `--open`, session-id / followup, and
`--grok-session-id`

```
agent-run resume --help -> --open, session-id, --grok-session-id
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
