# Scenario

**Feature**: WaitReady (injectable status) + BuildFollowUpCommand driver + OpenInNewTerminal hook

```
# status
stdout -> ParseTTYStatus / IsSessionReadyFromStatus

# wait
StatusFn poll -> WaitReady -> ok | timeout

# follow-up
FollowUpOpts -> BuildFollowUpCommand -> shell-quoted line (no --new-terminal)

# open
OpenInNewTerminal(OpenTerminal hook) -> dir + followUp recorded
```

## Preconditions

- Package `github.com/xhd2015/agent-pro/pkgs/agentrunapi` P2 exports (classic RED
  until implementer lands them). See root DOCTEST planned API.
- **No real agent-run binary, PATH LookPath, iTerm, or grok** in unit leaves.
- Nested tree: `DOCTEST_ROOT` is `tests/agentrunapi/wait-driver`; module root is
  `../../..` relative to that root.
- Parent P1 tree (`tests/agentrunapi` leaves outside this nested root) must stay
  GREEN independently.

## Steps

1. Root `Setup` seeds default session id / prompt / runner for follow-up leaves.
2. Grouping `Setup` sets `req.Mode`.
3. Leaf fills fixtures; `Run` calls P2 APIs; `Assert` checks outcomes.

## Context

- Status fixtures: `statusReadyFixture()` / `statusNotReadyFixture()` in DOCTEST.md.
- Default SessionID=`sess-wait-1`, Prompt=`hello wait-driver`, AgentRunner=`grok-tty`.

```go
import (
	"strings"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	if req.SessionID == "" && req.Mode != "status" && req.Mode != "source_wire" {
		// wait_ready missing-session-id leaf clears SessionID after this when needed.
		req.SessionID = "sess-wait-1"
	}
	if req.Prompt == "" {
		req.Prompt = "hello wait-driver"
	}
	if req.AgentRunner == "" && (req.Mode == "follow_up" || req.Mode == "open_new_terminal") {
		req.AgentRunner = "grok-tty"
	}
	return nil
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected harness error: %v", err)
	}
}

func assertAPIError(t *testing.T, resp *Response) {
	t.Helper()
	if resp == nil || resp.ErrString == "" {
		t.Fatal("expected API error, got nil/empty")
	}
}

func assertNoAPIError(t *testing.T, resp *Response) {
	t.Helper()
	if resp != nil && resp.ErrString != "" {
		t.Fatalf("unexpected API error: %s", resp.ErrString)
	}
}

func assertEqual(t *testing.T, field string, got, want any) {
	t.Helper()
	if got != want {
		t.Fatalf("%s: got %#v, want %#v", field, got, want)
	}
}

func assertContains(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("missing %q in %q", want, got)
	}
}

func assertNotContains(t *testing.T, got, forbidden string) {
	t.Helper()
	if strings.Contains(got, forbidden) {
		t.Fatalf("unexpected %q in %q", forbidden, got)
	}
}

func assertContainsFold(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(strings.ToLower(got), strings.ToLower(want)) {
		t.Fatalf("missing %q in %q", want, got)
	}
}
```
