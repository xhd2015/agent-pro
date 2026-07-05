# Scenario

**Feature**: `llm-mock run codex` subcommand smoke test

```
llm-mock run codex -> orchestrator -> fake codex (exit 0)
```

## Steps

1. Use default subcommand entry (`UseShortcut` false).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.UseShortcut = false
	return nil
}
```