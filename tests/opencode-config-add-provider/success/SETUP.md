# Scenario

**Feature**: add-provider success paths write the v1 provider entry and exit 0

```
# valid flags -> provider entry written, confirmation printed, exit 0
agent-pro opencode config add-provider --id <id> --base-url <url> --api-shape <shape> --model <m>
doctest <- config file: provider[id] = { npm, name, options.baseURL, models }
```

## Preconditions

- All mandatory flags are supplied: `--id`, `--base-url`, `--api-shape` (one of
  `anthropic`/`openai`), and at least one `--model`.

## Steps

1. Compose a full `opencode config add-provider` arg list in `req.Args`.
2. Run the binary (root `Run`) with isolated `HOME`.
3. Assert exit code 0, a confirmation line mentioning the provider id, and the
   parsed config file content (provider entry fields).

## Context

- Success leaves assert on the **parsed** config file via `readProviderEntry`,
  reusing `opencodecfg.ReadDir` so JSONC is tolerated.
- The default global target is used unless a leaf sets `--dir`.

```go
import (
	"strings"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	// Leaves under success/ set req.Args in their own Setup. Verify the shared
	// harness is ready for this subtree: the binary must have been built by
	// the root Setup.
	if req.Bin == "" {
		t.Fatalf("success setup: agent-pro binary not built (root Setup skipped?)")
	}
	return nil
}

// assertSuccessCommon checks the invariants every success leaf shares:
// exit 0, stderr empty (or near-empty), stdout mentions the provider id, and
// the config file exists with the expected provider id present.
func assertSuccessCommon(t *testing.T, req *Request, resp *Response, id string) {
	t.Helper()
	if resp.ExitCode != 0 {
		t.Fatalf("exit code = %d, stderr:\n%s", resp.ExitCode, resp.Stderr)
	}
	if !stderrContains(resp, id) && !strings.Contains(resp.Stdout, id) {
		t.Fatalf("provider id %q not mentioned in stdout/stderr:\nstdout:%s\nstderr:%s",
			id, resp.Stdout, resp.Stderr)
	}
	mustExist(t, resp.ConfigPath)
	entry, _, err := readProviderEntry(resp.ConfigPath, id)
	if err != nil {
		t.Fatalf("read provider[%s]: %v", id, err)
	}
	if got, ok := entry["npm"].(string); !ok || got == "" {
		t.Fatalf("provider[%s].npm missing or empty: %v", id, entry["npm"])
	}
	if _, ok := entry["options"].(map[string]interface{}); !ok {
		t.Fatalf("provider[%s].options missing or not a map", id)
	}
	if _, ok := entry["models"].(map[string]interface{}); !ok {
		t.Fatalf("provider[%s].models missing or not a map", id)
	}
}
```
