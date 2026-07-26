# Scenario

**Feature**: default list limit returns 20 newest of many sessions

```
# 25 session JSON files with ascending time.updated
writeOpencodeSession x25 -> sessions.List(dataDir, 20)

# only the 20 most recent sessions are returned
[]Session (len=20) -> newest timestamps kept
```

## Preconditions

- Default limit is 20 when `req.Limit` is zero.
- Twenty-five distinct session files exist.

## Steps

1. Create 25 sessions with `time.updated` from `2026-06-23T00:00:00.000Z`
   through `2026-06-23T00:24:00.000Z` (one per minute).
2. Leave `req.Limit` at zero to exercise the default.

```go
import (
	"fmt"
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	base, err := time.Parse(time.RFC3339, "2026-06-23T00:00:00.000Z")
	if err != nil {
		t.Fatalf("parse base time: %v", err)
	}
	for i := 0; i < 25; i++ {
		id := fmt.Sprintf("ses_limit_%02d", i+1)
		updated := base.Add(time.Duration(i) * time.Minute)
		writeOpencodeSession(t, req.DataDir, "proj_limit", id,
			fmt.Sprintf("session %d", i+1), "/tmp/opencode-limit", updated)
	}
	return nil
}
```