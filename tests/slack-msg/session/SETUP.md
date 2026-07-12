# Scenario

**Feature**: slack-msg session reply / history (SeaTalk-like session bridge)

```
Caller -> slack-msg session reply|history [options]
  -> resolve SLACK_MSG_SESSION_ID / --session-id
  -> load ~/.agent-pro/slack-local-bot/sessions.json
  -> reply: chat.postMessage channel-only (no thread_ts) + append messages.jsonl out
  -> history: print local messages.jsonl
```

## Preconditions

- Subcommand is always `session` as first arg; action is `reply` or `history`.
- Session store under `$HOME/.agent-pro/slack-local-bot/` (isolate via `req.HomeDir`).
- Config for reply: `--config` / `SLACK_MSG_CONFIG` / map `config_path`.
- Unit reply leaves use `CapturePosts` + slacktest for PostMessage (no thread_ts).

## Steps

1. Isolate workdir for session leaves.
2. Leaves set `req.Args` starting with `"session"`.
3. Validation leaves clear Slack env; unit leaves seed map/log + HomeDir.

## Context

- Reply posts at channel top-level (no MsgOptionTS / thread_ts).
- History prefers local log over live API.
- Success stdout ends with trailing `\n`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if req.WorkDir == "" {
		req.WorkDir = t.TempDir()
	}
	return nil
}
```
