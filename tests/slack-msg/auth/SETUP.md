# Scenario

**Feature**: slack-msg auth status inspects bot or app token validity

```
Caller -> slack-msg auth status [--app] [options]
  -> Using config from: <abs-path>|(none)
  -> bot: auth.test | app: apps.connections.open
  -> human or --json status (masked token) -> exit 0|1
```

## Preconditions

- Subcommand is always `auth` as first arg; action is `status`.
- Bot token via `--token`, `SLACK_BOT_TOKEN`, or config `botToken`.
- App token via `--app-token`, `SLACK_APP_TOKEN`, or config `appToken`.
- Unit leaves attach slacktest (`auth.test` success includes `bot_id`;
  `apps.connections.open` available by default).

## Steps

1. Isolate workdir for auth leaves.
2. Leaves set `req.Args` starting with `"auth"`.
3. Validation leaves clear Slack env; unit leaves set `SLACK_API_URL`.

## Context

- Always print `Using config from:` (absolute path or `(none)`).
- Masking: type prefix + `...` + last 4 chars; never full secret on stdout.
- Success stdout ends with trailing `\n`.
- App validation contract: **`apps.connections.open`** (Socket Mode / connections).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if req.WorkDir == "" {
		req.WorkDir = t.TempDir()
	}
	return nil
}
```
