# Scenario

**Feature**: default isolated GROK_HOME is a fresh temp directory

```
orchestrator -> create temp GROK_HOME + config.toml
fake grok -> echo GROK_HOME=$GROK_HOME
```

## Steps

1. Do not set `LLM_MOCK_GROK_HOME` (orchestrator picks temp dir).
2. Fake grok prints `GROK_HOME` for assertion.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.GrokHome = ""
	return nil
}
```
