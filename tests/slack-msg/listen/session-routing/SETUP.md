# Scenario

**Feature**: session mode routes via agentrunbridge (+ SYSTEM.md / map / env inject)

```
inbound event -> thread: RunInteractiveOpen + sessions.json + messages.jsonl + SYSTEM.md + open inject
                 Env: -e SLACK_MSG_SESSION_ID=… [-e SLACK_MSG_CONFIG=…]
             -> stateless: Run+CaptureStdout (no SYSTEM.md / map)
  -> mock agent-run argv log (launch only; tty status not counted)
```

## Preconditions

- Default `--session-mode thread` unless leaf sets stateless.
- Session id format (stable keys; not per-message ts):
  - Channel / group / MPIM (non-DM): `slack-channel-{channelID}` as `--session-id=…`
  - DM (`D…` channel): `slack-dm-{userID}` as `--session-id=…`
- Thread: `run` + `--auto-send-or-resume` + `--new-terminal` + `--open` (not `--keep-tty`, not `send`).
- Thread store under `$HOME/.agent-pro/slack-local-bot/` (sessions.json, SYSTEM.md, messages.jsonl).
- SYSTEM.md recipes: `session history` / `session reply` only (no raw send --channel/--thread).
- Stateless: `run` without `--open` / session-id; no SYSTEM.md / session map.
- Leaves that assert store paths set `req.HomeDir` for isolation.
- Event **dedupe** remains `channelID:ts` (orthogonal to session id).

## Steps

1. Start daemon with session mode flags.
2. Inject one or two events in same thread.
3. Assert launch argv matches open profile or stateless `run` (and SYSTEM.md when applicable).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ClearSlackEnv = true
	prependListenTokens(req)
	req.Daemon = true
	req.Args = []string{"--no-require-mention"}
	return nil
}
```
