# Scenario

**Feature**: slack-msg agent.go wires interactive open to agentrunapi

```
# interactive
runAgentInteractiveOpen -> agentrunapi.AutoSendOrResume (not RunInteractiveOpen)

# driver
SLACK_LISTEN_AGENT_RUN empty -> agent-run default
SLACK_LISTEN_AGENT_RUN set   -> that binary as DriverBinary

# stateless
runAgentStateless -> agentrunbridge.Run(Stateless+CaptureStdout) OK
```

## Preconditions

- Source under test: `cmd/slack-msg/agent.go` relative to agent-pro module root.
- `DOCTEST_ROOT` is `tests/slack-msg-agent-wire` → module root `../..`.
- Classic RED until interactive open imports/calls agentrunapi.
- No network, no Slack API, no real agent-run process in these leaves.
- Large `tests/slack-msg/**` suite is separate regression (not inherited here).

## Steps

1. Root Setup is a no-op marker for shared helpers (path resolved in Run).
2. Grouping Setup sets `req.Mode`.
3. Run reads `agent.go` and fills Response flags.
4. Leaf Assert checks wire contracts.

## Context

- Env const name: `SLACK_LISTEN_AGENT_RUN`.
- Helper name: `agentRunBinary`.
- Empty binary policy: library default `"agent-run"` (BuildFollowUp / DriverBinary).

```go
import (
	"fmt"
	"strings"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	// Source path is resolved in Run via DOCTEST_ROOT; groupings/leaves set Mode.
	t.Helper()
	if req == nil {
		return fmt.Errorf("nil Request")
	}
	return nil
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected harness error: %v", err)
	}
}

func assertContains(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("missing %q in agent.go (%d bytes)", want, len(got))
	}
}

func assertNotContains(t *testing.T, got, forbidden string) {
	t.Helper()
	if strings.Contains(got, forbidden) {
		t.Fatalf("unexpected %q in agent.go", forbidden)
	}
}
```
