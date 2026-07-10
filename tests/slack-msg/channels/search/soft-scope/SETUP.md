# Scenario

**Feature**: channels search soft-skips private_channel missing_scope then filters public

```
# default multi-type; per-type list then name filter
public ok + private missing_scope -> filter QUERY on merged public rows
  <- stderr warning: skipped private channels (missing groups:read); see: slack-msg --help --topic add-missing-scope
  -> exit 0 (hit or empty)
```

## Preconditions

- Private-missing-scope slacktest server (`ChannelsPrivateMissingScope`).
- Same soft-fail rules as list; search only adds name filter after merge.

## Steps

1. Clear Slack env; enable private-missing-scope mock.
2. Leaf sets `channels search` QUERY (and optional `--json`).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ClearSlackEnv = true
	req.ChannelsPrivateMissingScope = true
	return nil
}
```
