# Scenario

**Factor**: CLI entry point — `llm-mock run codex` subcommand vs `llm-mock-run-codex` shortcut

```
# both call shared run.RunCodex()
llm-mock run codex [args] -> orchestrator
llm-mock-run-codex [args] -> orchestrator
```

## Preconditions

- Both entry points share `agent/llm/llm-mock/run.RunCodex()`.
- Smoke tests use fake codex hook and `LLM_MOCK_CONFIG_FILE`.

## Steps

1. Grouping `Setup` sets plumbing defaults (config file env, fake codex).
2. Leaf `Setup` toggles `UseShortcut` for shortcut binary.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ConfigEnv = "file"
	req.FakeCodexCmd = fakeCodexPrintHome
	req.ConfigJSON = minimalMockConfigJSON(8080, "")
	return nil
}
```