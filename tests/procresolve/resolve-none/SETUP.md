# Scenario

**Feature**: no hard session — Kind=none without error

```
# no grok/codex session runner with parseable open files
ResolveFromPID -> Kind=none, SessionID="", Confidence="", err=nil
```

## Preconditions

- Pid exists in snapshot.
- Either no grok/codex candidates, or candidates have no session open paths
  (including excluded `grok update`).

## Steps

1. Leaf installs fixture without a resolvable session.
2. Assert Kind=none and empty SessionID; harness error is nil.

## Context

- Not an error path: callers treat `none` as soft miss.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Soft-miss branch: leaves must leave OpenFiles empty or non-session paths.
	if req.OpenFiles == nil {
		req.OpenFiles = map[int][]string{}
	}
	return nil
}
```
