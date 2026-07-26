# Scenario

**Feature**: BuildArgs emits -e KEY=VALUE for Env before open prompt

```
RunOpts{Open, Env: SLACK_MSG_SESSION_ID=…, SLACK_MSG_CONFIG=…}
  -> … --open -e SLACK_MSG_SESSION_ID=… -e SLACK_MSG_CONFIG=… -- <prompt>
```

## Preconditions

- Open profile so argv includes `--` separator.
- Env entries are full `KEY=VALUE` strings.

## Steps

1. Set session, open profile flags, two Env entries.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionID = "sess-env-1"
	req.Prompt = "with env"
	req.AgentRunner = "grok-tty"
	req.AutoSendOrResume = true
	req.NewTerminal = true
	req.Open = true
	req.Env = []string{
		"SLACK_MSG_SESSION_ID=sess-env-1",
		"SLACK_MSG_CONFIG=/tmp/slack-config.json",
	}
	return nil
}
```
