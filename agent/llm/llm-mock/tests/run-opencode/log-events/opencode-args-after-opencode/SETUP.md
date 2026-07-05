# Scenario

**Feature**: flag parsing stops at `opencode`; opencode argv passed through unchanged

```
llm-mock run --log-events f.jsonl opencode run --model llm-mock/mock-model hi
orchestrator -> opencode CLI with argv "run --model llm-mock/mock-model hi"
```

## Steps

1. Install fake `opencode` on PATH (echoes `$*` as `OPENCODE_ARGV=...`).
2. Pass `--log-events` and opencode args `run --model llm-mock/mock-model hi`.
3. Do not use `LLM_MOCK_RUN_OPENCODE_COMMAND` (hook ignores argv).

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	if err := installFakeOpencodeEchoArgv(t, req); err != nil {
		return err
	}
	req.LogEventsPath = filepath.Join(t.TempDir(), "session.jsonl")
	req.OpencodeArgs = []string{"run", "--model", "llm-mock/mock-model", "hi"}
	return nil
}
```