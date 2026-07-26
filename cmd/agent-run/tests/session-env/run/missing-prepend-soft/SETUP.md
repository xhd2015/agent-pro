# Scenario

**Feature**: missing `--prepend-path` directory is soft-allowed

```
run --prepend-path /abs/missing-tools --agent-runner-binary env-logger "prompt"
  -> exit 0; PATH still contains abs missing path; meta stores it
```

## Steps

1. Choose a nonexistent absolute path under temp (do not create it).
2. Run with that `--prepend-path` and env-logging runner.

```go
import (
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	prepareEnvLoggingRun(t, req)
	// Soft-allow: path must not exist.
	req.PrependPathDir = absPath(t, filepath.Join(req.TempDir, "missing-tools-dir"))
	req.SessionID = "sess-env-missing-prepend"
	req.Prompt = "missing prepend soft"
	req.Args = append(req.Args,
		"--session-id", req.SessionID,
		"--prepend-path", req.PrependPathDir,
		"--agent-runner-binary", req.AgentRunnerBinary,
		req.Prompt,
	)
	return nil
}
```
