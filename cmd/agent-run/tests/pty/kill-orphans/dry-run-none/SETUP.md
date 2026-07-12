# Scenario

**Feature**: dry-run with `--exe` unique path and no matching serves prints no-match

```
agent-run pty kill-orphans --dry-run --exe /tmp/no-such-agent-run-pty-test-bin
  -> clear "no orphans" / "no matching serves" line
  -> exit 0; trailing \n
```

## Steps

1. Do not spawn any serve.
2. Run dry-run with an executable path that cannot match host serves.
3. Assert exit 0, no-match wording, trailing newline.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.Mode = "kill-orphans"
	req.SpawnServe = false
	// Unique path under leaf temp dir — never equals a real agent-run on the host.
	uniqueExe := filepath.Join(req.TempDir, "no-such-agent-run-for-pty-dry-run")
	req.Args = []string{
		"pty", "kill-orphans",
		"--dry-run",
		"--exe", uniqueExe,
	}
	return nil
}
```
