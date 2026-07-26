# Scenario

**Feature**: channels list soft-skips private_channel when multi-type and missing groups:read

```
# default --types public,private; per-type conversations.list
public_channel ok -> private_channel missing_scope needed groups:read
  -> stdout public channels only
  <- stderr warning: skipped private channels (missing groups:read); see: slack-msg --help --topic add-missing-scope
  -> exit 0
```

## Preconditions

- Private-missing-scope slacktest server (`ChannelsPrivateMissingScope`).
- Default multi-type (no `--types` override unless leaf sets one).
- Full-scope happy leaves stay on the default mock; this branch isolates soft-fail.

## Steps

1. Clear Slack env; enable private-missing-scope mock.
2. Leaf sets `channels list` args (human or `--json`).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ClearSlackEnv = true
	req.ChannelsPrivateMissingScope = true
	return nil
}
```
