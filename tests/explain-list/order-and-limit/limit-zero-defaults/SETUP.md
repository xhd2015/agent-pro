# Scenario

**Feature**: --limit 0 is treated as default limit 10

```
# 12 sessions; explain list --limit 0 -> same as default: 10 newest
```

## Preconditions

- Twelve sessions so default-10 clipping is observable.

## Steps

1. Seed 12 sessions.
2. Args: `list --limit 0`.
3. Assert same clipping as default-limit-10.

## Context

- Spec: if `N <= 0`, treat as default 10 (also covers negative if implementer maps similarly; this leaf locks zero).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"list", "--limit", "0"}
	req.Sessions = seedNSessions(12, 10)
	return nil
}
```
