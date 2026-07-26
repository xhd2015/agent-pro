# Scenario

**Factor**: CLI entry point — `llm-mock run opencode` subcommand vs `llm-mock-run-opencode` shortcut

```
# both call shared run.RunOpencode()
llm-mock run opencode [args] -> orchestrator
llm-mock-run-opencode [args] -> orchestrator
```

## Preconditions

- Both entry points share `agent/llm/llm-mock/run.RunOpencode()`.
- Smoke tests use fake opencode hook and `LLM_MOCK_CONFIG_FILE`.

## Steps

1. Grouping `Setup` sets plumbing defaults (config file env, fake opencode).
2. Leaf `Setup` toggles `UseShortcut` for shortcut binary.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ConfigEnv = "file"
	req.FakeOpencodeCmd = fakeOpencodePrintConfigDir
	req.ConfigJSON = minimalMockConfigJSON(8080, "")
	return nil
}
```