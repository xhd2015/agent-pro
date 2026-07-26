---
label: e2e
---

## Expected

- Exit code 0.
- `CODEX_HOME` used equals explicit `LLM_MOCK_CODEX_HOME`.
- `config.toml` contains `model_providers.llm-mock`, `base_url`, `mock-model`, and `wire_api = "responses"`.

## Side Effects

- `$LLM_MOCK_CODEX_HOME/config.toml` created with mock endpoint URL.

## Exit Code

0

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	if resp.CodexHomeUsed != req.CodexHome {
		t.Fatalf("CodexHomeUsed = %q, want explicit %q", resp.CodexHomeUsed, req.CodexHome)
	}
	if resp.ConfigToml == "" {
		t.Fatalf("expected config.toml under %s", req.CodexHome)
	}
	lower := strings.ToLower(resp.ConfigToml)
	if !strings.Contains(lower, "base_url") {
		t.Fatalf("config.toml missing base_url:\n%s", resp.ConfigToml)
	}
	if !strings.Contains(resp.ConfigToml, "mock-model") {
		t.Fatalf("config.toml missing mock-model:\n%s", resp.ConfigToml)
	}
	if !strings.Contains(resp.ConfigToml, "llm-mock") {
		t.Fatalf("config.toml missing llm-mock provider:\n%s", resp.ConfigToml)
	}
	if !strings.Contains(resp.ConfigToml, "127.0.0.1") {
		t.Fatalf("config.toml missing mock localhost URL:\n%s", resp.ConfigToml)
	}
	cfgPath := filepath.Join(req.CodexHome, "config.toml")
	if _, err := os.Stat(cfgPath); err != nil {
		t.Fatalf("config.toml not found at %s: %v", cfgPath, err)
	}
}
```