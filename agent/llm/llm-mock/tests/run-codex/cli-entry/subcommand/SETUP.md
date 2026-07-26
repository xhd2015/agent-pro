# Scenario

**Feature**: `llm-mock run codex` subcommand smoke test

```
llm-mock run codex -> orchestrator -> fake codex (exit 0)
```

## Steps

1. Use default subcommand entry (`UseShortcut` false).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.UseShortcut = false
	return nil
}
```