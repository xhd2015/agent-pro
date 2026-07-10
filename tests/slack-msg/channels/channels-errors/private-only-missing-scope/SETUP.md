# Scenario

**Feature**: sole --types private with missing_scope hard-fails (no soft-skip)

```
slack-msg channels list --types private --token
  -> private_channel missing_scope needed groups:read
  -> stderr channels failed: missing_scope (needed groups:read); see: slack-msg --help --topic add-missing-scope
  -> exit 1
```

## Steps

1. Use private-missing-scope mock (private always fails).
2. Request only private type.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ChannelsPrivateMissingScope = true
	req.Args = []string{
		"channels", "list",
		"--token", slackTestToken,
		"--types", "private",
	}
	return nil
}
```
