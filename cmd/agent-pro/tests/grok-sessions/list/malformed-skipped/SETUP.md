# Scenario

**Feature**: malformed summary.json files are skipped during discovery

```
# two valid sessions plus corrupt summary.json entries
writeGrokSession x2 + writeRawSummaryFile (invalid) -> sessions.List

# only parseable sessions with valid metadata are returned
[]Session (len=2)
```

## Preconditions

- Mix of valid `summary.json` and corrupt or incomplete files on disk.
- Encoded cwd directory `encoded-a` is shared by one valid and one malformed entry.

## Steps

1. Write two valid sessions with distinct ids and timestamps.
2. Write invalid JSON in `bad-json/summary.json`.
3. Write JSON missing `info.id` in `missing-id/summary.json`.
4. Write empty file in `empty-file/summary.json`.
5. Set `req.Limit = 10`.

```go
import (
	"net/url"
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Limit = 10
	cwd := "/tmp/malformed-mix"
	absCwd, err := filepath.Abs(cwd)
	if err != nil {
		t.Fatalf("abs cwd: %v", err)
	}
	encoded := url.PathEscape(absCwd)

	writeGrokSession(t, req.GrokHome,
		"01900005-0000-7000-8000-000000000001",
		"2026-06-20T10:00:00.000Z", cwd, "valid one")
	writeGrokSession(t, req.GrokHome,
		"01900005-0000-7000-8000-000000000002",
		"2026-06-20T12:00:00.000Z", cwd, "valid two")

	writeRawSummaryFile(t, req.GrokHome, encoded, "bad-json", "{not valid json")
	writeRawSummaryFile(t, req.GrokHome, encoded, "missing-id",
		`{"info":{"cwd":"/tmp/malformed-mix"},"last_active_at":"2026-06-20T11:00:00.000Z"}`)
	writeRawSummaryFile(t, req.GrokHome, encoded, "empty-file", "")
	return nil
}
```