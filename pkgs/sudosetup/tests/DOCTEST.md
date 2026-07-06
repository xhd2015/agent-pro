# sudosetup Tests

Doc-style tests for `github.com/xhd2015/agent-pro/pkgs/sudosetup` — reusable
NOPASSWD sudoers setup used by the sudo-nopasswd POC and `ws-proxy vpn`.

# DSN (Domain Specific Notion)

`sudosetup.Manager` coordinates persistent sudoers installation for one privileged
command (optionally with an args pattern). Callers supply `Config` (cache dir name,
sudoers drop-in name, target username) and `Rule` (command path + args pattern).

**Participants**

- **Manager** — `Detect`, `EnsureInstalled`, `Remove`, `IsInstalled`,
  `RenderSudoersLine`, path helpers.
- **FS** — injectable file layer; tests use an in-memory map (no real `/etc` writes).
- **Runner** — injectable subprocess layer; tests record `sudo`/`visudo`/`install`
  invocations without executing them.
- **Sudoers drop-in** — `/etc/sudoers.d/<SudoersName>` with one NOPASSWD line.
- **Install manifest** — `~/.cache/<CacheDirName>/sudo-setup-manifest.json`
  recording username, command, args_pattern for persistent detection.

**Behaviors**

- `IsInstalled` / persistent detection: drop-in exists **and** manifest matches
  current rule/user — never uses `sudo -n` (avoids timestamp-cache false positives).
- `Detect` adds live probes: `sudo -n true` (cache warm) and `sudo -n <command> …`
  (non-interactive runnable).
- `EnsureInstalled` skips when already installed; otherwise requires interactive stdin
  (TTY) before `visudo -cf` → `install -m 0440` → write manifest.
- `Remove` requires interactive stdin when deleting an existing drop-in; always runs
  `sudo -k` when removal proceeds (noop path only flushes cache).

Tests inject fake `FS` + `Runner` only — no real sudo, visudo, or `/etc` writes.

## Version

0.0.2

## Decision Tree

```
pkgs/sudosetup/tests/
├── DOCTEST.md
├── SETUP.md
├── detect/                                    (grouping: Detect / IsInstalled)
│   ├── no-drop-in-no-manifest/               (LEAF) neither file → not installed
│   ├── drop-in-without-manifest/             (LEAF) orphaned drop-in
│   ├── manifest-mismatch-user/                (LEAF) manifest user ≠ current
│   ├── manifest-mismatch-command/            (LEAF) manifest command ≠ rule
│   ├── fully-installed/                      (LEAF) drop-in + manifest match
│   ├── cache-warm-without-rule/              (LEAF) sudo -n true OK, not installed
│   └── installed-and-runnable/               (LEAF) installed + sudo -n cmd OK
├── install/                                   (grouping: EnsureInstalled)
│   ├── skips-when-installed/                 (LEAF) no visudo/install when installed
│   ├── writes-sudoers-line/                  (LEAF) visudo -cf + install drop-in
│   ├── writes-manifest/                      (LEAF) manifest JSON matches rule
│   ├── visudo-failure/                       (LEAF) visudo error, no manifest
│   ├── non-tty-requires-terminal/            (LEAF) !stdin TTY → error before sudo
│   └── non-tty-skips-when-installed/         (LEAF) installed + !stdin TTY → noop
├── remove/                                    (grouping: Remove)
│   ├── removes-drop-in-and-manifest/         (LEAF) rm drop-in, delete manifest, sudo -k
│   ├── noop-when-missing/                    (LEAF) only sudo -k when nothing installed
│   └── non-tty-requires-terminal/            (LEAF) drop-in exists + !stdin TTY → error
└── render/                                    (grouping: RenderSudoersLine)
     ├── bare-command/                        (LEAF) hello.sh line without args
     └── command-with-wildcard-args/          (LEAF) sing-box `run -c *` pattern
```

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `detect/no-drop-in-no-manifest` | Neither drop-in nor manifest → not installed |
| 2 | `detect/drop-in-without-manifest` | Drop-in without manifest → orphaned detail |
| 3 | `detect/manifest-mismatch-user` | Manifest username mismatch |
| 4 | `detect/manifest-mismatch-command` | Manifest command mismatch |
| 5 | `detect/fully-installed` | Drop-in + manifest match → installed |
| 6 | `detect/cache-warm-without-rule` | Cache warm but rule not installed |
| 7 | `detect/installed-and-runnable` | Installed + non-interactive command probe OK |
| 8 | `install/skips-when-installed` | EnsureInstalled noop when already installed |
| 9 | `install/writes-sudoers-line` | Renders line, runs visudo + install |
| 10 | `install/writes-manifest` | Writes manifest JSON after install |
| 11 | `install/visudo-failure` | visudo error aborts, no manifest written |
| 12 | `install/non-tty-requires-terminal` | Non-TTY stdin errors before visudo/install |
| 13 | `install/non-tty-skips-when-installed` | Already installed skips TTY requirement |
| 14 | `remove/removes-drop-in-and-manifest` | Removes drop-in, manifest, runs sudo -k |
| 15 | `remove/noop-when-missing` | Missing install → only sudo -k |
| 16 | `remove/non-tty-requires-terminal` | Non-TTY stdin errors before sudo rm |
| 17 | `render/bare-command` | NOPASSWD line for bare script command |
| 18 | `render/command-with-wildcard-args` | NOPASSWD line includes args pattern |

## How to Run

```sh
doctest vet ./pkgs/sudosetup/tests
doctest test ./pkgs/sudosetup/tests
doctest test ./pkgs/sudosetup/tests/detect/fully-installed
```

```go
import (
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/sudosetup/sudosetuptest"
)

type Request = sudosetuptest.Request
type Response = sudosetuptest.Response
type ManifestSeed = sudosetuptest.ManifestSeed
type RunnerCall = sudosetuptest.RunnerCall

func Run(t *testing.T, req *Request) (*Response, error) {
	return sudosetuptest.Run(t, req)
}
```