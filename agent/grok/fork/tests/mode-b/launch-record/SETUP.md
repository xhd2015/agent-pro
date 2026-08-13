# Scenario

**Feature**: Mode B record path — RunForeground, no new window

```
fork.Main(["--session-id", id])
  -> RunForeground(grok-bin, [--resume id --fork-session], session cwd)
  -> OpenInNewTerminal not called
```

## Steps

1. Leaves keep `--session-id` (and optional `--dir`).

## Context

- Recorder only; grok bin does not need to exist.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	if len(req.Args) == 0 {
		req.Args = []string{"--session-id", fixtureSessionID}
	}
	return nil
}
```
