---
label: e2e
---

## Expected

- Exit code 0.
- `GROK_HOME` used equals explicit `LLM_MOCK_GROK_HOME`.
- `config.toml` contains `models_base_url` and `mock-model` default.

## Side Effects

- `$LLM_MOCK_GROK_HOME/config.toml` created with mock endpoint URL.

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
	if resp.GrokHomeUsed != req.GrokHome {
		t.Fatalf("GrokHomeUsed = %q, want explicit %q", resp.GrokHomeUsed, req.GrokHome)
	}
	if resp.ConfigToml == "" {
		t.Fatalf("expected config.toml under %s", req.GrokHome)
	}
	lower := strings.ToLower(resp.ConfigToml)
	if !strings.Contains(lower, "models_base_url") {
		t.Fatalf("config.toml missing models_base_url:\n%s", resp.ConfigToml)
	}
	if !strings.Contains(resp.ConfigToml, "mock-model") {
		t.Fatalf("config.toml missing mock-model:\n%s", resp.ConfigToml)
	}
	if !strings.Contains(resp.ConfigToml, "127.0.0.1") {
		t.Fatalf("config.toml missing mock localhost URL:\n%s", resp.ConfigToml)
	}
	cfgPath := filepath.Join(req.GrokHome, "config.toml")
	if _, err := os.Stat(cfgPath); err != nil {
		t.Fatalf("config.toml not found at %s: %v", cfgPath, err)
	}
}
```
