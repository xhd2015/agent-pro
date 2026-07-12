# Scenario

**Feature**: RunInteractiveOpen forwards Env to agent-run -e flags

```
InteractiveOpenOpts{Env: SLACK_MSG_SESSION_ID=…, SLACK_MSG_CONFIG=…}
  -> launch argv contains -e pairs (same as BuildArgs fill)
```

## Preconditions

- Spy via fake RunCommand captures launch argv.
- Wait-ready sequence succeeds (default status hold/seq ready).

## Steps

1. Mode interactive_open; set session, prompt, Env.
2. Status poll ready so wait succeeds.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Mode = "interactive_open"
	req.SessionID = "sess-io-env"
	req.Prompt = "open with env"
	req.AgentRunner = "grok-tty"
	req.Env = []string{
		"SLACK_MSG_SESSION_ID=sess-io-env",
		"SLACK_MSG_CONFIG=/abs/config.json",
	}
	// Parent interactive-open Setup scripts status ready when seq/hold empty.
	return nil
}
```
