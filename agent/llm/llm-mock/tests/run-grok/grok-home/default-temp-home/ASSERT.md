---
label: e2e
---

## Expected

- Exit code 0.
- `GROK_HOME` printed by fake grok is an existing directory (temp isolation).
- `config.toml` exists under that home with `models_base_url` pointing at mock.

## Side Effects

- `$GROK_HOME/config.toml` written by orchestrator.

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
	home := resp.GrokHomeUsed
	if home == "" {
		t.Fatalf("expected GROK_HOME in output, got:\n%s", resp.Stdout+resp.Stderr)
	}
	if st, statErr := os.Stat(home); statErr != nil || !st.IsDir() {
		t.Fatalf("GROK_HOME %q is not a directory: %v", home, statErr)
	}
	if resp.ConfigToml == "" {
		t.Fatalf("expected config.toml under %s", home)
	}
	if !strings.Contains(resp.ConfigToml, "models_base_url") {
		t.Fatalf("config.toml missing models_base_url:\n%s", resp.ConfigToml)
	}
	cfgPath := filepath.Join(home, "config.toml")
	if _, statErr := os.Stat(cfgPath); statErr != nil {
		t.Fatalf("config.toml not found at %s: %v", cfgPath, statErr)
	}
}
```
