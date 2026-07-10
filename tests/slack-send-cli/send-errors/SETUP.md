# Scenario

**Feature**: send failures after channel resolution

```
slack-send -> resolve channel -> PostMessage or lookup fails -> stderr send failed: ... -> exit 1
```

## Preconditions

- `slacktest` with `conversations.list` (channel-not-found) or failing `chat.postMessage` (api-post-failed).

## Steps

1. Leaf selects error mode via slacktest server variant.
2. Provide token, channel, and message.
3. Assert stderr pattern; no `OK ts=` in stdout.

## Context

- Channel-not-found: name absent from list and knownChannels.
- API post failure: custom handler returns `ok: false`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ClearSlackEnv = true
	return nil
}
```