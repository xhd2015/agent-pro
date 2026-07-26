# Scenario

**Feature**: default isolated CODEX_HOME is a fresh temp directory

```
orchestrator -> create temp CODEX_HOME + config.toml
fake codex -> echo CODEX_HOME=$CODEX_HOME
```

## Steps

1. Do not set `LLM_MOCK_CODEX_HOME` (orchestrator picks temp dir).
2. Fake codex prints `CODEX_HOME` for assertion.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.CodexHome = ""
	return nil
}
```