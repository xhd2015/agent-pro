# Scenario

**Feature**: Mode B happy record — foreground grok --resume --fork-session

```
fork.Main(["--session-id", id])
  -> one RunForeground; dir=session cwd; basename llm-mock-run-grok
```

## Steps

1. Args `--session-id` only.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"--session-id", fixtureSessionID}
	return nil
}
```
