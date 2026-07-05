# Scenario

**Feature**: `llm-mock-run-grok` forwards argv flags to grok

```
llm-mock-run-grok --always-approve
  -> RunGrok(grokArgs=["--always-approve"])
  -> fake grok on PATH echoes GROK_ARGV=--always-approve
```

## Steps

1. Install fake `grok` on PATH that echoes `$*` as `GROK_ARGV=...`.
2. Run shortcut with `--always-approve` only (no `LLM_MOCK_RUN_GROK_COMMAND` — hook ignores argv).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if err := installFakeGrokEchoArgv(t, req); err != nil {
		return err
	}
	req.UseShortcut = true
	req.OmitCLIRunFlags = true
	req.GrokArgs = []string{"--always-approve"}
	req.ExpectedExit = 0
	return nil
}
```