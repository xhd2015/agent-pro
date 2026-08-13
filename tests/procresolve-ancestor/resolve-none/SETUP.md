# Scenario

**Feature**: no grok ancestor → Kind=none, no error (descendant grok ignored)

```
FindAncestorGrok(start) -> ok=false
ResolveFromAncestors(start) -> Kind=none, err=nil
# a child grok of start must not become the hit
```

## Preconditions

- Start pid exists in the snapshot.
- No `IsGrokRunner` on the start-then-PPID chain.
- Leaves plant a descendant decoy grok with a session path so today’s
  `ResolveFromPID(start)` would hard-hit — ancestor APIs must still return none.

## Steps

1. Leaf installs a non-grok PPID chain plus decoy child grok.
2. Assert `AncestorOK=false` and `Kind=none`.

## Context

- Distinguishes ancestor miss from descendant-only resolve.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	if req.OpenFiles == nil {
		req.OpenFiles = map[int][]string{}
	}
	return nil
}
```
