---
name: verify-on-behalf-of-user/host
description: >-
  Opt-in host (outside sandbox) mode: real HOME/PATH, change-scoped install via
  wrk --reinstall-local or script/go install fallbacks, dry-run/plan before
  mutate, and mandatory warnings. Default remains sandbox unless user clearly opts in.
---

# Host (outside sandbox)

**Mode id:** `host`  
**Default is always `sandbox`.** Use host only when the user **explicitly** opts in.

This mode verifies with the **real** user environment: real `HOME`, real
`PATH` / GOBIN binaries. It can mutate installed tools. Prefer sandbox whenever
the user did not clearly authorize host mode.

## When to select host

Treat as **Mode: host** only if the request clearly authorizes outside-sandbox
work. Accept paraphrases with the same meaning, case-insensitive:

| Family | Examples |
|--------|----------|
| host | "host mode", "on the host", "real host" |
| outside sandbox | "outside sandbox", "outside the sandbox", "without sandbox", "no sandbox", "unsandboxed" |
| real install / env | "real HOME", "install to my real bin", "`wrk --reinstall-local`" as the verify method, "use GOBIN / ~/go/bin for verify" |
| explicit pairing | "verify outside sandbox", "verify in host mode" |

### Stay in sandbox (do not switch)

- Bare "verify on my behalf" / "verify before commit"
- "install", "build", "run it for real" **without** host / outside-sandbox language
- "strict sandbox" / `sandbox-exec` (still sandbox)
- Ambiguous or mixed signals without a clear opt-out of sandbox

**Rule:** if a reasonable reader could think the user still wanted isolation →
**Mode: sandbox**. Do not auto-promote to host.

## Mandatory warnings

Before any host install or other mutating step, show warnings to the user and
record them in the transcript:

```text
warning: host mode requested — verification uses real HOME and installs binaries
         into GOBIN (not ~/.sandbox/bin).
warning: mutating steps prefer --dry-run / a written plan first; plans are
         recorded in the transcript before apply.
```

Do **not** source `enter-sandbox.sh` in host mode.

## Resolve change-scoped targets (required)

Host install only covers tools **relevant to the change**. Resolve target names
`T` from (in order):

1. User wording ("verify `my-tool`", "I changed agent-pro")
2. Claim / agent summary of what was done
3. Project recipe surface/bin list for that feature
4. Git diff heuristics only if still unclear (touched `cmd/<name>/`, `script/<name>/`)

If no target can be named → do **not** fan out all installers; plan "targets
unknown" and either ask once or **FAIL** host install with a clear error.

For each target `T` (example: `my-tool`), only consider paths relevant to `T`:

| Path | Role |
|------|------|
| `./cmd/T` | `go install ./cmd/T` candidate |
| `./script/T/install` | preferred scoped installer |
| `./scripts/T/install` | plural twin if present |
| `./script/install` | only if it installs `T` / product that includes `T` (fan-in / compat entry) |
| `./scripts/install` | same rule |

**Do not** auto-run unrelated `./script/other-tool/install` or every
`./script/*/install`.

## Phase 2 — Install ladder (host only)

Prefer the first rung that can install the claimed targets. Record **Install
method**, **Targets**, **Checked**, **Chosen**, and **Skipped** in the transcript.

### 1. `wrk --reinstall-local` (preferred when available)

```sh
command -v wrk   # or shell function available
wrk --reinstall-local --dry-run    # always first
# show plan; ensure claimed targets are in intended outcomes
wrk --reinstall-local              # apply unless user said dry-run only
```

Optional: `wrk --main --reinstall-local` (and dry-run twin) when verifying from a
linked worktree but the user wants main-tip binaries.

If wrk's plan clearly will not cover claimed `T`, fall through or **supplement**
with script / `go install` for `T` only.

### 2. Repo install scripts (change-scoped)

If `wrk` is missing or insufficient for targets:

```text
warning: wrk not available (or insufficient for targets); falling back to
         change-scoped repo install scripts
```

Discovery for each `T`:

- `./script/T/install`, `./scripts/T/install`
- `./script/install`, `./scripts/install` only when relevant to `T`

Prefer `go run ./script/.../install` (or `go run ./script/install`) over guessing
shell wrappers.

**Synthetic plan** (scripts usually have no `--dry-run`): list exact commands,
cwd, and destination bin dir; then apply unless dry-run-only.

### 3. `go install ./cmd/T`

If scripts are missing but `./cmd/T` exists for each claimed target:

```sh
GOBIN="$(go env GOBIN)"; [ -n "$GOBIN" ] || GOBIN="$(go env GOPATH)/bin"
# planned (example T=my-tool):
go install ./cmd/my-tool
```

Prefer scoped packages over `go install ./...` unless the recipe says otherwise.
Emit a written plan, then apply.

### 4. None usable → FAIL

```text
Error: host mode could not install binaries for targets [<T>…]
  checked: wrk; ./script/<T>/install; ./script/install; ./cmd/<T>
Verdict: FAIL
```

## Dry-run / plan policy

| Op | Policy |
|----|--------|
| `wrk --reinstall-local` | Always `--dry-run` → record/show → apply |
| Tools with `--dry-run` / `--plan` | Prefer dry-run first when the op may write system/user state |
| Ops with no dry-run | Explicit `warning:` + synthetic plan (commands + dest); never silent |
| User said dry-run only | Stop after plans; do not apply |

Default after a successful plan: **apply in the same verify run** (unless
dry-run-only).

## What stays the same as sandbox mode

- Always label depth (`smoke` \| `scenario` \| `full` + reason)
- Scenario-first; browser-agent for UI; tty-watch for interactive TTY
- Transcript file **and** full inline body
- Transcript path remains `~/.sandbox/transcripts/<ISO8601>-<slug>.md` (logs only;
  does not imply sandbox runtime — header **Mode** must be `host`)

## Forbidden in host mode

- Pretending sandbox isolation while writing real GOBIN / real HOME configs
- Skipping dry-run/plan before `wrk --reinstall-local` or other mutating ops
- Installing unrelated binaries not in the change-scoped target list
- Sourcing `enter-sandbox.sh` then claiming host mode

## Transcript header (host)

| Field | Value |
|-------|-------|
| Mode | **host** |
| Install method | `wrk` \| `script` \| `go-install` \| `none` |
| Targets | e.g. `my-tool` |
| HOME / bin | real HOME; `GOBIN` or `$(go env GOPATH)/bin` |

See also: `sandbox` (default mode), `workflow` (mode-branched phases).
