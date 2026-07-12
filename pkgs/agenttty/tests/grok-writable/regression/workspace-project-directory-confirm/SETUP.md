# Scenario

**Bug**: Grok project-directory confirmation TUI must not be classified as idle/sendable

```
snapshot: "Run Grok Build in a project directory?" + "(○)" options + "Enter:submit"
  -> CheckWritable returns ready=false, state ≠ idle
  -> send-queue must not inject into the picker
```

Smoke of `agent-run run --open` with workspace `/tmp` captured this modal (not an
idle prompt). While it is up, user input is not a grok session turn, so discovery
cannot bind and writable must report not ready.

## Preconditions

- Fixture `grok-workspace-project-directory-confirm.txt` from live smoke capture
  (`/tmp/grok-workspace-confirm-from-smoke.txt`).

## Steps

1. Set `req.FixtureFile` to the workspace project-directory confirm fixture.

## Context

- Complements F1 table row; explicit ASSERT for this regression alone (W2).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.FixtureFile = fixtureWorkspaceProjectDirectoryConfirm
	return nil
}
```
