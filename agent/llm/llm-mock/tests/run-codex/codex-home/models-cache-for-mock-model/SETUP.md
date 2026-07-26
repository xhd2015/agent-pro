# Scenario

**Feature**: orchestrator seeds `CODEX_HOME/models_cache.json` so codex resolves `mock-model` metadata

Reproduces codex warning when `models_cache.json` is missing:

```
warning: Model metadata for `mock-model` not found. Defaulting to fallback metadata; ...
```

Codex looks up model metadata from `$CODEX_HOME/models_cache.json` (not `GET /v1/models` during exec).

## Preconditions

- Orchestrator has started mock server and written `config.toml` before codex child runs.
- Fake codex hook inspects `$CODEX_HOME/models_cache.json` after orchestrator setup.

## Steps

1. Use fake codex hook that prints whether `models_cache.json` contains slug `mock-model`.
2. Run `llm-mock run codex` with default config.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
const fakeCodexCheckModelsCache = `sh -c '
echo CODEX_HOME=$CODEX_HOME
if [ ! -f "$CODEX_HOME/models_cache.json" ]; then
  echo MODELS_CACHE=missing
  exit 0
fi
grep -q "\"slug\": \"mock-model\"" "$CODEX_HOME/models_cache.json" && echo MODELS_CACHE=mock-model-found || echo MODELS_CACHE=no-mock-model
'`

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ConfigEnv = "file"
	req.FakeCodexCmd = fakeCodexCheckModelsCache
	req.ConfigJSON = minimalMockConfigJSON(8080, "")
	return nil
}
```