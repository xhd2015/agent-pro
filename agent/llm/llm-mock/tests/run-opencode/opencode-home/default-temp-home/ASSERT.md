---
label: e2e
---

## Expected

- Exit code 0.
- `OPENCODE_CONFIG_DIR` printed by fake opencode is an existing directory (temp isolation).
- Config dir path is not the user's default `~/.config/opencode`.
- Orchestrator-generated provider config references mock localhost (`127.0.0.1`) and `llm-mock` provider.

## Side Effects

- Child env receives `OPENCODE_CONFIG_CONTENT` with `@ai-sdk/openai-compatible` routing to mock.

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
	configDir := resp.OpencodeConfigDirUsed
	if configDir == "" {
		t.Fatalf("expected OPENCODE_CONFIG_DIR in output, got:\n%s", resp.Stdout+resp.Stderr)
	}
	if st, statErr := os.Stat(configDir); statErr != nil || !st.IsDir() {
		t.Fatalf("OPENCODE_CONFIG_DIR %q is not a directory: %v", configDir, statErr)
	}
	userConfig, _ := filepath.Abs(filepath.Join(os.Getenv("HOME"), ".config", "opencode"))
	if absDir, err := filepath.Abs(configDir); err == nil && absDir == userConfig {
		t.Fatalf("OPENCODE_CONFIG_DIR must not be user default %q", userConfig)
	}
	combined := resp.Stdout + resp.Stderr
	if !strings.Contains(combined, "127.0.0.1") && !strings.Contains(combined, "llm-mock") {
		t.Fatalf("expected mock provider announcement in output:\n%s", combined)
	}
}
```