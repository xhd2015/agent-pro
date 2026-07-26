# Scenario

**Feature**: list with sessions still never starts the agent

```
# seed 1 session; EXPLAIN_AGENT_PATH=fake; explain list
-> exit 0, lists session, stderr has no FAKE_AGENT_INVOKED
```

## Preconditions

- Fake agent would fail loudly if started.

## Steps

1. Seed one valid session.
2. Args `list`.
3. Assert success list output and no fake invocation.

## Context

- Complements empty-store path which also must not call the agent.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"list"}
	req.Sessions = []SessionSeed{
		simpleSession(
			"2026-07-13-12-00-00-nollm-aaaaaaaa",
			"opencode", "deepseek-chat",
			"no llm q", "no llm a",
		),
	}
	return nil
}
```
