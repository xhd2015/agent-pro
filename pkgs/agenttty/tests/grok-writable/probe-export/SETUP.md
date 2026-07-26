# Scenario

**Feature**: grok-writable-probe exports curated fixtures from an existing capture dir

```
mini capture dir (captures.jsonl + snapshots/)
  -> go run ./script/debug/grok-writable-probe -export-fixtures=<out> -from=<capture>
  -> grok-*.txt + expectations.jsonl
```

## Preconditions

- Implementer adds `-export-fixtures` and `-from` flags to `script/debug/grok-writable-probe`.
- Leaf `testdata/mini-capture/` holds a trimmed probe run (3 unique hashes).

## Steps

1. Leaf `Setup` points `ExportFromDir` at bundled mini capture.
2. `Run` shells out to probe export into `t.TempDir()`.

## Context

- F5 round-trip; RED until export flags exist on the probe binary.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func miniCaptureDir(d *session.Doctest) string {
	return filepath.Join(d.DOCTEST_ROOT, "probe-export", "capture-dir-round-trip", "testdata", "mini-capture")
}

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.RunAllFixtures = false
	req.FixtureFile = ""
	return nil
}
```