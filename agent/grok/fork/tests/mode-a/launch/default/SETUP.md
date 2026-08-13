# Scenario

**Feature**: bare grok-fork opens a new window at session cwd

```
fork.Main([])
  -> OpenInNewTerminal(session cwd, <exe> --session-id <id>)
  -> stdout: Opened new window; launching grok-fork --session-id <id>
```

## Steps

1. Args empty (bare Mode A).

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{}
	return nil
}
```
