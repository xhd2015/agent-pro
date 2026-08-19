# Codex `DetectScreenStatus` on finished live chrome

Crime scene `idle-probe-10s-verify-v10`: after Codex answers `pong` and sits
at a sendable `›` prompt, `tty status` reports `screen_status=banner` so
`--exit-on-idle` never arms (`SampleIsIdle` requires `screen==idle`).
`CheckWritable` is already idle on the same chrome.

L2 in-process: `codex-tty` `DetectScreenStatus` must return **idle** on
finished live Codex chrome (no `CODEX_TTY_BANNER` stub). Isolated nested
root so sibling `codex-writable` leaves that still allow `banner|idle`
stay sealed.

**Out of scope:** L3 `agent-run` + `llm-mock-run-codex` e2e (mock-model
footer is `default ·`, so `input_box` stays occupied and an exits leaf
would stay RED after the screen-only alignment); L2 help/parse
(`run-exit-on-idle/`); grok-tty (`run-exit-on-idle-grok-tty/`).

# DSN (Domain Specific Notion)

Finished Codex main-chat chrome with a live `›`/`»` prompt and no stub
banner marker is an **idle** screen for `tty status` and the serve idle
watchdog.

**Participants**

- **Snapshot text** — tty-watch / `tty snapshot` printable scrollback
  (crime-scene host v10, mock-model scene, locked 0.147 empty-glued).
- **`codex-tty` `DetectScreenStatus`** — `agenttty.Get("codex-tty")`;
  desired token `idle` (not `banner` / `unknown`).
- **`CheckWritable`** — already `ready` / `idle` on these finished shapes
  (regression: screen must catch up).
- **`DetectInputBox`** — occupancy is independent; host / empty-glued
  are `empty` via ` medium · `; mock-model `default ·` may stay occupied.

**Behaviors**

- Host v10 finished composer (`›` + ` medium · `, MCP incomplete warning,
  no `CODEX_TTY_BANNER`) → `screen=idle`.
- Scene mock-model finished `›` (no stub marker) → `screen=idle`.
- Locked 0.147 empty-glued last line → `screen=idle`.
- Writable stays idle on the host / empty-glued shapes.

## Version

0.0.2

## Decision Tree

```
pkgs/agenttty/tests/codex-screen-status/
├── DOCTEST.md
├── SETUP.md
├── testdata/
└── finished-prompt/                         # post-turn live ›, no stub banner
    ├── host-v10/                            # crime-scene idle-probe-10s-verify-v10
    ├── mock-model-default/                  # llm-mock-run-codex scene chrome
    └── empty-glued-0.147/                   # locked 0.147 empty composer
```

Parameter ranking (most → least significant):

1. **Chrome class** — finished live prompt (no `CODEX_TTY_BANNER`)
2. **Capture source** — host v10 vs mock-model vs locked 0.147
3. **Occupancy** — empty (` medium · `) vs occupied (`default ·`); screen still idle

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `finished-prompt/host-v10` | Host crime-scene snapshot → `screen=idle`, box empty, writable idle |
| 2 | `finished-prompt/mock-model-default` | Scene mock-model `›` → `screen=idle` |
| 3 | `finished-prompt/empty-glued-0.147` | Locked 0.147 empty-glued line → `screen=idle` |

## How to Run

```sh
# From the agent-pro module root:
doctest vet ./pkgs/agenttty/tests/codex-screen-status
doctest test ./pkgs/agenttty/tests/codex-screen-status

doctest test -v ./pkgs/agenttty/tests/codex-screen-status/finished-prompt/host-v10
```

Unlabeled L2 (in-process `DetectScreenStatus`). Expect **RED** until
`detectCodexScreenStatus` returns `idle` on live `›` chrome.

```go
import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/agenttty"
	"github.com/xhd2015/doctest/session"
)

// Request injects snapshot text. Fixture basename under testdata/ wins when set.
type Request struct {
	Scrollback  string
	Fixture     string
	TestdataDir string
}

// Response is screen + occupancy + writable on the same snapshot.
type Response struct {
	Screen        string
	InputBox      string
	WritableReady bool
	WritableState string
	Text          string
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	text, err := resolveScrollback(d, req)
	if err != nil {
		return nil, err
	}
	provider, ok := agenttty.Get("codex-tty")
	if !ok {
		return nil, fmt.Errorf("codex-tty provider not registered")
	}
	if provider.DetectScreenStatus == nil {
		return nil, fmt.Errorf("codex-tty DetectScreenStatus is nil")
	}
	resp := &Response{
		Screen:   strings.TrimSpace(provider.DetectScreenStatus([]byte(text))),
		InputBox: strings.TrimSpace(fmt.Sprint(agenttty.DetectInputBox(text))),
		Text:     text,
	}
	if provider.CheckWritable != nil {
		w := provider.CheckWritable([]byte(text))
		resp.WritableReady = w.Ready
		resp.WritableState = w.State
	}
	return resp, nil
}

func resolveScrollback(d *session.Doctest, req *Request) (string, error) {
	if req.Fixture != "" {
		dir := req.TestdataDir
		if dir == "" {
			dir = filepath.Join(d.DOCTEST_ROOT, "testdata")
		}
		raw, err := os.ReadFile(filepath.Join(dir, req.Fixture))
		if err != nil {
			return "", fmt.Errorf("read fixture %s: %w", req.Fixture, err)
		}
		return string(raw), nil
	}
	return req.Scrollback, nil
}
```
