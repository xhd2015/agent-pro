# tty-watch CLI Doctests

End-to-end tests for the standalone `tty-watch` CLI: `run` subcommand embeds ptywrap
(writer attach), `list`, `watch`, `snapshot`, and `kill`.

# DSN (Domain Specific Notion)

**Participants**

- **tty-watch subprocess** — built from `./script/tty-watch`; subcommands `list`,
  `run`, `watch`, `attach`, `snapshot`, `kill`, `send`.
- **Embedded ptywrap server** — HTTP/WS listener on `127.0.0.1:0` inside the
  owning `tty-watch` process for each session.
- **Registry** — `$TTY_WATCH_HOME/registry/` JSON files (`session-N.json` or
  `<custom-id>.json`) with flock-based id reservation (groktty-style).
- **PTY child** — command argv executed inside the embedded terminal session.
- **Test harness** — isolated `TTY_WATCH_HOME` per test, PTY helpers for attach,
  Ctrl-C (`\x03`), Ctrl-] detach (`\x1d`), registry probes.

**Behaviors**

- Default run reserves `session-N`, starts ptywrap, writes registry, attaches as
  interactive writer; **host stays silent** (no session-id on stdout/stderr).
- Optional `run --session-id <id>` (before `<command>`) reserves a user-chosen id;
  pattern `[a-zA-Z0-9][a-zA-Z0-9._-]*`; live duplicate errors; stale entry pruned
  and reused; invalid id rejected before start.
- Ctrl-C forwards interrupt to PTY child; Ctrl-] detaches client only (session +
  registry survive).
- Command exit while attached removes registry entry and exits host.
- `list` scans **all** registry `*.json` files, TCP-probes `listen_addr`, fetches
  ptywrap session info for client counts, prints an aligned table
  (`SESSION`, `UPTIME`, `WATCH`, `ATTACHED`, `COMMAND`) when ≥1 session; prunes
  unreachable entries. `COMMAND` is joined argv, truncated to 64 chars with `...`
  when longer. `WATCH` = observer clients; `ATTACHED` = attach clients +
  1 when screen writer connected.
- `watch` attaches as observer: raw PTY bytes to stdout, no stdin forwarding.
- `attach` joins a live session as multi-writer attacher (`attach_mode=attach`):
  full write+resize, serialized input queue, multiplexed output to all clients.
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
 |    +-- bash-c-raw-tty-crlf/      (LEAF)  bash -c echo yes: CRLF on raw TTY (regression lock)
 |    +-- echo-raw-tty-crlf/        (LEAF)  echo yes: CRLF on raw TTY (regression lock)
 |    +-- cr-overwrite-raw-tty-crlf/ (LEAF)  \\r overwrite + CRLF on raw TTY (regression lock)
 |    |
 |    +-- custom-session-id/
 |         |
 |         +-- registers-custom-id/     (LEAF)  --session-id writes custom registry + list (RED)
 |         +-- duplicate-live-errors/   (LEAF)  live duplicate custom id errors (RED)
 |         +-- reuses-stale-id/          (LEAF)  stale custom registry pruned and reused (RED)
 |         +-- invalid-id-rejected/     (LEAF)  invalid id rejected before start (RED)
 |         +-- list-mixed-with-auto/    (LEAF)  list shows custom id + session-N (RED)
 |
 +-- unit/
 |    |
 |    +-- screen-snapshot-exit-marker/ (LEAF)  snapshot text no leading \\n; exit marker column 0 (RED)
 |    +-- observer-detach-stdin-before-cleanup/ (LEAF) detach restores stdin before stdout cleanup (RED)
 |    +-- observer-detach-kitty-pop-cleanup/ (LEAF) detach pops grok kitty keyboard stack with \\x1b[<u (RED)
 |    +-- attach-stdout-writer-crlf/ (LEAF) attachStdoutWriter normalizes LF-only output on raw TTY (regression lock)
 |    +-- normalize-tty-output/      (LEAF) normalizeTTYOutput LF→CRLF + CR preserve (regression lock)
 |
 +-- list/
 |    |
 |    +-- shows-command-uptime/       (LEAF)  list prints id, command, uptime (RED)
 |    +-- empty-when-none/            (LEAF)  empty registry yields empty list (RED)
 |    +-- second-run-after-exit/      (LEAF)  second run after first exit: list not empty (RED)
 |    +-- table-has-header-and-columns/ (LEAF)  table header + aligned columns (RED)
 |    +-- shows-zero-clients-idle/    (LEAF)  idle session WATCH=0 ATTACHED=0 (RED)
 |    +-- shows-watch-count/          (LEAF)  observer client WATCH=1 (RED)
 |    +-- shows-attached-count/       (LEAF)  attach client ATTACHED=1 (RED)
 |    +-- shows-attached-includes-writer/ (LEAF) screen writer ATTACHED>=1 (RED)
 |    +-- shows-both-counts/          (LEAF)  observer+attach WATCH=1 ATTACHED=1 (RED)
 |    +-- truncates-long-command/      (LEAF)  COMMAND >64 chars truncated with ... (RED)
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
 |    +-- codex-like-single-screen/   (LEAF)  alt-screen TUI: latest screen only, no smear (RED)
 |    +-- codex-cursor-drawn-mcp-boot/(LEAF)  ?2026h+CUP: full warning, no MCP smear/leaks (RED)
 |    +-- codex-mcp-boot-smeared/     (LEAF)  mid MCP boot: no stacked status smear/leaks (RED)
 |    +-- session-dimensions-wide/    (LEAF)  resized session: wide line not 80-col wrapped (RED)
 |    +-- unknown-session/            (LEAF)  missing session errors (RED)
 |
 +-- kill/
 |    |
 |    +-- terminates-detached/        (LEAF)  kill ends detached session (RED)
 |    +-- unknown-session/            (LEAF)  kill missing session errors (RED)
 |    +-- prunes-unreachable/         (LEAF)  stale registry pruned, exit 0 (RED)
 |
 +-- attach/
 |    |
 |    +-- attaches-to-detached-session/   (LEAF)  attach streams live session output (RED)
 |    +-- forwards-stdin-to-pty/          (LEAF)  attach stdin reaches PTY (RED)
 |    +-- second-attach-also-writes/      (LEAF)  two attachers write; both markers queued (RED)
 |    +-- write-visible-to-watch/         (LEAF)  attach write visible to watch (RED)
 |    +-- write-visible-to-other-attach/  (LEAF)  attach A write visible to attach B (RED)
 |    +-- write-visible-while-run-attached/ (LEAF) attach write visible to run writer + watch (RED)
 |    +-- resize-applies-to-pty/          (LEAF)  attach resize updates session cols (RED)
 |    +-- send-uses-input-queue/          (LEAF)  send shares input queue with attach (RED)
 |    +-- detach-survives/                (LEAF)  attach Ctrl-] detach; session survives (RED)
 |    +-- unknown-session/                (LEAF)  attach missing session errors (RED)
 |    +-- concurrent-writes-ordered/      (LEAF)  rapid dual attach writes serialized (RED)
 |
 +-- errors/
      |
      +-- unknown-subcommand/         (LEAF)  bogus subcommand errors (RED)
```

Parameter ranking (most → least significant):

1. **Operation mode** — default run vs list vs watch vs attach vs snapshot vs kill vs errors
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
| 24 | `run/bash-c-raw-tty-crlf` | `run bash -c 'echo yes'` emits CRLF on raw TTY so host prompt stays column 0 (regression lock) |
| 56 | `run/echo-raw-tty-crlf` | `run echo yes` emits CRLF on raw TTY so host prompt stays column 0 (regression lock) |
| 57 | `run/cr-overwrite-raw-tty-crlf` | `printf MARKER_A\\rMARKER_B\\n` preserves CR overwrite and CRLF on raw TTY (regression lock) |
| 47 | `run/custom-session-id/registers-custom-id` | `--session-id` detached run writes `test-with-grok.json`; list shows id + sleep (RED) |
| 48 | `run/custom-session-id/duplicate-live-errors` | Second `run --session-id` same live id exits 1; already in use (RED) |
| 49 | `run/custom-session-id/reuses-stale-id` | Stale custom registry pruned; id reused; live in list (RED) |
| 50 | `run/custom-session-id/invalid-id-rejected` | `run --session-id .bad` exits 1; invalid session id (RED) |
| 51 | `run/custom-session-id/list-mixed-with-auto` | Custom id and auto `session-N` both appear in list (RED) |
| 25 | `unit/screen-snapshot-exit-marker` | Snapshot text must not start with `\n`; exit marker at column 0 (RED) |
| 26 | `unit/attach-stdout-writer-crlf` | `attachStdoutWriter` must call `normalizeTTYOutput` on raw TTY stdout (regression lock) |
| 58 | `unit/normalize-tty-output` | `normalizeTTYOutput` expands bare LF to CRLF; preserves standalone CR (regression lock) |
| 37 | `unit/observer-detach-stdin-before-cleanup` | Detach restores stdin before stdout TTY cleanup (RED) |
| 39 | `unit/observer-detach-kitty-pop-cleanup` | Detach pops grok kitty keyboard stack with `\x1b[<u` (RED) |
| 10 | `list/shows-command-uptime` | `list` shows session id, command argv, uptime (RED) |
| 11 | `list/empty-when-none` | Empty registry prints no sessions (RED) |
| 46 | `list/second-run-after-exit` | Second run after first exit: list shows live session, not empty (RED) |
| 70 | `list/table-has-header-and-columns` | `list` prints SESSION/UPTIME/WATCH/ATTACHED/COMMAND header; columns align (RED) |
| 71 | `list/shows-zero-clients-idle` | Idle detached session: WATCH=0, ATTACHED=0 (RED) |
| 72 | `list/shows-watch-count` | Observer WS client held: WATCH=1 (RED) |
| 73 | `list/shows-attached-count` | Attach WS client held: ATTACHED=1 (RED) |
| 74 | `list/shows-attached-includes-writer` | Screen writer WS held: ATTACHED>=1 (RED) |
| 75 | `list/shows-both-counts` | Observer + attach concurrently: WATCH=1, ATTACHED=1 (RED) |
| 76 | `list/truncates-long-command` | COMMAND joined argv >64 chars: 61-char prefix + `...`, len ≤64 (RED) |
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
| 52 | `snapshot/codex-like-single-screen` | `snapshot` on codex-like alt-screen TUI shows latest screen only (RED) |
| 53 | `snapshot/codex-cursor-drawn-mcp-boot` | `snapshot` on ?2026h codex UI: full warning, no MCP smear/CSI leaks (RED) |
| 54 | `snapshot/codex-mcp-boot-smeared` | `snapshot` mid MCP boot: no stacked status smear/CSI leaks (RED) |
| 55 | `snapshot/session-dimensions-wide` | `snapshot` uses session cols/rows: 95-char line not 80-col wrapped (RED) |
| 15 | `snapshot/unknown-session` | `snapshot` on missing id fails (RED) |
| 16 | `kill/terminates-detached` | `kill` stops detached session and removes registry (RED) |
| 17 | `kill/unknown-session` | `kill` on missing id fails (RED) |
| 18 | `kill/prunes-unreachable` | `kill` on stale entry prunes registry, exit 0 (RED) |
| 19 | `errors/unknown-subcommand` | Unknown subcommand prints error (RED) |
| 59 | `attach/attaches-to-detached-session` | `attach` on detached session shows live output (RED) |
| 60 | `attach/forwards-stdin-to-pty` | Attach stdin marker echoed on attach stdout (RED) |
| 61 | `attach/second-attach-also-writes` | Two attach clients write; both markers present (RED) |
| 62 | `attach/write-visible-to-watch` | Attach write visible to `watch` observer (RED) |
| 63 | `attach/write-visible-to-other-attach` | Attach A write visible to attach B (RED) |
| 64 | `attach/write-visible-while-run-attached` | Attach write visible to run writer and watch (RED) |
| 65 | `attach/resize-applies-to-pty` | Attach resize cols=100; snapshot not 80-col wrapped (RED) |
| 66 | `attach/send-uses-input-queue` | `send` injects via shared input queue while attach connected (RED) |
| 67 | `attach/detach-survives` | Attach Ctrl-] detach exit 0; session in list (RED) |
| 68 | `attach/unknown-session` | Attach missing session id fails (RED) |
| 69 | `attach/concurrent-writes-ordered` | Concurrent attach writes; both end-markers intact (RED) |

## How to Run

```sh
doctest vet ./script/tty-watch/tests
doctest test ./script/tty-watch/tests/...
doctest test -v ./script/tty-watch/tests/run/registers-session
doctest test ./script/tty-watch/tests/run/custom-session-id/...
doctest test ./script/tty-watch/tests/attach/...
doctest test ./script/tty-watch/tests/list/...
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