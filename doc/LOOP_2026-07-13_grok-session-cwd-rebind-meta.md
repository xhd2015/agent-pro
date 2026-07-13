---
title: "Grok session cwd rebind via meta file experiment"
created: 2026-07-13
slug: grok-session-cwd-rebind-meta
path: doc/LOOP_2026-07-13_grok-session-cwd-rebind-meta.md
loop_kind: regression
dry_run_status: "VERIFY PASS"
---

# LOOP: grok-session-cwd-rebind-meta

## Kind

**regression** — (1) REPRO sticky tool cwd after workspace delete; (2) VERIFY
rebind via meta/session-dir patches. **Minimal fix found: `move_dir` alone.**

## Symptom

When the original workspace path is gone and we resume with a different process
cwd (no meta patch), Grok still knows the old path and shell tools do not cleanly
bind to the new dir. Observed signatures:

```
Session <id> found locally (originally in /…/ws-old)
The workspace path may be gone. …
The workspace directory is missing, which is why the shell cannot start.
…/ws-old
```

Also: `summary.json` → `info.cwd` remains the deleted path; tool
`current_dir` in `updates.jsonl` may still be the old workspace.

**Repro preconditions (document only — do not “fix” in steps 1–4):**

1. Isolated `GROK_HOME` (not `~/.grok`) seeded with `auth.json`.
2. Session created under temp `ws-old` with shell `pwd` once.
3. `ws-old` **deleted**; `ws-new` created.
4. Resume **without** editing session meta files (no step-5 patches).

## Goal

On command, steps 1–4 **reproduce** sticky/broken session cwd (inspect RED /
`REPRO:`). After step 5 meta experiments, optional verify mode can show whether
any field set rebinds tool `pwd` to `ws-new` (`VERIFY:`).

## Prerequisites (agent-auditable)

| Check | Command | Pass criteria |
|-------|---------|---------------|
| grok binary | `which grok` | exit 0 |
| auth for custom home | `test -f "$HOME/.grok/auth.json"` | file exists (copied into `GROK_HOME`) |
| network/API | first `grok -p …` under custom home | not “Not signed in” after auth copy |
| go toolchain | `which go` | exit 0 (for inspect script) |

**BLOCKER:** custom `GROK_HOME` without seeding `auth.json` → “Not signed in”.
Unblock: `cp ~/.grok/auth.json "$GROK_HOME/"` (or set `XAI_API_KEY` if that is
how this grok build authenticates).

## Surfaces

| Step | Surface |
|------|---------|
| Build | `go build` / `go run` of inspect script (no agent-run required) |
| Deploy | N/A (local only); optional `GROK_HOME` under `/tmp` |
| Run | real `grok -p` create + resume |
| Inspect | `script/debug/grok-session-cwd-rebind` |
| Fix | edit session meta under `$GROK_HOME/sessions/…` then `--expect=healthy` |

## Magic fields inventory (candidates for step 5)

Under `$GROK_HOME/sessions/<url-encoded-cwd>/<session-id>/`:

| File | Field / note |
|------|----------------|
| `summary.json` | `info.cwd` (primary); sometimes `git_root_dir` |
| `prompt_context.json` | `working_directory` |
| session **parent dir name** | URL-encoded cwd (`%2Fprivate%2Ftmp%2F…`) — may need rename/`move_dir` |
| `chat_history.jsonl` | embeds `Workspace Path:` in system/user blobs |
| `updates.jsonl` | historical `current_dir` (likely not re-read as live cwd) |
| `system_prompt.txt` | may embed workspace path |

## Loop steps

### 1. Build

```sh
cd "$(git rev-parse --show-toplevel)"
go build -o /tmp/grok-session-cwd-rebind ./script/debug/grok-session-cwd-rebind
test -x /tmp/grok-session-cwd-rebind
```

**Verify:** `/tmp/grok-session-cwd-rebind -h 2>&1 | head -5` shows usage flags
(`--expect`, `--patch`, `--keep-root`) or binary runs and prints `FAIL:` /
`PHASE_A` (no compile error).

### 2. Deploy / Update

Local-only. Ensure auth source exists:

```sh
test -f "$HOME/.grok/auth.json"
which grok
```

**Verify:** both exit 0.

### 3. Run

Trigger **only** the failure path (create session → delete old dir → resume with
`--cwd` NEW → **no** meta patch). Prefer the inspect script which encodes this:

```sh
cd "$(git rev-parse --show-toplevel)"
# keep evidence dir for dry-run log
KEEP="$(mktemp -d /tmp/grok-cwd-rebind-run-XXXXXX)"
go run ./script/debug/grok-session-cwd-rebind --expect=repro --keep-root "$KEEP"
echo "exit=$?"
```

Manual equivalent (if debugging by hand):

```sh
ROOT=$(mktemp -d /tmp/grok-cwd-manual-XXXXXX)
export GROK_HOME="$ROOT/grok-home"
mkdir -p "$GROK_HOME" "$ROOT/ws-old"
cp "$HOME/.grok/auth.json" "$GROK_HOME/auth.json"
chmod 600 "$GROK_HOME/auth.json"

# create
grok -p 'Use the shell tool to run exactly: pwd
Reply with a single line that is ONLY the absolute path from that command (no markdown).' \
  --cwd "$ROOT/ws-old" --always-approve --permission-mode=bypassPermissions \
  --output-format plain

# locate session id from $GROK_HOME/sessions/**/summary.json
# then:
rm -rf "$ROOT/ws-old"
mkdir -p "$ROOT/ws-new"
grok -p 'Use the shell tool to run exactly: pwd
Reply with a single line that is ONLY the absolute path from that command (no markdown).' \
  --resume "$SID" --cwd "$ROOT/ws-new" --always-approve \
  --permission-mode=bypassPermissions --output-format plain
```

**Verify:** script exits **non-zero** and prints `REPRO:` (or manual run shows
Symptom strings). Do **not** apply meta patches here.

### 4. Inspect / Feedback — Repro

```sh
# If step 3 used go run with --keep-root, evidence is under KEEP.
# Re-run inspect-only is the same command: symptom must be present.

go run ./script/debug/grok-session-cwd-rebind --expect=repro
echo "exit=$?"
```

**Verify (REPRO PASS):**

- Exit code **1** (non-zero)
- stdout contains `REPRO:`
- Evidence includes at least one of:
  - `originally in` + old workspace path
  - `workspace path may be gone` / `workspace directory is missing`
  - `TOOL_CURRENT_DIR` still equal to old cwd
  - `OUT_MENTIONS_OLD: true` with `HEALTHY: false`

### 4b. Inspect / Feedback — Verify (post step 5 only)

```sh
go run ./script/debug/grok-session-cwd-rebind \
  --expect=healthy \
  --patch=summary,prompt_context,move_dir
# optional variants:
# --patch=summary
# --patch=summary,prompt_context
# --patch=summary,prompt_context,move_dir,chat_history
```

**Verify (VERIFY PASS, after a working patch set):**

- Exit code **0**
- stdout contains `VERIFY:`
- `TOOL_CURRENT_DIR` or final `pwd` answer equals **new** workspace (not old)

Skip 4b during establishment dry-run (bug-repro only needs REPRO PASS).

### 5. Fix → return to step 1

**Working minimal fix (verified 2026-07-13):**

```sh
# Move session dir to new cwd key (url-encoded path). That alone rebinds tools.
go run ./script/debug/grok-session-cwd-rebind --expect=healthy --patch=move_dir
```

| Patch set | Result |
|-----------|--------|
| `summary` only | FAIL — tool stays on old |
| `prompt_context` only | FAIL |
| `summary,prompt_context` | FAIL |
| **`move_dir` only** | **PASS** |
| `summary,move_dir` / `prompt_context,move_dir` / full | PASS |

**Why:** Grok stores sessions at  
`$GROK_HOME/sessions/<url-encoded-cwd>/<session-id>/`.  
Lookup/binding is keyed by that path. In-place edits of `info.cwd` /
`working_directory` without relocating the directory do **not** rebind
`run_terminal_command` cwd.

Optional extras (not required if `move_dir` applied):

1. Patch `summary.json` `info.cwd`  
2. Patch `prompt_context.json` `working_directory`  
3. String-replace old path in `chat_history.jsonl`  

Product follow-up (agent-run): when `--dir` overrides a missing
`meta.workspace`, also **move** the Grok session under
`$GROK_HOME/sessions/<encoded-new-dir>/` (and optionally rewrite summary cwd).

## Pitfalls & blockers

| Issue | Mitigation |
|-------|------------|
| Custom `GROK_HOME` unsigned | Copy `~/.grok/auth.json` into new home |
| `/tmp` vs `/private/tmp` path drift | Compare with `realpath` / strip `/private` |
| Resume with `--cwd NEW` may create a **second** session under NEW | Prefer resume by session id; inspect `TOOL_CURRENT_DIR` on the **same** id |
| Agent **recreates** deleted `ws-old` via shell | Repro still valid if it first fails / mentions missing old path; script keeps evidence root |
| Long LLM latency | `--timeout` default 3m; increase if flaky |
| Do not use real `~/.grok` | Always isolate `GROK_HOME` under the experiment root |
| Mutating live user sessions | Never point `GROK_HOME` at `~/.grok` for patch experiments |

## Aux script

```
script/debug/grok-session-cwd-rebind/main.go
```

| Mode | Exit | Marker |
|------|------|--------|
| `--expect=repro` (default) | 1 when symptom present | `REPRO:` |
| `--expect=healthy` | 0 when pwd/tool cwd == NEW | `VERIFY:` |
| setup/auth failure | 2 | `FAIL:` |

Flags: `--patch`, `--keep-root`, `--timeout`, `--grok`.

## Dry-run log

| Timestamp (UTC) | Step | Result | Evidence |
|-----------------|------|--------|----------|
| 2026-07-13T06:34Z | 1 Build | OK | script compiles via `go run` |
| 2026-07-13T06:34Z | 2 Deploy | OK | `which grok`; `~/.grok/auth.json` present |
| 2026-07-13T06:34Z | 3 Run (manual phase A) | OK | `pwd` → `/private/tmp/grok-cwd-loop-0oRikh/ws-old`; sid `019f5a2e-fe10-7602-9a39-90c64c0cea33` |
| 2026-07-13T06:35Z | 3 Run (manual phase B, no patch) | OK trigger | deleted `ws-old`; resume `--cwd ws-new` |
| 2026-07-13T06:37Z | 4 Inspect (manual) | **REPRO PASS** | stderr: `Session … found locally (originally in /private/tmp/…/ws-old)`; stdout: `The workspace path may be gone…`; `info.cwd` still old |
| 2026-07-13T06:39Z | 1–4 script dry-run | **REPRO PASS** | exit **1**; `TOOL_CURRENT_DIR` still old; `REPRO:` lines |
| 2026-07-13T06:43Z | 5 + 4b full patch | almost | tool on NEW but inspect false-failed (sess.CWD compare bug) |
| 2026-07-13T06:44Z | 4b after inspect fix | **VERIFY PASS** | `--patch=summary,prompt_context,move_dir` exit 0 |
| 2026-07-13T06:44Z | 4 repro recheck | **REPRO PASS** | still sticky without patch |
| 2026-07-13T06:47Z | 5 minimal matrix | **VERIFY PASS** | **`move_dir` alone** sufficient; summary/prompt alone FAIL |

---

## Handoff

- **REPRO** and **VERIFY** both green for inspect script.
- Minimal rebind: **`move_dir`** (relocate session under new encoded cwd).
- Optional next product work: teach `agent-run resume --dir` to perform the
  same Grok session directory move when `meta.workspace` is missing/overridden.
