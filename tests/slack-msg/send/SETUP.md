# Scenario

**Feature**: slack-msg send posts a single MESSAGE via chat.postMessage

```
Caller -> slack-msg send [options] MESSAGE -> resolve token/channel -> PostMessage -> OK ts=... channel=...
```

## Preconditions

- Subcommand is always `send` as first arg.
- Exactly one positional MESSAGE required.

## Steps

1. Isolate workdir for send leaves.
2. Leaves set `req.Args` starting with `"send"`.
3. Validation leaves clear Slack env; unit leaves attach slacktest via `SLACK_API_URL`.

## Context

- Port of `tests/slack-send-cli` with `slack-msg send` argv.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if req.WorkDir == "" {
		req.WorkDir = t.TempDir()
	}
	return nil
}
```
