# Scenario

**Feature**: agent-aware enrich of P1 iTerm2 snapshot for busy panes

```
# L2 inject path
Caller -> Capture(opts{Snapshot, ResolveFromPID, NoEnrich})
  -> base Snapshot preserved
  -> busy sessions -> ResolveFromPID -> Agents[session.ID]

# skip paths
NoEnrich | Idle|unknown | resolve none/error -> no Agents entry
```

## Preconditions

- Package under test:
  `github.com/xhd2015/agent-pro/pkgs/itermsnapshot` (absent until implementer —
  Classic TDD RED).
- Depends on P1 `shell/iterm2/snapshot` (types only for fixtures) and
  `pkgs/procresolve` (Result types for inject).
- All leaves are **L2 in-process**: inject `CaptureOpts.Snapshot` +
  `CaptureOpts.ResolveFromPID`. **No** real `osascript`, `ps`, `lsof`, or
  live iTerm. **No kool import.**
- Parallel-safe: each leaf owns request inject data; Capture is pure per-call
  opts (no package globals).

## Steps

1. Root `Setup` zeros request fields for the leaf chain.
2. Grouping/leaf `Setup` builds fixture Snapshot and ResolveFromPID inject.
3. Root `Run` calls `itermsnapshot.Capture` with request opts.
4. Leaf `Assert` checks Result.Agents / Snapshot / error only.

## Context

- Locked API: `Capture`, `CaptureOpts`, `Result`, `SessionAgent`,
  `AgentTreeNode`.
- Agents keyed by `SnapshotSession.ID`.
- Busy-only attach (Idle non-nil and false); soft resolve failures.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.NoEnrich = false
	req.Snapshot = nil
	req.ResolveFromPID = nil
	req.ResolveCallPIDs = nil
	return nil
}
```
