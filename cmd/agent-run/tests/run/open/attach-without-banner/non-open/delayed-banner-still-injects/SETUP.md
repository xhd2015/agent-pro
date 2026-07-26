# Scenario

**Bug**: non-open delayed-banner path used to re-inject new-session prompt after
banner (double-submit with argv). New-session normal submit must **not** PTY
re-inject even when banner is delayed.

```
fake TUI sleep 0.3 → GROK_TTY_BANNER → timed read → probe
agent-run run --agent-runner grok-tty "hi"
  -> banner hard-wait succeeds (no "banner not detected")
  -> probe PROMPT_ARG=hi (argv path)
  -> probe STDIN_COUNT=0 (no re-inject after delayed banner)
```

## Preconditions

- New session + normal submit: prompt on argv; inject after banner is the bug.
- Delayed banner still within hard-wait window so readiness wait succeeds.
- Timed stdin read records inject absence without hanging when product skips inject.
- Resume / NoSubmit inject cases remain covered elsewhere (open NoSubmit;
  resume follow-up inject). This leaf is new-session non-open only.
- Sibling external suite: `cmd/agent-run/tests/grok-tty/run/waits-for-banner`
  (banner wait timing; may still assert inject until that suite is retargeted).

## Steps

1. Write delayed-banner argv/stdin probe fake TUI under temp dir.
2. Run with prompt `hi` (no `--open`).
3. Assert probe has argv prompt, no inject; no banner timeout error.

```go
import (
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	clearOpenInstantAttach(req)
	req.Prompt = "hi"
	probePath := filepath.Join(req.TempDir, "non-open-inject-probe.txt")
	setEnvKV(req, "DOCTEST_TUI_PROBE_PATH", probePath)
	// 0.3s delay mirrors historical delayed-banner fixture; timed read for no-inject.
	script := writeFakeTUIBannerArgvStdinProbe(t, req.TempDir, probePath, 0.3, 5, 2)
	setGrokTTYCommand(req, script)
	req.Args = []string{"run", "--agent-runner", "grok-tty", req.Prompt}
	// Banner delay is short; allow headless discovery/turn path to finish or fail
	// after inject policy — probe is the hard no-double-inject proof.
	req.ExecTimeout = 60 * time.Second
	return nil
}
```
