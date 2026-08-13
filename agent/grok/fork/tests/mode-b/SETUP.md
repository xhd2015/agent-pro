# Scenario

**Feature**: Mode B — `--session-id` forks in the current terminal

```
fork.Main(["--session-id", id])
  -> groksessions.Info
  -> RunForeground(grok-bin, --resume id --fork-session, cwd)
  -> never OpenInNewTerminal
```

## Preconditions

- Fixture session seeded; `GrokBin` basename is `llm-mock-run-grok`.
- No ancestor walk required (and `--pid` is forbidden with `--session-id`).

## Steps

1. Seed default session under GrokHome.
2. Leaf sets `--session-id` and optional flags.

## Context

- Optional future `--new-session-id` is out of scope.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	seedModeB(t, req)
	req.Args = []string{"--session-id", fixtureSessionID}
	return nil
}
```
