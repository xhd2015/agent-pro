# Scenario

**Feature**: `--no-children` / IncludeChildren=false skips child session dirs

```
seedStandardWorld (child exists on disk + linked in subagents meta)
  -> Backup(..., IncludeChildren=false)
  -> payload has parent; child session dir absent under payload/sessions
```

## Preconditions

- Child session directory exists under grok home (linked via subagents meta).
- `NoChildren` true → `IncludeChildren` pointer false.

## Steps

1. Use standard world.
2. Set `req.NoChildren = true`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.NoChildren = true
	return nil
}
```
