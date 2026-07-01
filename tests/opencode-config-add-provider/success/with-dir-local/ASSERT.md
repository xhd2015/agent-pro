## Expected

- Exit code 0.
- Config file exists at `<project-dir>/.opencode/opencode.json`.
- No config file is created under `$HOME/.config/opencode/` (the global path
  is not used when `--dir` is given).
- `provider.localprov` is present with `npm` == `@ai-sdk/anthropic`.

## Side Effects

- `<project-dir>/.opencode/opencode.json` created (dir made if missing).

## Errors

- None.

## Exit Code

- 0.

```go
import (
	"os"
	"path/filepath"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccessCommon(t, req, resp, "localprov")

	wantPath := filepath.Join(req.ProjectDir, ".opencode", "opencode.json")
	if resp.ConfigPath != wantPath {
		t.Fatalf("config path = %s, want %s", resp.ConfigPath, wantPath)
	}
	// Global path must NOT have been written.
	globalPath := filepath.Join(resp.Home, ".config", "opencode", "opencode.json")
	if _, err := os.Stat(globalPath); err == nil {
		t.Fatalf("global config unexpectedly written at %s", globalPath)
	}
}
```
