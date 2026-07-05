# Scenario

**Feature**: default isolated `OPENCODE_CONFIG_DIR` is a fresh temp directory

```
orchestrator -> create temp OPENCODE_CONFIG_DIR + OPENCODE_CONFIG_CONTENT
fake opencode -> echo OPENCODE_CONFIG_DIR=$OPENCODE_CONFIG_DIR
```

## Steps

1. Do not set `LLM_MOCK_OPENCODE_CONFIG_DIR` or `LLM_MOCK_OPENCODE_HOME` (orchestrator picks temp dirs).
2. Fake opencode prints `OPENCODE_CONFIG_DIR` for assertion.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.OpencodeConfigDir = ""
	req.OpencodeHome = ""
	return nil
}
```