# grok-writable snapshot fixtures

Rendered **tty-watch snapshot text** (plain printable output from `ttywatch.SnapshotText`
/ `tty-watch snapshot`), passed to `checkGrokWritable` — not raw PTY ANSI bytes.

## Layout

```
pkgs/agenttty/testdata/grok-writable/
├── README.md
├── expectations.jsonl     # one JSON object per line: file, ready, state, reason?, tags, source?
└── grok-*.txt             # snapshot fixtures
```

### expectations.jsonl fields

| Field | Required | Description |
|-------|----------|-------------|
| `file` | yes | Basename of the `.txt` fixture in this directory |
| `ready` | yes | Expected `WritableStatus.Ready` |
| `state` | yes | Expected `WritableStatus.State` (`idle`, `busy`, `loading`, `unknown`, …) |
| `reason` | no | Expected `WritableStatus.Reason` when non-empty |
| `tags` | yes | String array for filtering / documentation |
| `source` | no | How the frame was captured or synthesized |
| `banner_detected_legacy` | no | Expected legacy banner detector result (open-lifecycle characterization) |
| `open_ready` | no | Expected open-lifecycle readiness (`OpenReady`) |
| `screen_class` | no | Coarse class: `empty` \| `starting` \| `busy` \| `idle` \| `modal` \| `unknown` |

F1 `all-expectations-match` asserts only `ready`/`state`/`reason`. Optional open-ready fields
document desired open-lifecycle behavior and are enforced by regression leaves.
Writable idle/busy for modern chrome is section-based (`ParseGrokFrame` /
`judgeGrokFrameBusy`); legacy `Grok ›` / prompt-region substrings are not used.

Naming convention (probe export):

```
grok-<phase>-<state>-<detail>-<hash8>.txt
```

Regression / synthesized frames use descriptive suffixes (e.g.
`grok-after_git-idle-false-positive-session18-synth.txt`).

## Regenerate from probe capture

From repo root, with `tty-watch` and `grok` on PATH:

```sh
# 1. Capture live session snapshots
go run ./script/debug/grok-writable-probe \
  -scenario=follow-up \
  -duration=20s \
  -out=/tmp/grok-writable-capture

# 2. Export curated fixtures + expectations.jsonl (after implementer adds -export-fixtures)
go run ./script/debug/grok-writable-probe \
  -export-fixtures=pkgs/agenttty/testdata/grok-writable \
  -from=/tmp/grok-writable-capture
```

Re-export from an existing capture dir without re-probing:

```sh
go run ./script/debug/grok-writable-probe \
  -export-fixtures=pkgs/agenttty/testdata/grok-writable \
  -from=/tmp/grok-writable-run1
```

## Manual single-frame capture

```sh
# Detached grok session (example session id)
tty-watch run --detach --session-id=grok-debug ~/.grok/bin/grok --always-approve

# When reproducing the session-18 false positive, run git status first so scrollback
# contains "git working tree status", then snapshot while the prompt is idle:
tty-watch send grok-debug $'run git status\n'
# wait for turn to complete…
tty-watch snapshot grok-debug > pkgs/agenttty/testdata/grok-writable/grok-manual-after_git-idle-<note>.txt
```

Add a matching line to `expectations.jsonl` with the expected `ready`/`state`/`reason`.

## Verify

```sh
doctest vet ./pkgs/agenttty/tests/grok-writable
doctest test ./pkgs/agenttty/tests/grok-writable
go test ./pkgs/agenttty/ -run TestCheckGrokWritable -v
```

## Current seed sources

- **Probe run** `/tmp/grok-writable-run1` — 15 unique snapshot hashes across phases
  `boot`, `after_ls`, `after_recap`, `after_git`, `after_hello`
- **Session-18 false positive** — `grok-after_git-idle-false-positive-session18-synth.txt`
  (scrollback contains `git working tree status`, prompt idle; must be `ready=true` after fix)
- **Inline regression strings** from `pkgs/agenttty/writable_test.go` (busy thinking, legacy
  prompt, heavy post-turn prompt)
- **Workspace project-directory confirm** —
  `grok-workspace-project-directory-confirm.txt` (smoke `/tmp` open; must be
  `ready=false`, not `idle`; legacy banner FP; `open_ready=false`, `screen_class=modal`)
- **Modern SeaTalk chrome** (from `script/debug/grok-screen-snapshots/`):
  - `grok-modern-starting-session-chrome.txt` ← `02-early-tui-chrome-or-input.txt`
  - `grok-modern-busy-thinking-tasks.txt` ← `03-busy-working.txt`
  - `grok-modern-idle-input-post-turn.txt` ← `04-idle-input-ready.txt`
  Writable option A recorded as-is; open-ready true for starting/busy/idle (implementer)
- **Post-turn Recap + `Ctrl+e:expand thinking` footer** —
  `grok-after_recap-expand-thinking-idle-01a03d6f.txt` (`kck grok snapshot`
  `01a03d6f-…`; `Worked for` + Recap + placeholder `Build anything`). Desired
  `ready=true`/`idle`.
- **Status chrome fixtures (captured)**:
  - Small (`w106`, llm-mock-run-grok + tty-watch):
    - `grok-status-waiting-for-response-busy-w106.txt`
    - `grok-status-responding-busy-w106.txt`
    - `grok-status-worked-for-idle-w106.txt`
  - Wide live (`kck grok snapshot`, ~w160–w188 vs baseline w175):
    - `grok-status-waiting-for-response-busy-live-w185.txt`
    - `grok-status-waiting-for-response-busy-live-w167.txt` (expand-thinking box above Waiting + `[stop]`)
    - `grok-status-thinking-busy-live-w188.txt`
    - `grok-status-worked-for-idle-live-w186.txt`
    - `grok-status-command-still-running-busy-live-w160.txt` (`◎ N command still running · send a message to interrupt`)
    - `grok-status-running-tool-busy-live-w187.txt` (spinner + tool title + `Ctrl+b:send to bg`; live tool-run chrome)
  Desired: waiting/thinking/responding/running-tool → `status_above_composer` → busy; command-still-running → `running_indicator` → busy;
  worked-for → idle (bare `stop`). StatusAboveComposer beats an earlier WorkedFor in the same frame.
  Literal `Running command` chrome still not seen in this harvest; tool-run spinner is the live equivalent captured.
