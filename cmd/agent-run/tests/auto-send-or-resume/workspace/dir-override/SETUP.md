# Scenario

**Feature**: resume `--dir` override wins over `meta.workspace` for child cwd

```
meta.workspace=/…/created-ws; --dir=/…/override-ws; CLI cwd=/…/cli-cwd
  -> agent-run resume --dir override-ws … "hi"
  -> child pwd = override-ws
```

## Steps

1. Grouping Setup created created-ws + cli-cwd.
2. Create override-ws directory.
3. Seed meta with Workspace=created-ws.
4. Run resume with `--dir` override.

```go
import (
	"os"
	"path/filepath"
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	override := filepath.Join(req.TempDir, "override-ws")
	if err := os.MkdirAll(override, 0755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(override, "README.md"), []byte("override\n"), 0644); err != nil {
		return err
	}
	req.DirOverride = override

	req.SessionID = "ws-dir-e3"
	req.RunnerSessionID = "550e8400-e29b-41d4-a716-446655440e33"
	req.MetaStatus = "finished"
	req.TerminalSessionID = "term-ws-e3"
	req.InitialPrompt = "created in ws dir override"
	req.WriteRegistry = false
	// req.Workspace still created-ws from grouping.
	seedBoundExitedDeadTerminal(t, req)

	installArgvCwdRunner(t, req)
	req.FollowupPrompt = "hi from dir override e3"
	req.Args = []string{
		"resume",
		"--dir", req.DirOverride,
		"--agent-runner-binary", req.AgentRunnerBinary,
		req.SessionID,
		req.FollowupPrompt,
	}
	req.ExecTimeout = 60 * time.Second
	req.Mode = "read-probes"
	return nil
}
```
