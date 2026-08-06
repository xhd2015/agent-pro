# Scenario

**Feature**: dry-run still errors when session has a live grok PID

```
seed parent (inactive in active_sessions)
  + DryRun + inject ListProcs/Lsof hard hit
  -> Backup -> error; nothing written under OutDir
```

## Preconditions

- Session exists; **not** listed in `active_sessions.json`.
- Injected process is a grok runner with open-file hard hit on session id.
- Dry-run does **not** bypass the busy gate.

## Steps

1. Ensure inactive active_sessions list.
2. Inject one matching live PID.
3. Set OutDir; keep DryRun.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	writeActiveSessions(t, req.GrokHome /* none */)
	req.Procs = []FixtureProc{
		{PID: 6201, PPID: 1, Cmd: "/usr/local/bin/grok"},
	}
	req.OpenFiles = map[int][]string{
		6201: {grokOpenPath(req.SessionID)},
	}
	req.OutDir = filepath.Join(req.TempDir, "dry-run-pid-live-out")
	req.ArchivePath = ""
	return nil
}
```
