# tty-watch CLI Doctests

End-to-end tests for the standalone `tty-watch` CLI: `run` subcommand embeds ptywrap
(writer attach), `list`, `watch`, `snapshot`, and `kill`.

# DSN (Domain Specific Notion)

**Participants**

- **tty-watch subprocess** — built from `./script/tty-watch`; subcommands `list`,
  `run`, `list`, `watch`, `snapshot`, `kill`.
- **Embedded ptywrap server** — HTTP/WS listener on `127.0.0.1:0` inside the
  owning `tty-watch` process for each session.
- **Registry** — `$TTY_WATCH_HOME/registry/` JSON files (`session-N.json`) with
  flock-based id reservation (groktty-style).
- **PTY child** — command argv executed inside the embedded terminal session.
- **Test harness** — isolated `TTY_WATCH_HOME` per test, PTY helpers for attach,
  Ctrl-C (`\x03`), Ctrl-] detach (`\x1d`), registry probes.

**Behaviors**

- Default run reserves `session-N`, starts ptywrap, writes registry, attaches as
  interactive writer; **host stays silent** (no session-id on stdout/stderr).
- Ctrl-C forwards interrupt to PTY child; Ctrl-] detaches client only (session +
  registry survive).
- Command exit while attached removes registry entry and exits host.
- `list` scans registry, TCP-probes `listen_addr`, prints id + command + uptime;
  prunes unreachable entries.
- `watch` attaches as observer: raw PTY bytes to stdout, no stdin forwarding.
- `snapshot` attaches as snapshot mode; prints scrollback with ANSI/C0 sanitized.
- `kill` DELETEs session via ptywrap API, SIGTERM owner PID, removes registry;
  unreachable server idempotently prunes registry (exit 0).

## Version

0.0.2

## Decision Tree

```
[tty-watch CLI]
 |
 +-- run/
 |    |
 |    +-- registers-session/          (LEAF)  default run writes registry entry (RED)
 |    +-- attaches-pty-output/        (LEAF)  writer attach shows child output (RED)
 |    +-- host-stays-silent/          (LEAF)  no session-id on host stdout/stderr (RED)
 |    +-- ctrl-c-reaches-child/      (LEAF)  \x03 interrupts PTY child (RED)
 |    +-- detach-survives/           (LEAF)  \x1d detach; session in list (RED)
 |    +-- exits-cleans-registry/      (LEAF)  attached exit prunes registry (RED)
 |    +-- echo-prints-and-exits/      (LEAF)  run echo yes prints and exits promptly (RED)
 |    +-- bash-c-prints-and-exits/    (LEAF)  run bash -c echo yes prints and exits (RED)
 |    +-- cr-overwrite-preserved/     (LEAF)  \\r cursor positioning not stripped (RED)
 |    +-- interactive-bash-layout/    (LEAF)  interactive bash errors not smeared right (RED)
 |    +-- echo-clean-two-lines/      (LEAF)  echo yes: no leading blank line; yes + [Terminal exited] (RED)
 |    +-- bash-c-clean-two-lines/    (LEAF)  bash -c echo yes: no leading blank line smear (RED)
 |    +-- exit-marker-column-zero/  (LEAF)  no leading blank line; [Terminal exited] column 0 (RED)
 |
 +-- unit/
 |    |
 |    +-- screen-snapshot-exit-marker/ (LEAF)  snapshot text no leading \\n; exit marker column 0 (RED)
 |    +-- observer-detach-stdin-before-cleanup/ (LEAF) detach restores stdin before stdout cleanup (RED)
 |    +-- observer-detach-kitty-pop-cleanup/ (LEAF) detach pops grok kitty keyboard stack with \\x1b[<u (RED)
 |
 +-- list/
 |    |
 |    +-- shows-command-uptime/       (LEAF)  list prints id, command, uptime (RED)
 |    +-- empty-when-none/            (LEAF)  empty registry yields empty list (RED)
 |
 +-- watch/
 |    |
 |    +-- streams-output/             (LEAF)  observer streams raw PTY bytes (RED)
 |    +-- readonly-no-input/          (LEAF)  stdin not forwarded to session (RED)
 |    +-- readonly-tty-no-local-echo/ (LEAF)  TTY: no local echo of typed/mouse input (RED)
 |    +-- ctrl-c-detaches/            (LEAF)  Ctrl-C detaches watch; session survives (RED)
 |    +-- ctrl-c-detaches-sigint/     (LEAF)  SIGINT detaches watch when stdin not raw (RED)
 |    +-- ctrl-c-detaches-nonraw-stdin/ (LEAF) TTY stdout + pipe stdin: SIGINT detaches (RED)
 |    +-- ctrl-c-detaches-grok-modes-kitty-ctrl-c/ (LEAF) grok mode preamble + kitty Ctrl-C detaches (RED)
 |    +-- ctrl-c-detaches-real-grok-kitty-ctrl-c/ (LEAF) real grok alt-screen + kitty Ctrl-C detaches (RED)
 |    +-- ctrl-c-detaches-grok-modes-tty-cleanup/ (LEAF) grok modes + iTerm kitty Ctrl-C restores observer TTY (RED)
 |    +-- ctrl-c-detaches-grok-modes-post-detach-kitty-garbage/ (LEAF) post-detach kitty input must not garble shell (RED)
 |    +-- grok-like-prompt-clean/      (LEAF)  grok-like TUI: no CSI/C0 in watch output (RED)
 |    +-- grok-tui-tty-raw-mirror/     (LEAF)  watch TTY mirrors alt-screen via raw ESC (RED)
 |    +-- grok-tui-single-screen-state/(LEAF)  watch pipe: one screen, no stacked redraws (RED)
 |    +-- grok-tui-tty-no-mixed-snapshot-sgr/ (LEAF)  TTY: no plain snapshot + raw SGR mix (RED)
 |
 +-- snapshot/
 |    |
 |    +-- prints-sanitized/           (LEAF)  snapshot strips ANSI/C0 controls (RED)
 |    +-- unknown-session/            (LEAF)  missing session errors (RED)
 |
 +-- kill/
 |    |
 |    +-- terminates-detached/        (LEAF)  kill ends detached session (RED)
 |    +-- unknown-session/            (LEAF)  kill missing session errors (RED)
 |    +-- prunes-unreachable/         (LEAF)  stale registry pruned, exit 0 (RED)
 |
 +-- errors/
      |
      +-- unknown-subcommand/         (LEAF)  bogus subcommand errors (RED)
```

Parameter ranking (most → least significant):

1. **Operation mode** — default run vs list vs watch vs snapshot vs kill vs errors
2. **Session lifecycle** — attach, detach, exit, kill, stale prune
3. **I/O contract** — silent host, raw watch, sanitized snapshot
4. **Session lookup** — known vs unknown id

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `run/registers-session` | Default run creates `session-N.json` in registry (RED) |
| 2 | `run/attaches-pty-output` | Writer attach shows `RUN_OK` from child command (RED) |
| 3 | `run/host-stays-silent` | Host prints no `session-` id during attach/detach (RED) |
| 4 | `run/ctrl-c-reaches-child` | Ctrl-C (`\x03`) triggers child interrupt handler (RED) |
| 5 | `run/detach-survives` | Ctrl-] (`\x1d`) detaches; session remains in list (RED) |
| 6 | `run/exits-cleans-registry` | Attached `true` exit removes registry entry (RED) |
| 7 | `run/echo-prints-and-exits` | `run echo yes` prints `yes` and exits 0 without hang (RED) |
| 8 | `run/bash-c-prints-and-exits` | `run bash -c 'echo yes'` prints and exits promptly (RED) |
| 9 | `run/cr-overwrite-preserved` | `\r` overwrite preserved; no `MARKER_AMARKER_B` smear (RED) |
| 20 | `run/interactive-bash-layout` | Interactive bash profile errors stay left; no smeared `bash:` (RED) |
| 21 | `run/echo-clean-two-lines` | `run echo yes` has no leading blank line; only `yes` + `[Terminal exited]` (RED) |
| 22 | `run/bash-c-clean-two-lines` | `run bash -c 'echo yes'` has no leading blank line before `yes` (RED) |
| 23 | `run/exit-marker-column-zero` | No leading blank line; `[Terminal exited]` at column 0; trailing newline (RED) |
| 24 | `unit/screen-snapshot-exit-marker` | Snapshot text must not start with `\n`; exit marker at column 0 (RED) |
| 37 | `unit/observer-detach-stdin-before-cleanup` | Detach restores stdin before stdout TTY cleanup (RED) |
| 39 | `unit/observer-detach-kitty-pop-cleanup` | Detach pops grok kitty keyboard stack with `\x1b[<u` (RED) |
| 10 | `list/shows-command-uptime` | `list` shows session id, command argv, uptime (RED) |
| 11 | `list/empty-when-none` | Empty registry prints no sessions (RED) |
| 12 | `watch/streams-output` | `watch` streams raw `WATCH_MARKER` output (RED) |
| 13 | `watch/readonly-no-input` | `watch` ignores stdin; no echo of probe input (RED) |
| 29 | `watch/readonly-tty-no-local-echo` | `watch` TTY silently drops typed/mouse input (RED) |
| 30 | `watch/ctrl-c-detaches` | `watch` Ctrl-C detaches observer; session survives (RED) |
| 31 | `watch/ctrl-c-detaches-sigint` | `watch` SIGINT detaches when Ctrl-C is not delivered as `\x03` (RED) |
| 32 | `watch/ctrl-c-detaches-nonraw-stdin` | `watch` TTY stdout + non-TTY stdin: SIGINT detaches (RED) |
| 33 | `watch/ctrl-c-detaches-grok-modes-kitty-ctrl-c` | `watch` after grok terminal modes: kitty Ctrl-C detaches (RED) |
| 34 | `watch/ctrl-c-detaches-real-grok-kitty-ctrl-c` | `watch` on real grok: kitty Ctrl-C detaches (RED, label real-grok) |
| 35 | `watch/ctrl-c-detaches-grok-modes-tty-cleanup` | `watch` after grok modes: iTerm kitty Ctrl-C restores observer TTY (RED) |
| 38 | `watch/ctrl-c-detaches-grok-modes-post-detach-kitty-garbage` | Post-detach kitty key bytes must not garble host shell (RED) |
| 25 | `watch/grok-like-prompt-clean` | `watch` on grok-like TUI shows prompt without CSI/C0 leaks (RED) |
| 26 | `watch/grok-tui-tty-raw-mirror` | `watch` on TTY passes raw ESC to mirror grok alt-screen UI (RED) |
| 27 | `watch/grok-tui-single-screen-state` | `watch` pipe capture shows one screen, no duplicate redraw smear (RED) |
| 28 | `watch/grok-tui-tty-no-mixed-snapshot-sgr` | `watch` TTY: no plain snapshot before raw SGR (no input garbage) (RED) |
| 14 | `snapshot/prints-sanitized` | `snapshot` prints plain text without escape sequences (RED) |
| 15 | `snapshot/unknown-session` | `snapshot` on missing id fails (RED) |
| 16 | `kill/terminates-detached` | `kill` stops detached session and removes registry (RED) |
| 17 | `kill/unknown-session` | `kill` on missing id fails (RED) |
| 18 | `kill/prunes-unreachable` | `kill` on stale entry prunes registry, exit 0 (RED) |
| 19 | `errors/unknown-subcommand` | Unknown subcommand prints error (RED) |

## How to Run

```sh
doctest vet ./script/tty-watch/tests
doctest test ./script/tty-watch/tests/...
doctest test -v ./script/tty-watch/tests/run/registers-session
```

```go
import (
	"testing"

	"github.com/xhd2015/agent-pro/script/tty-watch/ttywatchtest"
)

type Request = ttywatchtest.Request
type Response = ttywatchtest.Response

func Run(t *testing.T, req *Request) (*Response, error) {
	return ttywatchtest.Run(t, req)
}

func buildTTYWatch(t *testing.T) string {
	return ttywatchtest.BuildTTYWatch(t)
}

func isolatedHome(t *testing.T) string {
	return ttywatchtest.IsolatedHome(t)
}
```