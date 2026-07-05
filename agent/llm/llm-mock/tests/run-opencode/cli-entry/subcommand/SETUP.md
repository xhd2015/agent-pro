# Scenario

**Feature**: `llm-mock run opencode` subcommand smoke test

```
llm-mock run opencode -> orchestrator -> fake opencode (exit 0)
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