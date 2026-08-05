# codex-writable snapshot fixtures

Rendered **tty-watch snapshot text** (plain printable output from `ttywatch.SnapshotText`
/ `tty-watch snapshot`), passed to `checkCodexWritable` — not raw PTY ANSI bytes.

## Layout

```
pkgs/agenttty/testdata/codex-writable/
├── README.md
├── expectations.jsonl     # one JSON object per line: file, ready, state, reason?, tags, source?
└── codex-*.txt            # snapshot fixtures
```

### expectations.jsonl fields

| Field | Required | Description |
|-------|----------|-------------|
| `file` | yes | Basename of the `.txt` fixture in this directory |
| `ready` | yes | Expected `WritableStatus.Ready` |
| `state` | yes | Expected `WritableStatus.State` (`idle`, `busy`, `loading`, `unknown`, …) |
| `reason` | no | Expected substring of `WritableStatus.Reason` when non-empty |
| `tags` | yes | String array for filtering / documentation |
| `source` | no | How the frame was captured or synthesized |

## Live capture seed (status Fetch investigation)

| Fixture | Screen (post-fix expectation) |
|---------|--------------------------------|
| `codex-update-available-modal.txt` | Update picker only (`› 1. Update now…`) → **not** idle (`ready=false`, `state=loading`) |
| `codex-update-plus-model-loading.txt` | Update banner + `model: loading` + prompt placeholder → `loading` |
| `codex-mcp-incomplete-prompt.txt` | MCP incomplete + main chat `›` → `ready=true`, `idle` |
| `codex-main-prompt-mcp-incomplete.txt` | Alternate main prompt + MCP incomplete → `ready=true`, `idle` |
| `codex-empty-snapshot.txt` | Empty bytes → `unknown` |
| `codex-double-angle-prompt-idle.txt` | Codex 0.146 idle with only `»` (U+00BB) + usage-limit bullet → `ready=true`, `idle` |
| `codex-double-angle-mcp-incomplete.txt` | MCP incomplete + only `»` (no `›`) → `ready=true`, `idle` |

Sources: live captures from `/tmp/codex-status-fixtures-for-req/` (content preserved) plus
synthetic Codex 0.146 double-angle fixtures from the SeaTalk incident.

**Critical bugs:**

1. Update-available modal was classified `ready=true` / `state=idle` because of `›` on the
   menu option line, so `FetchStatus` / `waitForPrompt` sent `/status` into the modal.
2. Codex **v0.146.0** main-chat prompt is `»` (U+00BB); heuristics that only match `›`
   (U+203A) leave `ready=false` / `state=unknown` until timeout.

## Verify

```sh
doctest vet ./pkgs/agenttty/tests/codex-writable
doctest test ./pkgs/agenttty/tests/codex-writable
doctest test -v ./pkgs/agenttty/tests/codex-writable/regression/update-modal-not-idle
```
