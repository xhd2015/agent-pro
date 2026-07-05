## Expected

- Exit code 0.
- `CODEX_HOME` printed by fake codex is an existing directory (temp isolation).
- `config.toml` exists under that home with `base_url` pointing at mock and `wire_api = "responses"`.

## Side Effects

- `$CODEX_HOME/config.toml` written by orchestrator with `approval_policy = "never"` and `shell_tool = false`.

## Exit Code

0

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	home := resp.CodexHomeUsed
	if home == "" {
		t.Fatalf("expected CODEX_HOME in output, got:\n%s", resp.Stdout+resp.Stderr)
	}
	if st, statErr := os.Stat(home); statErr != nil || !st.IsDir() {
		t.Fatalf("CODEX_HOME %q is not a directory: %v", home, statErr)
	}
	if resp.ConfigToml == "" {
		t.Fatalf("expected config.toml under %s", home)
	}
	if !strings.Contains(resp.ConfigToml, "base_url") {
		t.Fatalf("config.toml missing base_url:\n%s", resp.ConfigToml)
	}
	if !strings.Contains(resp.ConfigToml, "wire_api") || !strings.Contains(resp.ConfigToml, "responses") {
		t.Fatalf("config.toml missing wire_api = responses:\n%s", resp.ConfigToml)
	}
	if !strings.Contains(resp.ConfigToml, "127.0.0.1") {
		t.Fatalf("config.toml missing mock localhost URL:\n%s", resp.ConfigToml)
	}
	cfgPath := filepath.Join(home, "config.toml")
	if _, statErr := os.Stat(cfgPath); statErr != nil {
		t.Fatalf("config.toml not found at %s: %v", cfgPath, statErr)
	}
}
```