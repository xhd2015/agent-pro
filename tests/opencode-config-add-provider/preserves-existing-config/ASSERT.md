---
label: e2e
---

## Expected

- Exit code 0.
- `provider.newprov` is present (newly added) with `npm` == `@ai-sdk/anthropic`.
- `provider.other` is still present and unchanged (`name` == `other`,
  `options.baseURL` == `https://other.example.com/v1`).
- The unrelated top-level key `permission` is still present (an empty map).

## Side Effects

- The pre-existing config file is rewritten (atomic temp+rename) with the new
  provider merged in; no other top-level keys removed.

## Errors

- None.

## Exit Code

- 0.

```go
import (
	"path/filepath"
	"strings"
	"testing"

	opencodecfg "github.com/xhd2015/agent-pro/agent/opencode/config"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	// Shared success invariants (inlined here because assertSuccessCommon is
	// only compiled into the success/* packages via success/SETUP.md).
	if resp.ExitCode != 0 {
		t.Fatalf("exit code = %d, stderr:\n%s", resp.ExitCode, resp.Stderr)
	}
	if !stderrContains(resp, "newprov") && !strings.Contains(resp.Stdout, "newprov") {
		t.Fatalf("provider id %q not mentioned in stdout/stderr:\nstdout:%s\nstderr:%s",
			"newprov", resp.Stdout, resp.Stderr)
	}
	mustExist(t, resp.ConfigPath)
	entry, _, e := readProviderEntry(resp.ConfigPath, "newprov")
	if e != nil {
		t.Fatalf("read provider[newprov]: %v", e)
	}
	if got, ok := entry["npm"].(string); !ok || got == "" {
		t.Fatalf("provider[newprov].npm missing or empty: %v", entry["npm"])
	}
	if got := entry["npm"]; got != "@ai-sdk/anthropic" {
		t.Fatalf("provider[newprov].npm = %v, want @ai-sdk/anthropic", got)
	}
	if _, ok := entry["options"].(map[string]interface{}); !ok {
		t.Fatalf("provider[newprov].options missing or not a map")
	}
	if _, ok := entry["models"].(map[string]interface{}); !ok {
		t.Fatalf("provider[newprov].models missing or not a map")
	}

	// Re-read the full config to check preservation.
	cfg, e := opencodecfg.ReadDir(filepath.Dir(resp.ConfigPath))
	if e != nil {
		t.Fatal(e)
	}
	// Unrelated top-level key preserved.
	if _, ok := cfg.Data["permission"]; !ok {
		t.Fatalf("top-level key 'permission' was removed: %v", cfg.Data)
	}
	prov, _ := cfg.Data["provider"].(map[string]interface{})
	// New provider added.
	if _, ok := prov["newprov"]; !ok {
		t.Fatalf("provider.newprov not added: %v", prov)
	}
	// Pre-existing provider preserved.
	other, ok := prov["other"].(map[string]interface{})
	if !ok {
		t.Fatalf("provider.other removed: %v", prov)
	}
	if got := other["name"]; got != "other" {
		t.Fatalf("provider.other.name = %v, want other", got)
	}
	opts, _ := other["options"].(map[string]interface{})
	if got := opts["baseURL"]; got != "https://other.example.com/v1" {
		t.Fatalf("provider.other.options.baseURL = %v, want unchanged", got)
	}
}
```
