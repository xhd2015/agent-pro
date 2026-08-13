# Scenario

**Feature**: Mode A — bare grok-fork resolves nearest ancestor and opens a new window

```
# grok 4242 (session fixture) → bash → grok-fork 6000
fork.Main([] / --dry-run / --dir / --pid)
  -> OpenInNewTerminal(cwd, <exe> --session-id <id>)
```

## Preconditions

- Default ancestor chain and fixture session are seeded here.
- Leaves add flags / override procs.

## Steps

1. Seed default workspace session + Lsof on pid 4242.
2. Leaf sets Args.

## Context

- Default start pid 6000; ancestor grok 4242; session `019f283b-aaaa-…`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	seedModeA(t, req)
	return nil
}
```
