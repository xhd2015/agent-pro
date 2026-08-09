# Scenario

**Feature**: list --grep never starts the agent runner

```
# seed matching session; EXPLAIN_AGENT_PATH=fake; explain list --grep marker
-> exit 0, lists session, stderr has no FAKE_AGENT_INVOKED
```

## Preconditions

- Fake agent would fail loudly if started (root harness).

## Steps

1. Seed one matching session.
2. Args: `list --grep marker-nollm`.
3. Assert success list + no fake invocation.

## Context

- Complements plain `dispatch/no-llm-on-list` for the grep code path.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"list", "--grep", "marker-nollm"}
	req.Sessions = []SessionSeed{
		simpleSession(
			"2026-07-18-16-00-00-nollm-grep-aaaaaaaa",
			"opencode", "deepseek-chat",
			"grep path marker-nollm", "no llm a",
		),
	}
	return nil
}
```
