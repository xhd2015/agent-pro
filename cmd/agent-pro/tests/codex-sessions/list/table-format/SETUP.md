# Scenario

**Feature**: list table output shows expected column headers

```
# one rollout session with known metadata
writeRolloutSession -> sessions.List -> FormatListTable

# table includes SESSION ID, STARTED, CWD columns
terminal table text
```

## Preconditions

- Table format is the default list output (not JSON).

## Steps

1. Create one session with id `01900004-aaaa-7aaa-aaaa-aaaaaaaaaaaa`.
2. Leave `req.Format` empty for table mode.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Limit = 10
	writeRolloutSession(t, req.CodexHome,
		"01900004-aaaa-7aaa-aaaa-aaaaaaaaaaaa",
		"2026-06-23T10:00:00.000Z", "/tmp/project-a")
	return nil
}
```