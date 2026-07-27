# Scenario

**Feature**: no transcript found emits cleaned terminal scrollback

```
CODEX_HOME has no rollout JSONL
  -> fake TUI scrollback has fallback assistant text
  -> stdout includes fallback assistant text after PTY exit
```

## Preconditions

- This leaf covers the transcript-missing branch only.

## Steps

1. Run the inherited fake TUI without creating a transcript.
2. Assert fallback text is emitted.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.SkipCodexSessionDir = true
	return nil
}
```
