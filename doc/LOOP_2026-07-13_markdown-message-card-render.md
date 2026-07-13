---
title: "Markdown render on thinking + response message cards"
created: 2026-07-13
slug: markdown-message-card-render
path: doc/LOOP_2026-07-13_markdown-message-card-render.md
loop_kind: bug-repro
dry_run_status: "VERIFY PASS"
---

# LOOP: markdown-message-card-render

## Kind

**bug-repro** — prove that thinking progress cards and assistant response cards
currently show **raw markdown as plain text** (literal `**`, ``` fences) instead
of rendered HTML (bold / code / pre). After a product fix, switch to
`--expect=healthy` for the “looks good” VERIFY gate.

## Symptom (verbatim class)

User reference session:

```text
http://127.0.0.1:8821/sessions/web_da6177509c41c080
```

Assistant payload already contains markdown, e.g.:

```text
**pwd:** `/Users/.../working`

**ls:**
```
0.ad_hoc
...
```
```

and:

```text
You said: **run ls and pwd**
```

**Observed (live DOM / screenshots):**

- `[data-testid="assistant-message"]` `innerHTML` is plain text (no `<strong>`,
  no `<pre>` / `<code>`).
- `textContent` still includes raw `**` and ``` fence markers.
- Thinking cards (`.progress-card` with label “Thinking”) are plain
  `pre-wrap` strings; markers in think text are not markdown structure.

Inspect evidence (orchestrator default / repro mode):

```text
REPRO: markdown not rendered on thinking/response message cards
REPRO: markdown not rendered on assistant response: no <strong>, no <pre>/<code>, literal **
```

### Preconditions (do not “fix” in steps 1–4)

- Product code as-is (do **not** add `react-markdown` / change
  `SessionPage` message-body rendering during steps 1–4).
- `go`, `bun`, `playwright-debug` on `PATH`.
- Run from repo root.
- Isolated `AGENT_RUN_HOME` with seeded markdown events (orchestrator).

### Optional live URL

To probe the user’s running server instead of the seeded local web:

```sh
go run ./script/debug/markdown-message-card-render \
  -url=http://127.0.0.1:8821 \
  -session=/sessions/web_da6177509c41c080
```

## Goal

Steps 1–4 **reproduce** raw-markdown-on-cards on command (**REPRO PASS**).
Step 5 is fix guidance only during establishment; iterate with `/loop-workflow`
until:

```sh
go run ./script/debug/markdown-message-card-render --expect=healthy
```

exits 0 (**VERIFY** — “looks good”).

### What “looks good” means (healthy criteria)

| Check | Pass when |
|-------|-----------|
| Bold | Assistant card has `<strong>` (or `<b>`) for `**…**` |
| Code | Assistant has `<pre>` and/or `<code>` for fences / inline |
| No raw markers | Assistant `textContent` does **not** contain `**` or ``` |
| Thinking | If think body still has markdown markers, body has `<strong>`/`<code>` |
| Artifacts | Screenshots under probe out dir for visual review |

## Surfaces

| Step | Surface |
|------|---------|
| Build | `bun run build` in `frontend-agent-run`; `go build` `agent-run` + inspect script |
| Deploy | Local: start `agent-run web` with seeded session `layout-md-render` |
| Run | Playwright opens `/sessions/layout-md-render` |
| Inspect | `script/debug/markdown-message-card-render` exit **1** + `REPRO:` lines |

## Pitfalls & blockers

| Issue | Notes |
|-------|-------|
| Real grok / auth | Seeded path uses open API + `fake-opencode` idle session; no token |
| Port busy | Orchestrator picks free port (`-port=0` default) |
| playwright-debug missing | Install / ensure `~/go/bin` on PATH |
| Live session missing | `-url` path requires session still on that server |
| Do not “fix” during REPRO | No frontend markdown renderer changes in steps 1–4 |
| Truncated think | Progress cards clamp height; healthy mode still requires structure in DOM |

---

## Step 1 — Build

```sh
cd "$(git rev-parse --show-toplevel)"

(cd frontend-agent-run && bun run build)
test -f frontend-agent-run/dist/index.html

OUT="${TMPDIR:-/tmp}/markdown-message-card-render-build"
mkdir -p "$OUT/bin"
go build -o "$OUT/bin/agent-run" ./cmd/agent-run
go build -o /dev/null ./script/debug/markdown-message-card-render
test -x "$OUT/bin/agent-run"
```

**Verify:** all builds exit 0; `frontend-agent-run/dist/index.html` exists.

---

## Step 2 — Deploy / Update

Local only. Confirm tools; orchestrator starts web with seed:

```sh
which go bun playwright-debug
# Orchestrator seeds:
#   $OUT/home/.agent-run/sessions/layout-md-render/{meta.json,events.jsonl}
# with think + user + assistant (bold + fenced ls).
```

**Verify:** `playwright-debug` and `bun` resolve on `PATH`.

---

## Step 3 — Run

```sh
cd "$(git rev-parse --show-toplevel)"

OUT="${TMPDIR:-/tmp}/markdown-message-card-render-$(date +%s)"
go run ./script/debug/markdown-message-card-render -out="$OUT"
# expected exit 1 in repro mode when symptom present
```

Or one-shot against live UI:

```sh
go run ./script/debug/markdown-message-card-render \
  -out="$OUT" \
  -url=http://127.0.0.1:8821 \
  -session=/sessions/web_da6177509c41c080
```

**Verify:** process exits **1** with `REPRO:` lines (not exit 2 infrastructure fail).
Screenshots land under `$OUT/probe/*.png`.

---

## Step 4 — Inspect / Feedback

Default mode already prints inspect output. Re-read report:

```sh
# after step 3
test -f "$OUT/probe-report.json"
grep -E 'REPRO:|symptomPresent' "$OUT/probe-report.json" /dev/null || true
python3 - <<'PY' "$OUT/probe-report.json"
import json,sys
r=json.load(open(sys.argv[1]))
print("symptomPresent", r.get("symptomPresent"))
print("okHealthy", r.get("okHealthy"))
print("issues", r.get("issues"))
print("reason", (r.get("symptom") or {}).get("reason"))
print("screenshots", r.get("screenshots"))
PY
ls -la "$OUT/probe"/*.png
```

**Verify (REPRO):** `symptomPresent=true`, exit code 1, screenshots exist.

### Step 4b — Inspect VERIFY (after fix only)

```sh
go run ./script/debug/markdown-message-card-render --expect=healthy -out="$OUT-verify"
```

**Verify (VERIFY):** exit 0, stdout contains `VERIFY: thinking + response cards render markdown (looks good)`.

---

## Step 5 — Fix (only after REPRO PASS; not during establishment)

When iterating with `/loop-workflow`:

1. Add markdown renderer in `frontend-agent-run` (e.g. `react-markdown` + `remark-gfm`).
2. Render **assistant** `.message-body` and **think** `.progress-card-body` via a shared `MarkdownBody`.
3. CSS for `.markdown-body` (`p`, `pre`, `code`, lists); drop `pre-wrap` on markdown hosts.
4. Re-run:

```sh
go run ./script/debug/markdown-message-card-render --expect=healthy
```

5. Optionally re-check live:

```sh
go run ./script/debug/markdown-message-card-render --expect=healthy \
  -url=http://127.0.0.1:8821 \
  -session=/sessions/web_da6177509c41c080
```

Return to step 1 after product changes (frontend rebuild is inside the orchestrator).

---

## Dry-run log

| Timestamp (UTC-ish local) | Step | Result | Evidence |
|-----------|------|--------|----------|
| 2026-07-13 ~07:57 | 1 Build | PASS | `bun run build` + `go build` agent-run + inspect package OK |
| 2026-07-13 ~07:57 | 2 Deploy | PASS | web `http://127.0.0.1:62101`, health ok, open API, seeded `layout-md-render` |
| 2026-07-13 ~07:57 | 3 Run | PASS | Playwright opened session; 5 screenshots under `/tmp/markdown-message-card-render-dryrun-1783900655/probe/` |
| 2026-07-13 ~07:57 | 4 Inspect | **REPRO** (exit 1) | `symptomPresent=true` — no `<strong>`/`<pre>`/`code`, literal `**` and ```; think card plain markers |
| 2026-07-13 ~08:04 | 4 (live) | **REPRO** (exit 1) | `-url=http://127.0.0.1:8821 -session=/sessions/web_da6177509c41c080` same symptom; screenshot shows raw `**pwd:**` / fence list |
| 2026-07-13 | Dry-run status | **REPRO PASS** | Seeded + live both show markdown-not-rendered |
| 2026-07-13 ~08:13 | 5 Fix + VERIFY | **VERIFY PASS** (exit 0) | `MarkdownBody` + `react-markdown`/`remark-gfm`; assistant bold/code/pre; think compact markdown; screenshots under `/tmp/markdown-message-card-render-verify-1783901785/probe/` |

---

## Cleanup

```sh
# orchestrator stops web on exit; remove temp outs if desired
rm -rf /tmp/markdown-message-card-render-*
```
