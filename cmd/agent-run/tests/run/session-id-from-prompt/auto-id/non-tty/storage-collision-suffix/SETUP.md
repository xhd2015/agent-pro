# Scenario

**Feature**: when `base-timestamp` storage already exists, append `-1`, `-2`, …

```
seed sessions/fake-codex/hello-world-<nearby-ts>/meta.json
agent-run run --session-id-from-prompt "hello world"
  -> id hello-world-YYYYMMDD-HHMMSS-N  (N ≥ 1)
```

## Steps

1. Pre-create storage sessions for `hello-world-<ts>` across a local-time window.
2. Run auto-id with prompt `hello world`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Prompt = "hello world"
	seedStorageCollisionsForBase(t, req.Home, "fake-codex", "hello-world")
	req.Args = append(req.Args, req.Prompt)
	return nil
}
```
