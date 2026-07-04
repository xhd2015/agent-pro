# Scenario

**Feature**: list returns sessions sorted by time.updated descending

```
# three sessions with known updated timestamps
writeOpencodeSession x3 -> sessions.List

# newest updated first; tie-break by session id when equal
[]Session ordered desc by time.updated
```

## Preconditions

- Three sessions with distinct `time.updated` values.

## Steps

1. Create sessions `ses_sort_03`, `ses_sort_02`, `ses_sort_01` with ascending updated times.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	req.Limit = 10
	base, err := time.Parse(time.RFC3339, "2026-06-23T10:00:00.000Z")
	if err != nil {
		t.Fatalf("parse base time: %v", err)
	}
	writeOpencodeSession(t, req.DataDir, "proj_sort", "ses_sort_01", "first",
		"/tmp/sort-a", base)
	writeOpencodeSession(t, req.DataDir, "proj_sort", "ses_sort_02", "second",
		"/tmp/sort-b", base.Add(1*time.Minute))
	writeOpencodeSession(t, req.DataDir, "proj_sort", "ses_sort_03", "third",
		"/tmp/sort-c", base.Add(2*time.Minute))
	return nil
}
```