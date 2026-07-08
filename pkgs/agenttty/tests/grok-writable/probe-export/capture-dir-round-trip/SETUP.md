# Scenario

**Feature**: export from mini capture produces parseable fixtures and manifest

```
-export-fixtures=<tmpdir> -from=testdata/mini-capture
  -> at least one grok-*.txt
  -> expectations.jsonl parses
```

## Preconditions

- Mini capture bundled alongside this leaf (`testdata/mini-capture/`).

## Steps

1. Set `req.ExportFromDir` to `miniCaptureDir()`.
2. `Run` leaves `ExportToDir` empty so export writes to `t.TempDir()`.

## Context

- Validates probe export reproduces the testdata layout documented in README.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ExportFromDir = miniCaptureDir()
	return nil
}
```