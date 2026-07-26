# Scenario

**Factor**: isolated `GROK_HOME` — default temp dir vs explicit `LLM_MOCK_GROK_HOME`

```
orchestrator -> write GROK_HOME/config.toml (models_base_url -> mock)
orchestrator -> set GROK_HOME env for grok child
fake grok <- GROK_HOME path (printed or preset)
```

## Preconditions

- Orchestrator writes `config.toml` pointing all LLM traffic at the mock server.
- Default path is a fresh temp dir removed on exit.
- Explicit `LLM_MOCK_GROK_HOME` uses a caller-provided directory.

## Steps

1. Grouping `Setup` enables config.toml post-check and fake grok that prints `GROK_HOME`.
2. Leaf `Setup` chooses default temp vs explicit home.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ConfigEnv = "file"
	req.FakeGrokCmd = fakeGrokPrintHome
	req.ConfigJSON = minimalMockConfigJSON(8080, "")
	req.ExpectConfig = true
	return nil
}
```
