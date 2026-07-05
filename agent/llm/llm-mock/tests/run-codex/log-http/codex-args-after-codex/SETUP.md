# Scenario

**Feature**: flag parsing stops at `codex`; codex argv passed through with `--log-http` set

```
llm-mock run --log-http f.jsonl codex exec -m mock-model "hi"
orchestrator -> codex CLI with argv "exec -m mock-model hi"
```

## Steps

1. Install fake `codex` on PATH (echoes `$*` as `CODEX_ARGV=...`).
2. Pass `--log-http` and codex args `exec -m mock-model "hi"`.
3. Do not use `LLM_MOCK_RUN_CODEX_COMMAND` (hook ignores argv).

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	if err := installFakeCodexEchoArgv(t, req); err != nil {
		return err
	}
	req.LogHTTPPath = filepath.Join(t.TempDir(), "http.jsonl")
	req.CodexArgs = []string{"exec", "-m", "mock-model", "hi"}
	return nil
}
```