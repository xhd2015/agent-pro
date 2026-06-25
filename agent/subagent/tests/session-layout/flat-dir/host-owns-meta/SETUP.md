# Scenario

**Feature**: flat layout with HostOwnsMeta — subagent skips meta I/O, reports via callback

```
# host pre-writes meta.json; subagent trusts SessionID + ResumeInnerSessionID
host meta.json -> subagent.Run(HostOwnsMeta) -> OnAgentComplete(InnerSessionID) ; no meta write
```

## Preconditions

- `SessionLayout.Dir` points to flat session directory with host-owned `meta.json`.
- `HostOwnsMeta` is true on `subagent.Options`.

## Steps

1. Descendant `Setup` creates flat dir, writes host meta, enables `HostOwnsMeta`.
2. `invokeRun` wires `OnAgentComplete` to capture `AgentRunInfo` on `Request`.

## Context

- Legacy `merged-meta` / `meta-alias` leaves keep `HostOwnsMeta=false` (subagent patches meta).
- Descendant leaves assert no meta writes and callback/resume behavior.

```go
import (
	"testing"

	"github.com/xhd2015/agent-pro/agent/subagent"
)

func enableHostOwnsMeta(t *testing.T, req *Request) {
	t.Helper()
	req.HostOwnsMeta = true
}

func wireOnAgentComplete(req *Request) {
	req.OnAgentComplete = func(info subagent.AgentRunInfo) error {
		req.CallbackCalled = true
		req.CallbackInnerSessionID = info.InnerSessionID
		req.CallbackAgentRunner = info.AgentRunner
		return nil
	}
}

func Setup(t *testing.T, req *Request) error {
	enableHostOwnsMeta(t, req)
	wireOnAgentComplete(req)
	return nil
}```
