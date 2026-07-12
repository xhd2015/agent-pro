# Scenario

**Feature**: collision keeps newer updated_at at bare id; renames older to `{id}__{runner}`

```
fake-codex/shared (older) + fake-opencode/shared (newer)
-> sessions/shared (opencode) + sessions/shared__fake-codex
```

## Preconditions

- Both nested sessions share id `shared`.
- Newer has later `updated_at`.

## Steps

1. Seed older codex and newer opencode nested sessions with distinct events.
2. Run migrator.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	writeNestedSession(t, req.Home, "fake-codex", "shared", "finished", "2026-07-01T10:00:00Z", "from-codex")
	writeNestedSession(t, req.Home, "fake-opencode", "shared", "finished", "2026-07-01T12:00:00Z", "from-opencode")
	req.Args = []string{"--home", req.Home}
	return nil
}
```
