# Scenario

**Feature**: unknown `--kind` value is rejected with exit 1

```
agent-run pty kill-orphans --kind=not-a-real-kind --dry-run --exe <unique>
  -> exit 1
  -> stderr explains invalid/unknown kind
```

## Preconditions

- No serve spawn required.
- Unique `--exe` path so a mis-implemented filter cannot hit host processes.

## Steps

1. Run kill-orphans with an unknown kind value.
2. Assert exit 1 and stderr error; do not require successful stdout trailing `\n`.

```go
import (
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Mode = "kill-orphans"
	req.SpawnServe = false
	req.SpawnPlan = nil
	uniqueExe := filepath.Join(req.TempDir, "no-such-agent-run-for-unknown-kind")
	req.Args = []string{
		"pty", "kill-orphans",
		"--kind=not-a-real-kind",
		"--dry-run",
		"--exe", uniqueExe,
	}
	return nil
}
```
