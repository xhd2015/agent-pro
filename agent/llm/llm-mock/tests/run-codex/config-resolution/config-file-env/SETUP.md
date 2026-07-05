# Scenario

**Feature**: `LLM_MOCK_CONFIG_FILE` resolves config and runs fake codex successfully

```
orchestrator -> mockconfig loader <- LLM_MOCK_CONFIG_FILE
orchestrator -> fake codex (exit 0)
```

## Steps

1. Set `ConfigEnv` to `file` so `Run` exports `LLM_MOCK_CONFIG_FILE`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ConfigEnv = "file"
	return nil
}
```