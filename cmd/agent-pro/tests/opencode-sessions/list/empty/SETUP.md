# Scenario

**Feature**: empty storage tree yields no sessions

```
# DataDir exists but storage/session has no JSON files
sessions.List(dataDir, limit) -> empty slice

# table formatter reports no sessions
FormatListTable -> "No sessions found"
```

## Preconditions

- No session JSON files are written in this leaf.

## Steps

1. Do not write any session fixtures (root temp dir only).
2. Set `req.Limit = 20`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Limit = 20
	return nil
}
```