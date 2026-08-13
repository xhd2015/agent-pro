# Scenario

**Feature**: Mode A launch records OpenInNewTerminal (no real iTerm2)

```
fork.Main(bare or flags) -> OpenInNewTerminal once -> success stdout
```

## Preconditions

- Not `--dry-run` unless a leaf adds it (none do here).

## Steps

1. Leaves set Args / override fixture.
2. Assert one open call + success line.

## Context

- Default auto color is off (pipe / buffer).

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	if req.Args == nil {
		req.Args = []string{}
	}
	return nil
}
```
