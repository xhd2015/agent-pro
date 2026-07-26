# Scenario

**Feature**: `llm-mock-run-grok` shortcut binary smoke test

```
llm-mock-run-grok [args] -> run.RunGrok() -> fake grok (exit 0)
```

## Steps

1. Set `UseShortcut` true to invoke shortcut binary directly.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.UseShortcut = true
	return nil
}
```
