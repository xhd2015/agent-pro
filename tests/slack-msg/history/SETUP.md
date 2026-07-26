# Scenario

**Feature**: slack-msg history fetches conversation history or thread replies

```
Caller -> slack-msg history [options] [CHANNEL] -> conversations.history|replies -> oldest→newest lines or --json
```

## Preconditions

- Subcommand is always `history` as first arg.
- Channel via `--channel`, positional, `SLACK_CHANNEL`, or config default.

## Steps

1. Isolate workdir for history leaves.
2. Leaves set `req.Args` starting with `"history"`.
3. Unit leaves attach slacktest with history/replies handlers.

## Context

- API mock returns newest-first; CLI must print oldest→newest.
- Human format: `[ts] user: text` with trailing newline after last line.

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
