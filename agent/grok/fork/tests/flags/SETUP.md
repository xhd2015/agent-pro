# Scenario

**Feature**: parse-time flag conflicts and unknown input

```
fork.Main(conflicting or unknown args) -> error; no launch
```

## Preconditions

- Errors happen before ancestor resolve / session lookup.

## Steps

1. Leaf sets invalid Args.

## Context

- `--pid` + `--session-id` is locked as an error (C).

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.PID = pidStart
	return nil
}
```
