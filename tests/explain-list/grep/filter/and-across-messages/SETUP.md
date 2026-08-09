# Scenario

**Feature**: --and allows patterns to match different messages in one session

```
# session: Q has alpha, A has beta (neither has both)
# orphan: Q+A only alpha
# explain list --grep alpha --grep beta --and
-> cross-message session kept; orphan dropped
```

## Preconditions

- One multi-message session with split patterns; one incomplete session.

## Steps

1. Seed cross-message session and alpha-only session.
2. Args: `list --grep alpha --grep beta --and`.
3. Assert cross session kept via markers; orphan absent.

## Context

- AND is session-level, not per-message.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"list", "--grep", "alpha", "--grep", "beta", "--and"}
	req.Sessions = []SessionSeed{
		{
			DirName:     "2026-07-05-10-00-00-alphaonly-aaaaaaaa",
			AgentRunner: "opencode",
			Model:       "deepseek-chat",
			Messages: []Msg{
				{Role: "user", Message: "talk about alpha marker-orphan"},
				{Role: "assistant", Message: "still only alpha here"},
			},
		},
		{
			DirName:     "2026-07-06-10-00-00-cross-cccccccc",
			AgentRunner: "opencode",
			Model:       "deepseek-chat",
			Messages: []Msg{
				{Role: "user", Message: "talk about alpha marker-cross"},
				{Role: "assistant", Message: "and about beta in the answer"},
			},
		},
	}
	return nil
}
```
