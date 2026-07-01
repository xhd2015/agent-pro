## Expected

- Exit code 0.
- Config file exists at `$HOME/.config/opencode/opencode.json` (the global
  default target — no `--dir` given).
- `provider.myprov.npm` == `@ai-sdk/anthropic`.
- `provider.myprov.name` == `myprov` (defaults to id when `--name` omitted).
- `provider.myprov.options.baseURL` == `https://api.example.com/v1`.
- `provider.myprov.models` has exactly one key `sonnet` whose value is
  `{"name":"sonnet"}`.
- stdout mentions the provider id `myprov`.

## Side Effects

- The file `$HOME/.config/opencode/opencode.json` is created (dir made if
  missing) with 2-space JSON.

## Errors

- None.

## Exit Code

- 0.

```go
import (
	"path/filepath"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccessCommon(t, req, resp, "myprov")

	// Global default target path.
	wantPath := filepath.Join(resp.Home, ".config", "opencode", "opencode.json")
	if resp.ConfigPath != wantPath {
		t.Fatalf("config path = %s, want %s", resp.ConfigPath, wantPath)
	}

	entry, _, e := readProviderEntry(resp.ConfigPath, "myprov")
	if e != nil {
		t.Fatal(e)
	}
	if got := entry["npm"]; got != "@ai-sdk/anthropic" {
		t.Fatalf("npm = %v, want @ai-sdk/anthropic", got)
	}
	if got := entry["name"]; got != "myprov" {
		t.Fatalf("name = %v, want myprov", got)
	}
	opts, _ := entry["options"].(map[string]interface{})
	if got := opts["baseURL"]; got != "https://api.example.com/v1" {
		t.Fatalf("options.baseURL = %v, want https://api.example.com/v1", got)
	}
	models, _ := entry["models"].(map[string]interface{})
	if len(models) != 1 {
		t.Fatalf("models has %d entries, want 1: %v", len(models), models)
	}
	m, ok := models["sonnet"]
	if !ok {
		t.Fatalf("models.sonnet missing: %v", models)
	}
	mm, _ := m.(map[string]interface{})
	if mm["name"] != "sonnet" {
		t.Fatalf("models.sonnet.name = %v, want sonnet", mm["name"])
	}
}
```
