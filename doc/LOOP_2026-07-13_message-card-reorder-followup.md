---
title: "Message card reorder on grok-tty follow-up"
created: 2026-07-13
slug: message-card-reorder-followup
path: doc/LOOP_2026-07-13_message-card-reorder-followup.md
loop_kind: bug-repro
dry_run_status: "VERIFY PASS"
---

# LOOP: message-card-reorder-followup

## Kind

**bug-repro** — reproduce live UI reorder of the first user message card after a
grok-tty follow-up send. Refresh restores correct order.

## Symptom (verbatim class)

User flow:

1. Run `agent-run web` on a given port
2. Open it in the browser
3. Choose **grok-tty**
4. Send `run ls`
5. After it finished, send `what did I say`

**Observed bug (live, no refresh):** the message card `run ls` **changes its
position** to sit **above** the `what did I say` card (typically adjacent user
cards with **no assistant between them**, or assistant content for turn 1
appears **before** both user cards). **Refresh** restores a correct timeline.

Inspect evidence (orchestrator):

```text
REPRO: live message-card reorder on grok-tty follow-up
REPRO: "run ls" immediately above "what did I say" with no assistant between (live reorder)
```

or

```text
REPRO: assistant at N appears before first user "run ls" at M
```

### Preconditions (do not “fix” in steps 1–4)

- Product code as-is (do **not** change `mergeSessionEvents` / grok-tty user-strip
  in `frontend-agent-run/src` during steps 1–4).
- `go`, `bun`, `playwright-debug` on `PATH`.
- Run from repo root.
- Isolated `AGENT_RUN_HOME` + mock `llm-mock-run-grok` (orchestrator sets these).

### Why refresh is green

Server-side session events after reload are chronological. The live bug is
client-side: after follow-up send, grok-tty path filters server user events while
`lastSentUserPromptRef` is set, then re-appends optimistic users at the **end**
of the timeline (`mergeSessionEvents` + strip in `SessionPage` refresh).

## Goal

Steps 1–4 **reproduce** the live reorder on command (**REPRO PASS**).
Step 5 is fix guidance only during establishment; iterate with `/run-the-loop`
until:

```sh
go run ./script/debug/message-card-reorder-followup --expect=healthy
```

exits 0 (**VERIFY**).

## Surfaces

| Step | Surface |
|------|---------|
| Build | `bun run build` in `frontend-agent-run`; `go build` `agent-run` + `llm-mock-run-grok` |
| Deploy | Local: start `agent-run web` (open API, grok-tty mock) |
| Run | Playwright: home → grok-tty → `run ls` → wait → `what did I say` |
| Inspect | `script/debug/message-card-reorder-followup` exit **1** + `REPRO:` lines |

## Pitfalls & blockers

| Issue | Notes |
|-------|-------|
| Real grok / auth | Orchestrator uses open API + `llm-mock-run-grok`; no token |
| Port busy | Orchestrator picks free port (`-port=0` default) |
| playwright-debug missing | Install / ensure `~/go/bin` on PATH |
| Single-turn mock only | Hook must complete turn 1 and turn 2 (orchestrator multi-turn hook) |
| Do not “fix” during REPRO | No frontend merge changes in steps 1–4 |
| Reload-only green | `--expect=healthy` requires **live** order fixed, not only after reload |

---

## Step 1 — Build

```sh
cd "$(git rev-parse --show-toplevel)"

# Frontend embed assets
(cd frontend-agent-run && bun run build)
test -f frontend-agent-run/dist/index.html

# Binaries (also built inside the inspect orchestrator)
OUT="${TMPDIR:-/tmp}/message-card-reorder-followup-build"
mkdir -p "$OUT/bin"
go build -o "$OUT/bin/agent-run" ./cmd/agent-run
go build -o "$OUT/bin/llm-mock-run-grok" ./agent/llm/llm-mock/llm-mock-run-grok
go build -o /dev/null ./script/debug/message-card-reorder-followup
test -x "$OUT/bin/agent-run" && test -x "$OUT/bin/llm-mock-run-grok"
```

**Verify:** all `go build` / `bun run build` exit 0; `frontend-agent-run/dist/index.html` exists.

---

## Step 2 — Deploy / Update

Local only. Confirm tools and that the orchestrator can start web:

```sh
which go bun playwright-debug
playwright-debug -e 'console.log("pw-ok")'
test -f script/debug/message-card-reorder-followup/probe.js
test -f script/debug/message-card-reorder-followup/main.go
```

**Verify:** `playwright-debug` prints `pw-ok`; probe paths exist.

(Web server start is owned by step 3/4 orchestrator so the port, `AGENT_RUN_HOME`,
and mock hook stay isolated.)

---

## Step 3 — Run (trigger failure only)

Do **not** edit product merge/timeline code.

**Preferred (orchestrated — build + web + playwright):**

```sh
cd "$(git rev-parse --show-toplevel)"
go run ./script/debug/message-card-reorder-followup --expect=repro
# default: free port, mock grok-tty, open API
```

**Manual equivalent (if debugging probe alone):**

```sh
# After step 1 binaries + frontend build:
export AGENT_RUN_HOME=/tmp/reorder-loop-home
export GROK_HOME=/tmp/reorder-loop-grok
mkdir -p "$AGENT_RUN_HOME" "$GROK_HOME"
# start agent-run web --port <PORT> --no-open --agent-runner grok-tty \
#   --grok-home "$GROK_HOME" --grok-tty-runner-binary <mock> \
#   with LLM_MOCK_RUN_GROK_COMMAND multi-turn hook (see main.go)
playwright-debug run script/debug/message-card-reorder-followup/probe.js \
  "http://127.0.0.1:<PORT>" /tmp/reorder-probe-out
```

**Verify:** step 3 alone may exit non-zero — that is expected for unfixed product.
Do not “fix” here.

---

## Step 4 — Inspect / Feedback (symptom present)

```sh
cd "$(git rev-parse --show-toplevel)"
go run ./script/debug/message-card-reorder-followup --expect=repro
```

**Verify (REPRO / RED):**

| Signal | Expected |
|--------|----------|
| Exit code | **1** |
| Stdout | contains `REPRO: live message-card reorder` |
| Stdout | contains a second `REPRO:` reason line about adjacent users or assistant-before-user |
| Artifacts | `probe-report.json` under the printed `out=` dir; screenshots `01`–`05` |

If exit **2** and “symptom not observed”, the loop is not established — fix the
probe/hook (not product) until RED is reliable.

### Step 4b — Inspect verify (post-fix only; not required for establishment)

```sh
go run ./script/debug/message-card-reorder-followup --expect=healthy
```

**Verify (GREEN after fix):** exit **0**, line `VERIFY: follow-up keeps correct message order`.

---

## Step 5 — Fix (only after REPRO PASS)

Likely product touch points (do **not** apply during establishment dry-run):

- `frontend-agent-run/src/pages/SessionPage.tsx` — grok-tty refresh path that
  strips **all** server user events while `lastSentUserPromptRef` is set, then
  merges optimistic users.
- `frontend-agent-run/src/lib/timeline.ts` — `mergeSessionEvents` appends
  missing optimistic users at the **end** of the server list (causes jump).

Sound fix direction (hypothesis for `/run-the-loop`, not applied here):

1. Stop stripping all user events on follow-up refresh; only suppress the
   in-flight duplicate of the last sent prompt until server confirms it.
2. Or merge optimistic users **in chronological place** (by timestamp / relative
   order) instead of always appending after server events.
3. Keep reload path green; ensure **live** path matches (healthy mode checks live).

After a candidate fix, return to step 1, then step 4b `--expect=healthy`.

---

## Dry-run log

| Time (UTC) | Step | Result | Evidence |
|------------|------|--------|----------|
| 2026-07-12T22:00Z | 1 Build | PASS | `bun run build` frontend-agent-run; `go build` agent-run + llm-mock-run-grok + inspect package |
| 2026-07-12T22:00Z | 2 Deploy tools | PASS | `playwright-debug` on PATH (`pw-ok`); probe.js + main.go present |
| 2026-07-12T22:00–22:02Z | 3 Run | PASS (trigger) | Orchestrator started open-API web + mock grok-tty; Playwright home→grok-tty→`run ls`→`what did I say` |
| 2026-07-12T22:02Z | 4 Inspect REPRO | **REPRO PASS** | exit **1**; `REPRO: live message-card reorder on grok-tty follow-up`; `REPRO: "run ls" immediately above "what did I say" with no assistant between (live reorder)`; live users=`[run ls, what did I say]` |
| 2026-07-12T22:07Z | 4 re-run (fresh home) | REPRO PASS | dry3 `/tmp/message-card-reorder-followup-dry3` same REPRO lines; screenshots 01–05 |
| 2026-07-12T22:10–22:12Z | 4 re-run (dry5) | REPRO PASS | dry5 exit 1; same adjacent-user REPRO; confirms reliability with fresh `homes/run-*` |

**Establishment dry-run: REPRO PASS** (pre-fix).  
**Post-fix: VERIFY PASS** (`--expect=healthy` exit 0, 2026-07-12T23:16Z).

Notes from dry-run (pre-fix):

- Live timeline often showed assistant fragment **before** both user cards, or the two user cards **adjacent** — client merge stripped **all** server users then re-appended optimistic ones.
- Mock grok bind can race on turn 1 (`Cannot resolve session id`); user bubbles still appear.
- Never pre-create **empty** `updates.jsonl` in the mock hook (discovery hang; dry4 false negative).

### Fix (applied)

- `SessionPage` refresh: strip only the **pending** `lastSentUserPromptRef` user event, not all users.
- SSE: only suppress the matching pending user event (do not drop other history users).
- Debug: `refresh:strip-pending-user-only`, `refresh:merge-result`, `sse:grok-pending-user-confirmed`.
- Verify: `go run ./script/debug/message-card-reorder-followup --expect=healthy` →  
  live order `user:run ls` → `assistant` → `user:what did I say`.

---

## Cleanup

Orchestrator SIGTERMs the web process it started. Manual leftovers:

```sh
# only if you started web by hand
# kill the agent-run web PID for that port
rm -rf /tmp/message-card-reorder-followup-*
```
