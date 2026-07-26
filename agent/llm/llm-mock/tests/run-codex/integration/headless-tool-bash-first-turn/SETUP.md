# Scenario

**Feature**: first codex turn with `--mock-events-preset=tool-bash` must run mocked bash and show output

Reproduces user report: interactive `llm-mock run --mock-events-preset=tool-bash codex` first prompt
(`hasdfas`) shows no assistant/tool feedback; only later turns show genStream text.

Headless proxy: `codex exec` first user message should execute preset bash `echo preset-bash`
and surface `preset-bash` in output.

## Preconditions

- Real `codex` on PATH.
- Preset command: `echo preset-bash`.

## Steps

1. Run `llm-mock run --mock-events-preset=tool-bash codex exec ... "hasdfas"`.
2. Assert stdout/stderr contains bash echo output on first turn.

```go
import (
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SkipRealCodex = true
	req.FakeCodexCmd = ""
	req.ConfigEnv = ""
	req.MockEventsPreset = "tool-bash"
	req.ExecTimeout = 60 * time.Second
	req.CodexArgs = []string{
		"exec",
		"--skip-git-repo-check",
		"--dangerously-bypass-approvals-and-sandbox",
		"-m", "mock-model",
		"hasdfas",
	}
	return nil
}
```