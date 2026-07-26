# Scenario

**Feature**: --dir <project> writes to <dir>/.opencode/opencode.json

```
# --dir given: project-local target instead of global
agent-pro opencode config add-provider --id localprov --base-url https://api.example.com/v1 --api-shape anthropic --model m1 --dir <tmp-project>
doctest <- <tmp-project>/.opencode/opencode.json : provider.localprov
```

## Preconditions

- `--dir` points at a temp project directory under the isolated `HOME`.
- Otherwise a minimal valid command (anthropic, one model).

## Steps

1. Create a temp project dir.
2. Set `req.Args` with `--dir <tmp-project>`.
3. Run and assert the config file landed at `<dir>/.opencode/opencode.json`,
   NOT under `$HOME/.config/opencode`.

## Context

- This leaf isolates the **target resolution** factor (global vs project-local);
  other success factors (shape, name, models) are covered by sibling leaves.

```go
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ProjectDir = filepath.Join(req.Home, "my-project")
	if err := os.MkdirAll(req.ProjectDir, 0o755); err != nil {
		return err
	}
	req.Args = []string{
		"opencode", "config", "add-provider",
		"--id", "localprov",
		"--base-url", "https://api.example.com/v1",
		"--api-shape", "anthropic",
		"--model", "m1",
		"--dir", req.ProjectDir,
	}
	return nil
}
```
