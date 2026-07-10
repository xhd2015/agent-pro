# Scenario

**Feature**: channels API and hard-failure paths

```
# non-scope API error
slack-msg channels list -> conversations.list error -> stderr channels failed: -> exit 1

# sole-type missing_scope
slack-msg channels list --types private -> missing_scope needed groups:read
  -> channels failed: missing_scope (needed groups:read); see: slack-msg --help --topic add-missing-scope
  -> exit 1

# all types missing_scope (no successful type)
slack-msg channels list -> every type missing_scope
  -> channels failed: …; see: slack-msg --help --topic add-missing-scope -> exit 1
```

## Preconditions

- Leaf selects failure mode: `ChannelsAPIFail`, `ChannelsPrivateMissingScope`,
  or `ChannelsAllMissingScope`.
- Soft-skip multi-type partial success lives under `list/soft-scope` and
  `search/soft-scope`, not here.
- Missing-scope hard-fail stderr appends the help topic pointer
  `see: slack-msg --help --topic add-missing-scope` (non-scope API errors need not).

## Steps

1. Clear Slack env; leaf enables fail server and args.
2. Assert hard-fail stderr and exit 1 (no partial channel stdout).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ClearSlackEnv = true
	return nil
}
```
