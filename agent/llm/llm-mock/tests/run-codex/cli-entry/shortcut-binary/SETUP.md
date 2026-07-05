# Scenario

**Feature**: `llm-mock-run-codex` shortcut binary smoke test

```
llm-mock-run-codex [args] -> run.RunCodex() -> fake codex (exit 0)
```

## Steps

1. Set `UseShortcut` true to invoke shortcut binary directly.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.UseShortcut = true
	return nil
}
```