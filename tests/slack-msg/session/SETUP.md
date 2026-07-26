# Scenario

**Feature**: slack-msg session list / info / update / reply / history

```
Caller -> slack-msg session list|info|update|reply|history [options]
  -> resolve SLACK_MSG_SESSION_ID / --session-id (info/update/reply/history)
  -> load ~/.agent-pro/slack-local-bot/sessions.json
  -> list: sorted map rows (human table or --json)
  -> info: one entry + message_count + session_dir
  -> update: set dir (abs) on map entry
  -> reply: chat.postMessage channel-only (no thread_ts) + append messages.jsonl out
  -> history: print local messages.jsonl
```

## Preconditions

- Subcommand is always `session` as first arg; action is one of
  `list` | `info` | `update` | `reply` | `history`.
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
- List/info/update are local-store only (no live Slack).
- Success stdout ends with trailing `\n`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if req.WorkDir == "" {
		req.WorkDir = t.TempDir()
	}
	return nil
}
```
