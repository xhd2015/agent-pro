---
title: "llm-mock: add /alpha/generate endpoint for Command Code sandbox mode"
created: 2026-07-14
slug: llm-mock-alpha-generate
path: doc/LOOP_2026-07-14_llm-mock-alpha-generate.md
loop_kind: bug-repro
dry_run_status: "REPRO PASS"
---

# LOOP: llm-mock-alpha-generate

## Kind

**bug-repro** — `cmd` in sandbox mode fails because `llm-mock` does not serve
the `/alpha/*` endpoints Command Code requires.

## Symptom

When `cmd` runs in sandbox mode against the existing `llm-mock` server:

```
COMMANDCODE_SANDBOX=true \
COMMANDCODE_API_URL=http://127.0.0.1:$PORT \
COMMAND_CODE_API_KEY=sk-mock \
cmd -p "hello" --yolo --model claude-sonnet-5
```

**Actual sequence** (4 requests observed):

1. `GET /alpha/whoami` → 404 (mock doesn't serve it)
2. `POST /alpha/lifecycle-events` → 404
3. `POST /alpha/fingerprint/record` → 404
4. `POST /alpha/generate` → 404 (the actual LLM call)

**Expected error (current, symptom):**
`Error: Body is unusable: Body has already been read`

The mock server returns `Content-Type: text/plain; charset=utf-8` with "404 page
not found" for unknown paths, which `cmd` can't parse.

**Root cause:** `llm-mock/main.go` only registers these handlers:
- `/v1/chat/completions`
- `/v1/responses`
- `/v1/messages`
- `/v1/models`
- `/admin/requests`

None of the `/alpha/*` endpoints exist.

**Repro preconditions:**
1. `cmd` installed and functional (`which cmd`, `cmd --version` → 0.44.1).
2. `llm-mock` binary built from `agent/llm/llm-mock/`.
3. `COMMANDCODE_SANDBOX=true` + `COMMANDCODE_API_URL` pointing at mock.

## Goal

Steps 1–4 reliably reproduce:
1. mock `/alpha/*` endpoints missing → `cmd -p "hello"` fails.
2. After adding `/alpha/generate`, `cmd` connects, sends its request, and gets
   a parsable response — exit 0.

## Endpoint Reference

From the binary capture, Command Code's `/alpha/generate` request shape:

```json
{
  "config": { "workingDir": "...", "date": "...", "environment": "...",
    "structure": [...], "isGitRepo": true, "currentBranch": "...",
    "mainBranch": "...", "gitStatus": "...", "recentCommits": [...] },
  "memory": "...",
  "taste": null,
  "skills": "<available_skills>...</available_skills>",
  "params": {
    "tools": [{ "name": "...", "description": "...", "input_schema": {...} }, ...],
    "stream": true,
    "max_tokens": 64000,
    "temperature": 0.3,
    "messages": [{ "role": "user", "content": "hello" }],
    "model": "claude-sonnet-5"
  },
  "threadId": "uuid"
}
```

Response is SSE (`Content-Type: text/event-stream`). Each chunk:

```
data: <json>\n\n
```

Final event: `data: [DONE]\n\n`

Minimal successful response JSON shape (per turning-gears events):

```json
{"type": "assistant", "message": {"id": "msg_...", "type": "message",
  "role": "assistant", "content": [{"type": "text", "text": "Hello from mock!"}],
  "model": "claude-sonnet-5", "stop_reason": "end_turn",
  "stop_sequence": null, "usage": {"input_tokens": 10, "output_tokens": 5}}}
```

## Steps

### 1. Build

Build the `llm-mock` server (no `/alpha/*` endpoints yet):

```sh
go build -o /tmp/llm-mock ./agent/llm/llm-mock/
```

**Verify:** binary exists

```sh
test -x /tmp/llm-mock && echo "OK"
```

### 2. Deploy / Update

No deploy. The mock server runs locally.

### 3. Run

Start mock, then run `cmd` against it:

```sh
# Start llm-mock on port 19980 (no /alpha handlers)
echo '{"port":19980,"exchanges":[]}' > /tmp/mock-config.json
/tmp/llm-mock --config /tmp/mock-config.json &
sleep 1

# Run cmd in sandbox mode pointing at mock
COMMANDCODE_SANDBOX=true \
  COMMANDCODE_API_URL=http://127.0.0.1:19980 \
  COMMAND_CODE_API_KEY=sk-mock \
  cmd -p "hello" --yolo --model claude-sonnet-5 2>&1
RC=$?
echo "CMD_EXIT=$RC"
```

**Expected (symptom):** Non-zero exit code. Error message contains something
about the response being unparseable (404, body already read, etc.).

### 4. Inspect

```sh
# cmd exits non-zero due to 404 on /alpha/whoami (mock doesn't record unhandled paths)
echo "CMD_EXIT was $RC (expected non-zero)"
# The stderr from step 3 should contain an error message
echo "$STDERR_OUTPUT" | grep -q "Body is unusable\|Error:" && echo "REPRO: symptom confirmed" || echo "CHECK: symptom not matched"
```

**Expected (REPRO):** Step 3 exits non-zero. Stderr contains `Error:` and
`Body is unusable` (cmd can't parse the mock's 404 HTML response).
Note: `admin/requests` will be `[]` because unhandled paths (like `/alpha/*`)
don't reach the mock's request-recording logic — the 404 is served by Go's
default mux.

### 5. Fix

1. Add handlers in `llm-mock/main.go` for `/alpha/whoami` (GET, returns `{success:true,user:{...}}`),
   `/alpha/lifecycle-events` (POST no-op), `/alpha/fingerprint/record` (POST no-op),
   and `/alpha/generate` (POST — core handler).
2. `/alpha/generate` extracts `params.messages` → converts to chat format → reuses
   `findMatch()`. Streams **newline-delimited JSON** (`Content-Type: text/plain`)
   with events: `text-delta`, `tool-use`, `tool-delta`, `finish`.
3. Fallback when no match: emits auto-generated text via `findGeneratedMatch`.

**Post-fix expected:** `cmd -p "hello" --yolo` exits 0 and prints mock response.

Return to step 1 after each fix iteration.

## Pitfalls & blockers

| Pitfall | Mitigation |
|---------|------------|
| Port already in use | Use `lsof -ti :19980 \| xargs kill` before start |
| `cmd` reads auth from `~/.commandcode/auth.json` | `COMMAND_CODE_API_KEY=sk-mock` overrides it |
| `/alpha/generate` uses NDJSON, not SSE | Stream as `text/plain` with `\n`-delimited JSON lines |
| `/alpha/whoami` shape matters | Must return `{success:true, user:{id,name,email,userName}}` |

### Fix applied (2026-07-14)

**Files changed:**
- `agent/llm/llm-mock/main.go` — Added 4 alpha endpoint handlers + NDJSON streaming
- `agent/llm/anthropic/messages.go` — No effective change (reverted unused helper)

**Format:** `/alpha/generate` streams newline-delimited JSON (`Content-Type: text/plain; charset=utf-8`).
Event types: `text-delta`, `tool-use`, `tool-delta`, `finish`. Reuses `findMatch()`.
Fallback uses `findGeneratedMatch()`.

## Dry-run log

| Step | Timestamp | Result | Evidence |
|------|-----------|--------|----------|
| 1 Build | 07:56 | OK | `/tmp/llm-mock` 832-line Go binary, compiles from `agent/llm-mock/` |
| 2 Deploy | 07:56 | OK | No deploy needed; local binary |
| 3 Run | 07:57 | REPRO | `CMD_EXIT=1`, stderr: `Error: Body is unusable: Body has already been read` |
| 4 Inspect | 07:57 | REPRO | admin/requests `[]` (unhandled paths not recorded); exit code non-zero confirms failure |
