# Scenario

**Feature**: every seeded fixture matches its expectations.jsonl row

```
for each expectations.jsonl row
  -> read codex-*.txt
  -> CheckWritable
  -> assert ready/state/reason-substring
```

## Preconditions

- Full fixture table enabled by parent `fixture-table/SETUP.md`.

## Steps

1. Inherited `RunAllFixtures=true`; no per-leaf mutation beyond path defaults.

## Context

- Leaf is the runnable F1 gate; no additional `Setup` logic required beyond defaults.

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
		req.TestdataDir = filepath.Join(req.RepoRoot, "pkgs", "agenttty", "testdata", "codex-writable")
	}
	return nil
}
```
