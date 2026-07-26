# Scenario

**Feature**: `status --grok-session-id` with no matching meta exits not-found

```
empty home (or unrelated sessions) -> status --grok-session-id UUID
  -> exit 1; stderr mentions not found / grok session id
```

## Steps

1. Do not seed any meta with the requested UUID.
2. Run `status --grok-session-id <unknown-uuid>`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.RunnerSessionID = "550e8400-e29b-41d4-a716-446655440903"
	// Optional noise session with a different provider id (must not match).
	seedExtraSessionMeta(t, req,
		"test-gsid-noise-s1",
		"grok-tty",
		"550e8400-e29b-41d4-a716-446655440999",
		"finished",
		"term-gsid-noise-1",
		"noise session",
	)
	req.Args = []string{"status", "--grok-session-id", req.RunnerSessionID}
	return nil
}
```
