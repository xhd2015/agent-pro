# Scenario

**Feature**: resolving a non-existent pid exits non-zero with an error on stderr

```
agent-pro proc resolve 999999999
  -> exit != 0
  -> stderr mentions not found / pid
```

## Preconditions

- No `AGENT_PRO_PROCRESOLVE_TEST_SNAPSHOT` (live path or production empty).
- Pid `999999999` is extremely unlikely to exist; if it somehow does, leaf may
  flake — implementer may treat snapshot-unset + missing pid via live list.

## Steps

1. Args: `proc resolve 999999999`.
2. Assert non-zero exit and stderr error token.

## Context

- Error text should mention `pid` and/or `not found` (same spirit as library
  `pid not found`). Case-insensitive match on `not found` or `pid`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"proc", "resolve", "999999999"}
	req.Snapshot = nil
	return nil
}
```
