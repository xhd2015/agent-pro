# Scenario

**Feature**: multiple busy sessions attach independent agents by session ID

```
two busy sessions (pid A/B) + resolveByPID
  -> Agents[sess-a]=grok, Agents[sess-b]=codex (independent)
```

## Steps

1. Build Snapshot with two busy sessions in one tab.
2. Resolve maps PID → distinct hard hits.

```go
import (
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/procresolve"
	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2/snapshot"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	const (
		idA  = "sess-multi-a"
		idB  = "sess-multi-b"
		pidA = 7101
		pidB = 7202
	)
	req.Snapshot = &snapshot.Snapshot{
		CapturedAt: "2026-07-25T12:00:00Z",
		Host:       "testhost",
		Source:     "iterm2",
		Summary:    snapshot.SnapshotSummary{Windows: 1, Tabs: 1, Sessions: 2, Busy: 2},
		Windows: []snapshot.SnapshotWindow{
			{
				Index: 1,
				Name:  "Multi",
				Tabs: []snapshot.SnapshotTab{
					{
						Index: 1,
						Name:  "Both",
						Sessions: []snapshot.SnapshotSession{
							{
								Index: 1, ID: idA, Name: "pane-a", TTY: "/dev/ttys021",
								Profile: "Default", Idle: boolPtr(false), PID: intPtr(pidA),
							},
							{
								Index: 2, ID: idB, Name: "pane-b", TTY: "/dev/ttys022",
								Profile: "Default", Idle: boolPtr(false), PID: intPtr(pidB),
							},
						},
					},
				},
			},
		},
	}
	req.ResolveFromPID = resolveByPID(map[int]*procresolve.Result{
		pidA: {
			InputPID: pidA, Kind: "grok", SessionID: "grok-a", Confidence: "hard",
			GrokTitle: "Agent A",
			Tree:      []procresolve.ProcNode{{PID: pidA, PPID: 1, Role: "input", Cmd: "zsh"}},
		},
		pidB: {
			InputPID: pidB, Kind: "codex", SessionID: "codex-b", Confidence: "hard",
			Tree: []procresolve.ProcNode{{PID: pidB, PPID: 1, Role: "input", Cmd: "bash"}},
		},
	})
	return nil
}
```
