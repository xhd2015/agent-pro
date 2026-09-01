# Scenario

**Feature**: agent-pro codex models list from synthetic Codex home

```
# harness builds synthetic Codex home with config.toml / models_cache.json
test harness -> codex/models.List -> Catalog

# format as human text or JSON for CLI output
Catalog -> FormatText / FormatJSON -> terminal text
```

## Preconditions

- Package `agent/codex/models` exposes List, FormatText, and FormatJSON.
- Tests never read the real user `~/.codex` directory.

## Steps

1. Root Setup allocates `req.CodexHome` as `{temp}/.codex`.
2. Leaf Setup writes config/cache fixtures as needed.
3. Run calls List + FormatText/FormatJSON based on `req.Format`.

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	codexmodels "github.com/xhd2015/agent-pro/agent/codex/models"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.CodexHome = filepath.Join(t.TempDir(), ".codex")
	return nil
}

func writeCodexConfig(t *testing.T, home, contents string) {
	t.Helper()
	if err := os.MkdirAll(home, 0o755); err != nil {
		t.Fatalf("mkdir codex home: %v", err)
	}
	path := filepath.Join(home, codexmodels.DefaultConfigFile)
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
}

func writeCodexCache(t *testing.T, home, contents string) {
	t.Helper()
	if err := os.MkdirAll(home, 0o755); err != nil {
		t.Fatalf("mkdir codex home: %v", err)
	}
	path := filepath.Join(home, codexmodels.ModelsCacheFile)
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write cache: %v", err)
	}
}

func assertSuccess(t *testing.T, resp *Response) {
	t.Helper()
	if resp.Err != nil {
		t.Fatalf("operation failed: %v", resp.Err)
	}
}

func assertContains(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("missing %q in:\n%s", want, got)
	}
}

func assertNotContains(t *testing.T, got, want string) {
	t.Helper()
	if strings.Contains(got, want) {
		t.Fatalf("unexpected %q in:\n%s", want, got)
	}
}
```
