# Scenario

**Feature**: agent-pro grok models list from synthetic GROK home

```
# harness builds synthetic Grok home with config.toml / models_cache.json
test harness -> grok/models.List -> Catalog

# format as human text or JSON for CLI output
Catalog -> FormatText / FormatJSON -> terminal text
```

## Preconditions

- Package `agent/grok/models` exposes List, FormatText, and FormatJSON.
- Tests never read the real user `~/.grok` directory.

## Steps

1. Root Setup allocates `req.GrokHome` as `{temp}/.grok`.
2. Leaf Setup writes config/cache fixtures as needed.
3. Run calls List + FormatText/FormatJSON based on `req.Format`.

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	grokmodels "github.com/xhd2015/agent-pro/agent/grok/models"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.GrokHome = filepath.Join(t.TempDir(), ".grok")
	return nil
}

func writeGrokConfig(t *testing.T, home, contents string) {
	t.Helper()
	if err := os.MkdirAll(home, 0o755); err != nil {
		t.Fatalf("mkdir grok home: %v", err)
	}
	path := filepath.Join(home, grokmodels.DefaultConfigFile)
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
}

func writeGrokCache(t *testing.T, home, contents string) {
	t.Helper()
	if err := os.MkdirAll(home, 0o755); err != nil {
		t.Fatalf("mkdir grok home: %v", err)
	}
	path := filepath.Join(home, grokmodels.ModelsCacheFile)
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
```
