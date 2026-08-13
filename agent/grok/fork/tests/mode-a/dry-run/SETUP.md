# Scenario

**Feature**: Mode A `--dry-run` prints the new-window plan and does not launch

```
fork.Main(["--dry-run"])
  -> stdout locked plan (ancestor 4242, grok id, cwd, command)
  -> OpenInNewTerminal not called
```

## Preconditions

- Default ancestor + fixture session from parent Setup.

## Steps

1. Args `["--dry-run"]`.

## Context

- Command must use injected Executable + `--session-id`, not `grok --resume`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"--dry-run"}
	return nil
}
```
