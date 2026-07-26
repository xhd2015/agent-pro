# Scenario

**Feature**: resume `--prepend-path` appends to stored paths and persists

```
seed meta.prepend_paths=[orig]
  -> resume --prepend-path /more <id> "followup"
  -> PATH starts with orig then /more
  -> meta.prepend_paths = [orig, /more]
```

## Steps

1. Seed meta with one stored prepend path.
2. Resume with an additional `--prepend-path` (different abs dir).

```go
import (
	"os"
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	prepareEnvLoggingResume(t, req)

	req.SessionID = "sess-env-append-path"
	req.PrependPathDir = absPath(t, filepath.Join(req.TempDir, "tools-orig"))
	if err := os.MkdirAll(req.PrependPathDir, 0755); err != nil {
		return err
	}
	req.PrependPathMore = absPath(t, filepath.Join(req.TempDir, "tools-more"))
	if err := os.MkdirAll(req.PrependPathMore, 0755); err != nil {
		return err
	}

	req.SeedPrependPaths = []string{req.PrependPathDir}
	seedBoundExitedMeta(t, req)

	req.Prompt = "resume append prepend"
	req.Args = []string{
		"resume",
		"--prepend-path", req.PrependPathMore,
		"--agent-runner-binary", req.AgentRunnerBinary,
		req.SessionID,
		req.Prompt,
	}
	return nil
}
```
