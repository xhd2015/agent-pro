# Scenario

**Feature**: session reply posts channel-level message for bound session

```
slack-msg session reply [options] MESSAGE
  -> resolve session map + config/token
  -> chat.postMessage(channel_id) without thread_ts
  -> OK ts=… channel=…
```

## Preconditions

- Action subcommand is `reply` as second arg.
- Exactly one positional MESSAGE required on success paths.

## Steps

1. Leaves set args starting with `"session", "reply"`.
2. Error leaves clear env; success leaves seed map + CapturePosts.

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
