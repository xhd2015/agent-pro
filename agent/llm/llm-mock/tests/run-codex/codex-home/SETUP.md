# Scenario

**Factor**: isolated `CODEX_HOME` — default temp dir vs explicit `LLM_MOCK_CODEX_HOME`

```
orchestrator -> write CODEX_HOME/config.toml (base_url -> mock, wire_api=responses)
orchestrator -> set CODEX_HOME env for codex child
fake codex <- CODEX_HOME path (printed or preset)
```

## Preconditions

- Orchestrator writes codex `config.toml` pointing LLM traffic at the mock server.
- Default path is a fresh temp dir removed on exit.
- Explicit `LLM_MOCK_CODEX_HOME` uses a caller-provided directory.

## Steps

1. Grouping `Setup` enables config.toml post-check and fake codex that prints `CODEX_HOME`.
2. Leaf `Setup` chooses default temp vs explicit home.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ConfigEnv = "file"
	req.FakeCodexCmd = fakeCodexPrintHome
	req.ConfigJSON = minimalMockConfigJSON(8080, "")
	req.ExpectConfig = true
	return nil
}
```