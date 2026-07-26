# Scenario

**Feature**: `llm-mock-run-opencode` shortcut binary smoke test

```
llm-mock-run-opencode [args] -> run.RunOpencode() -> fake opencode (exit 0)
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