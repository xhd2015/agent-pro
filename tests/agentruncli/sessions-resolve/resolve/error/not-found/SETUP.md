# Scenario

**Feature**: resolve not-found when UUID has no grok-family meta

```
seed unrelated session with different UUID
  -> RunSessions resolve --grok-session-id missing-UUID
  -> err not-found + UUID; stdout empty
```

## Steps

1. Seed one session bound to a different UUID.
2. Resolve a never-seen UUID.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Seeds = []SeedMeta{{
		SessionID:       "other-sess",
		Runner:          "grok-tty",
		RunnerSessionID: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
		Status:          "finished",
	}}
	req.Args = []string{"resolve", "--grok-session-id", "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"}
	return nil
}
```
