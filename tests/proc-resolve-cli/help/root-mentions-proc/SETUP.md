# Scenario

**Feature**: agent-pro root help indexes the `proc` command (resolve)

```
agent-pro -h|--help
  -> exit 0
  -> combined output mentions "proc" and "resolve"
```

## Preconditions

- P5 polish backfill: production root help already lists `proc` with a resolve
  description; this leaf locks the product-ready top-level index.
- Subcommand flag help (`proc resolve -h` → `--json`, etc.) stays in
  `help/resolve-help` (P2).

## Steps

1. Set `Args = ["-h"]` (root help).
2. Assert exit 0; combined text has `proc` and `resolve`.

## Context

- Either `-h` or `--help` is fine; leaf locks `-h`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"-h"}
	return nil
}
```
