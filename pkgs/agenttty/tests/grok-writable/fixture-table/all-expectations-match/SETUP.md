# Scenario

**Feature**: every seeded fixture matches its expectations.jsonl row

```
for each expectations.jsonl row
  -> read grok-*.txt
  -> CheckWritable
  -> assert ready/state/reason
```

## Preconditions

- Full fixture table enabled by parent `fixture-table/SETUP.md`.

## Steps

1. Inherited `RunAllFixtures=true`; no per-leaf mutation.

## Context

- Leaf is the runnable F1 gate; no additional `Setup` logic required.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.RunAllFixtures = true
	req.FixtureFile = ""
	if req.TestdataDir == "" {
		if req.RepoRoot == "" {
			req.RepoRoot = filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", "..", "..", ".."))
		}
		req.TestdataDir = filepath.Join(req.RepoRoot, "pkgs", "agenttty", "testdata", "grok-writable")
	}
	return nil
}
```