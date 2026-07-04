# Scenario

**Feature**: `llm-mock run grok` subcommand smoke test

```
llm-mock run grok -> orchestrator -> fake grok (exit 0)
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
