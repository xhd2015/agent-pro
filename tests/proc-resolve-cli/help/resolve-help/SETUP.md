# Scenario

**Feature**: `agent-pro proc resolve -h` lists resolve and --json

```
agent-pro proc resolve -h
  -> exit 0
  -> combined output mentions "resolve" and "--json"
```

## Preconditions

- Prefer `-h` on the resolve subcommand (most specific help).
- If implementer only documents flags on `proc --help`, either form is fine as
  long as this leaf’s argv is wired; locked argv here is `proc resolve -h`.

## Steps

1. Set `Args = ["proc", "resolve", "-h"]`.
2. Assert exit 0; combined text has `resolve` and `--json`.

## Context

- Preferably also mentions `--ascii-tree` and/or `--no-enrich` (soft: not hard
  fail if missing; only resolve + --json are required).

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"proc", "resolve", "-h"}
	return nil
}
```
