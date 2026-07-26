# Scenario

**Feature**: `llm-mock run grok` subcommand smoke test

```
llm-mock run grok -> orchestrator -> fake grok (exit 0)
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
