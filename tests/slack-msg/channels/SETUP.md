# Scenario

**Feature**: slack-msg channels lists or searches workspace channels

```
Caller -> slack-msg channels list|search [options] [QUERY] -> conversations.list -> human/JSON rows
```

## Preconditions

- Subcommand is always `channels` as first arg; action is `list` or `search`.
- Token via `--token`, `SLACK_BOT_TOKEN`, or config `botToken`.

## Steps

1. Isolate workdir for channels leaves.
2. Leaves set `req.Args` starting with `"channels"`.
3. Unit leaves attach slacktest with conversations.list fixtures.

## Context

- Exclude archived by default; sort by name ascending.
- Human: `id  #name  kind  member|-` (two spaces between columns).
- JSON: `{"channels":[{id,name,is_private,is_member,is_archived}]}` trailing `\n`.
- Empty results: empty human stdout or `{"channels":[]}\n`; exit 0.

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
