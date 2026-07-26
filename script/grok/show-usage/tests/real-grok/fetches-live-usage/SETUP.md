# Scenario

**Feature**: live grok returns parseable usage line patterns

```
# real grok TUI -> /usage show -> Weekly limit: N% + Next reset: <date>
grok-show-usage -> real grok -> pattern assertions
```

## Preconditions

- Real `grok` on PATH (inherited from `real-grok/SETUP.md`).
- No fake TUI hook.

## Steps

1. Run CLI against production grok argv.
2. Assert exit 0; stdout matches usage line patterns; stderr empty.

## Context

- Optional leaf for manual/CI-with-grok verification; skipped without `--label real-grok`.

```go
import (
	"os/exec"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if _, err := exec.LookPath("grok"); err != nil {
		t.Skip("grok not found in PATH")
	}
	req.SkipFakeCommand = true
	req.TimeoutSeconds = "60"
	return nil
}
```