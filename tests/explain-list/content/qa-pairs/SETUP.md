# Scenario

**Feature**: multi-turn session prints all Q/A pairs with meta header

```
# 2 user + 2 assistant messages
explain list -> header (time, runner/model, 2 turns) + Q/A/Q/A lines
```

## Preconditions

- One session with two turns; known timestamp and model.

## Steps

1. Seed multi-turn session at `2026-07-13-14-30-05-...`.
2. Run plain list.
3. Assert full card layout (exact template).

## Context

- Locks the pretty format from the requirement example (single card).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"list"}
	req.Sessions = []SessionSeed{
		{
			DirName:     "2026-07-13-14-30-05-goroutine-abcd1234",
			AgentRunner: "opencode",
			Model:       "deepseek-chat",
			Messages: []Msg{
				{Role: "user", Message: "What is a goroutine?"},
				{Role: "assistant", Message: "A goroutine is a lightweight thread managed by the Go runtime."},
				{Role: "user", Message: "How does the scheduler work?"},
				{Role: "assistant", Message: "The Go scheduler multiplexes goroutines onto OS threads."},
			},
		},
	}
	return nil
}
```
