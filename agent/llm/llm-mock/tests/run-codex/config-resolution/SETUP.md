# Scenario

**Factor**: config source — none (default empty) vs `LLM_MOCK_CONFIG_FILE`

```
orchestrator -> mockconfig loader
  # file env -> config path; neither -> default {"exchanges": []}
mockconfig <- LLM_MOCK_CONFIG_FILE / (none)
```

## Preconditions

- Config resolution is a high-impact parameter for run-codex.
- Success leaves use fake codex (`LLM_MOCK_RUN_CODEX_COMMAND`) and minimal mock config when file env is set.
- `no-config-starts-codex` omits config env and config file; expects exit 0 with fake codex.

## Steps

1. Grouping `Setup` sets default fake codex and minimal config JSON.
2. Leaf `Setup` chooses `ConfigEnv` mode (`file` or neither).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.FakeCodexCmd = fakeCodexPrintHome
	req.ConfigJSON = minimalMockConfigJSON(8080, "")
	return nil
}
```