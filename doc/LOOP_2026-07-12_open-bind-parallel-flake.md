---
title: "Open-bind parallel doctest flake (session id not resolved)"
created: 2026-07-12
slug: open-bind-parallel-flake
path: doc/LOOP_2026-07-12_open-bind-parallel-flake.md
loop_kind: bug-repro
dry_run_status: "VERIFY PASS"
---

# LOOP: open-bind-parallel-flake

## Kind

**bug-repro** — reproduce parallel-only failures in
`cmd/agent-run/tests/status-resume/run-open-background-bind` where isolation
passes but parallel package runs fail with hard bind error.

## Symptom (verbatim class)

Under parallel `doctest test` of the open-background-bind group (or full
`status-resume` suite), these leaves fail:

- `run-open-background-bind/hard-require-without-grok-home-env`
- `run-open-background-bind/prompt-fallback-cwd-mismatch`

```text
agent-run: error: grok session id not resolved for session sess_<pid>
--- FAIL: TestGeneratedCaseHardRequireWithoutGrokHomeEnv
--- FAIL: TestGeneratedCasePromptFallbackCwdMismatch
```

Typical wall time ~24–26s (≈ `openGrokBindPostDetachGrace` 20s + overhead), then
exit 1. Isolation (`doctest test …/<single-leaf>`) is **green**.

Measured reliability (pre-fix, group parallel, 3/3 runs):

| Attempt | Pass | Fail | Flake leaves |
|---------|------|------|--------------|
| 1 | 4 | 2 | both |
| 2 | 4 | 2 | both |
| 3 | 4 | 2 | both |

### Preconditions (do not “fix” in steps 1–4)

- Current branch product code as-is (do **not** serialize tests or widen grace
  in steps 1–4 — those are Fix candidates).
- `doctest` + Go toolchain available.
- Run from repo root.
- Parallel trigger: whole group or full suite (not single leaf).

### Why isolation is green

Each leaf uses isolated `AGENT_RUN_HOME` / `GROK_HOME` (or `HOME`+NoGrokHomeEnv).
Alone, delayed materialize / preseeded prompt-fallback bind succeeds within
budget. Failure appears only when multiple open-bind packages race together.

## Goal

Steps 1–4 **reproduce** the parallel flake on command (**REPRO PASS**).
Step 5 is fix guidance only during establishment; iterate with `/loop-workflow`
until `go run ./script/debug/open-bind-parallel-flake --expect=healthy` is green.

## Surfaces

| Step | Surface |
|------|---------|
| Build | `go build` agent-run (doctest builds via shared session cache) |
| Deploy | N/A local — use tree under `cmd/agent-run/tests/status-resume` |
| Run | parallel `doctest test -count=1` on open-background-bind group |
| Inspect | `script/debug/open-bind-parallel-flake` exit **1** + `REPRO:` lines |

## Pitfalls & blockers

| Issue | Notes |
|-------|-------|
| Isolation confuses “fixed” | Always run **group** parallel; single leaf can stay green forever |
| Cache hides flake | Always `-count=1` |
| Flake “miss” rare | Group parallel was 3/3 RED; if green once, re-run or `--scope=full-suite` |
| frontend embed | status-resume Setup stubs dist; ensure first build can compile |
| Fix must not only `t.Parallel()`-off | Prefer product discovery isolation / timeouts; serial tests are last resort |

---

## Step 1 — Build

```sh
cd "$(git rev-parse --show-toplevel)"
mkdir -p frontend-agent-run/dist frontend/dist
test -f frontend-agent-run/dist/index.html || echo '<!doctype html><title>stub</title>' > frontend-agent-run/dist/index.html
test -f frontend/dist/index.html || echo '<!doctype html><title>stub</title>' > frontend/dist/index.html

# Compile inspect helper + ensure agent-run builds
go build -o /dev/null ./script/debug/open-bind-parallel-flake
go build -o "${TMPDIR:-/tmp}/agent-run-open-bind-parallel-flake" ./cmd/agent-run
test -x "${TMPDIR:-/tmp}/agent-run-open-bind-parallel-flake"
```

**Verify:** both `go build` commands exit 0.

---

## Step 2 — Deploy / Update

Local only — no remote deploy. Confirm tools:

```sh
which doctest go
doctest version 2>/dev/null || doctest -h 2>&1 | head -3
test -d cmd/agent-run/tests/status-resume/run-open-background-bind
```

**Verify:** `doctest` on PATH; open-background-bind directory exists.

---

## Step 3 — Run (trigger failure only)

Do **not** change product code, widen discovery grace, or force serial tests.

**Preferred (orchestrated):**

```sh
cd "$(git rev-parse --show-toplevel)"
go run ./script/debug/open-bind-parallel-flake
# default --expect=repro --scope=group
```

**Manual equivalent:**

```sh
# isolation (expect PASS)
doctest test -count=1 \
  ./cmd/agent-run/tests/status-resume/run-open-background-bind/hard-require-without-grok-home-env
doctest test -count=1 \
  ./cmd/agent-run/tests/status-resume/run-open-background-bind/prompt-fallback-cwd-mismatch

# parallel group (expect FAIL + not resolved)
doctest test -count=1 \
  ./cmd/agent-run/tests/status-resume/run-open-background-bind
```

**Verify:** step 3 alone may exit non-zero — that is expected. Do not “fix” here.

---

## Step 4 — Inspect / Feedback (symptom present)

```sh
cd "$(git rev-parse --show-toplevel)"
go run ./script/debug/open-bind-parallel-flake --expect=repro --scope=group
echo "inspect_exit=$?"
```

**REPRO PASS criteria (bug still present):**

- Inspect exit code **1** (non-zero)
- Stdout contains `REPRO:` lines
- Evidence includes `grok session id not resolved` and/or FAIL of
  `hard-require-without-grok-home-env` / `prompt-fallback-cwd-mismatch`

**Not REPRO** (exit 2 / missing symptom): parallel green this run, or different
failure mode — re-run or use `--scope=full-suite`.

### Step 4b — Inspect verify (post-fix only; after step 5)

```sh
go run ./script/debug/open-bind-parallel-flake --expect=healthy --scope=group
# optional stress:
# for i in 1 2 3; do go run ./script/debug/open-bind-parallel-flake --expect=healthy || exit 1; done
```

**VERIFY PASS:** exit 0, `VERIFY: parallel open-bind … fully green`.

---

## Step 5 — Fix (guidance only during establishment)

When REPRO is reliable, investigate and fix (then re-run steps 1→4b). Candidates:

1. **Discovery path under load** — `pkgs/agentui/open_bind.go` worker +
   `pkgs/agenttty.DiscoverSession` / `scanAllSessionsForPrompt`; confirm
   `GrokHomeForRunner` / `HOME` / `GROK_HOME` match materialization path for
   O1 (`NoGrokHomeEnv`) and O3 (cwd-mismatch preseed).
2. **created_at vs runStart** — `sessionNotBefore` grace is 2s; delayed materialize
   or slow Setup under CPU contention may need product/test timestamp alignment
   (prefer fixing discovery robustness, not only increasing grace).
3. **Post-detach budget** — failures burn full ~20s `openGrokBindPostDetachGrace`
   then hard-error; if session exists on disk but unmatched, fix match keys
   (prompt, cwd encoding, hook id), not only longer wait.
4. **Parallel harness contamination** — shared env hooks
   (`AGENT_RUN_GROK_TTY_*`, ambient `GROK_HOME` from parent) leaking across
   packages; ensure child env fully isolated in root SETUP.
5. **Last resort** — doctest/go test serialization for these leaves only if product
   is proven correct and flake is pure harness (document why).

Return to step 1 after each fix attempt. Establishment of this loop does **not**
require applying the fix.

---

## Dry-run log

| Timestamp (UTC) | Step | Result | Evidence |
|-----------------|------|--------|----------|
| 2026-07-12T00:20:51Z | isolation hard-require | PASS | 1/1 in ~10s |
| 2026-07-12T00:21:00Z | isolation prompt-fallback | PASS | 1/1 in ~3s |
| 2026-07-12T00:21:05Z | parallel group (manual) | RED | prompt-fallback FAIL, not resolved |
| 2026-07-12T00:21:50Z | parallel group ×3 | RED 3/3 | both leaves fail; pass=4 fail=2 each time |
| 2026-07-12T00:24:35Z | step 1 build | PASS | go build open-bind-parallel-flake + agent-run |
| 2026-07-12T00:24:40Z | step 2 tools | PASS | doctest + go on PATH |
| 2026-07-12T00:25:14Z | step 3–4 inspect repro | **REPRO PASS** | isolation both exit 0; parallel exit 1; `REPRO: "grok session id not resolved" x1`; leaf `prompt-fallback-cwd-mismatch` FAIL (5/6); inspect exit 1 |
| 2026-07-12T00:30:35Z | debug logs | root cause | session on disk; `discErr=context canceled`; `sessionDiscoveryGrace=2s` rejected preseed/materialize under parallel lag |
| 2026-07-12T00:32:09Z | step 5 fix | applied | `sessionDiscoveryGrace` 2s → 5m in `pkgs/agenttty/paths.go` |
| 2026-07-12T00:32:48Z | step 4b verify | **VERIFY PASS** | `--expect=healthy` exit 0; pass=6 fail=0 |
| 2026-07-12T00:32:53Z | stress ×3 | **VERIFY PASS** | 3/3 healthy green |
| 2026-07-12T00:35:21Z | full status-resume | PASS | 29/29 |

---

## Aux scripts

| Path | Role |
|------|------|
| `script/debug/open-bind-parallel-flake/main.go` | Isolation preflight + parallel trigger; `REPRO:` / `VERIFY:` |

Related (different bug — soft bind unbound on real grok):

- `doc/LOOP_2026-07-11_open-bind-runner-unbound.md`
- `script/debug/open-bind-smoke`

---

## Handoff

- **Loop kind:** bug-repro  
- **Dry-run status:** REPRO PASS  
- **Next:** `/loop-workflow` fix open-bind parallel flake until  
  `go run ./script/debug/open-bind-parallel-flake --expect=healthy` exits 0  
  (optionally stress 3× and full `status-resume` suite).
