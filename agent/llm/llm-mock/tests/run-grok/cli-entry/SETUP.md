# Scenario

**Factor**: CLI entry point — `llm-mock run grok` subcommand vs `llm-mock-run-grok` shortcut

```
# both call shared run.RunGrok()
llm-mock run grok [args] -> orchestrator
llm-mock-run-grok [args] -> orchestrator
```

## Preconditions

- Both entry points share `agent/llm/llm-mock/run.RunGrok()`.
- Smoke tests use fake grok hook and `LLM_MOCK_CONFIG_FILE`.

## Steps

1. Grouping `Setup` sets plumbing defaults (config file env, fake grok).
2. Leaf `Setup` toggles `UseShortcut` for shortcut binary.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ConfigEnv = "file"
	req.FakeGrokCmd = fakeGrokPrintHome
	req.ConfigJSON = minimalMockConfigJSON(8080, "")
	return nil
}
```
