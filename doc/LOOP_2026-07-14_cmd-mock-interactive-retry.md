---
title: "llm-mock-run-commandcode: interactive cmd shows Connection Issue retry loop"
created: 2026-07-14
slug: cmd-mock-interactive-retry
path: doc/LOOP_2026-07-14_cmd-mock-interactive-retry.md
loop_kind: bug-repro
dry_run_status: "REPRO PASS"
---

# LOOP: cmd-mock-interactive-retry

## Kind

**bug-repro** — interactive `cmd` inside `llm-mock-run-commandcode` (no `-p` flag)
produces "Connection Issue. Retrying (attempt N)" in the PTY scrollback instead of
a response.

## Symptom

Interactive `cmd` run inside the PTY and sent "hello" via tty-watch send produces:

```
❯ hello
Connection Issue. Retrying (attempt 7)
```

The `-p` (print) mode works fine — `cmd -p "hello"` returns the mock response.
Only interactive mode is broken.

**Root cause hypothesis:** `cmd` in interactive mode sends content as an **array** of content blocks (`"content":[{"text":"hello","type":"text"}]`), but the mock's `alphaMessage.Content` field is typed `string` — JSON unmarshaling fails.

**Root cause confirmed** via HTTP logging. Index 6 in `/tmp/mock-http.jsonl` shows:
```json
"messages":[{"content":[{"text":"hello","type":"text"}],"role":"user"}]
```
vs print mode which sends `"content":"hello"` (string).

**Fix (2026-07-14):** Changed `alphaMessage.Content` from `string` to `json.RawMessage`. Added `extractAlphaContentText` helper that tries string first, then array of `{"text":"..."}` blocks. Applied to both `server/server.go` and `main.go`.

**Repro preconditions:**
1. `tty-watch` binary built from `./script/tty-watch`
2. `llm-mock-run-commandcode` binary built from `./agent/llm/llm-mock/llm-mock-run-commandcode`
3. `cmd` CLI installed and on PATH

## Goal

Steps 1–4 reproduce the symptom: interactive cmd shows "Connection Issue. Retrying".

## Steps

### 1. Build

```sh
go build -o /tmp/tty-watch ./script/tty-watch
go build -o /tmp/llm-mock-run-commandcode ./agent/llm/llm-mock/llm-mock-run-commandcode
```

**Verify:**

```sh
test -x /tmp/tty-watch && echo "tty-watch: OK"
test -x /tmp/llm-mock-run-commandcode && echo "llm-mock-run-commandcode: OK"
```

### 2. Deploy / Update

No deploy. Local binaries in `/tmp/`.

### 3. Run

Start `llm-mock-run-commandcode` in interactive mode inside a PTY, then send hello:

```sh
# Start interactive cmd in PTY (no -p, no positional prompt = interactive mode)
/tmp/tty-watch run --detach --session-id=repro-cmd /tmp/llm-mock-run-commandcode

# Wait for cmd TUI to appear
sleep 5

# Send "hello" and Enter
/tmp/tty-watch send repro-cmd "hello"
sleep 1
/tmp/tty-watch send repro-cmd $'\r'

# Wait for cmd to process
sleep 15
```

### 4. Inspect

```sh
SNAPSHOT=$(/tmp/tty-watch snapshot repro-cmd 2>&1)
echo "$SNAPSHOT"

# Assert symptom: scrollback contains "Connection Issue" or "Retrying"
echo "$SNAPSHOT" | grep -q -E "Connection Issue|Retrying" \
  && echo "REPRO: symptom confirmed" \
  || echo "CHECK: symptom not matched"
```

**Expected (REPRO):** PTY scrollback contains "Connection Issue. Retrying (attempt N)".

### 5. Fix

**Diagnosis tool:** `script/debug/cmd-mock-inspect/main.go`

Uses `ttywatch.EphemeralSession` to run `llm-mock-run-commandcode -p "hello"` in
a PTY, poll for scrollback, and print the raw PTY output. Confirms `-p` mode works.

**To fix interactive mode, investigate:**

1. **`/alpha/whoami` response shape** — compare what the mock returns vs what real
   Command Code returns. The mock returns:
   ```json
   {"success":true,"user":{"id":"mock-user-id","name":"mock-user","email":"mock@example.com","userName":"mock-user"},"org":null}
   ```

2. **`/alpha/generate` request format in interactive mode** — enable mock HTTP logging
   (`LLM_MOCK_RUN_COMMANDCODE_DEBUG=1`) and capture the raw HTTP requests to
   compare print vs interactive mode requests.

3. **Environment variable propagation** — verify `COMMANDCODE_SANDBOX=true` and
   `COMMANDCODE_API_URL=http://127.0.0.1:$PORT` are visible to cmd's child
   processes in interactive mode.

4. **Request routing difference** — interactive cmd may call different endpoints
   or use different header/body formats than print mode.

Return to step 1 after each fix iteration.

## Pitfalls & blockers

| Pitfall | Mitigation |
|---------|------------|
| `--detach` mode kills PTY on child exit | cmd interactive runs forever; PTY stays alive |
| `tty-watch send` requires raw newlines | Use `$'\r'` for Enter key |
| Mock HTTP logs need explicit flag | Use `LLM_MOCK_RUN_FLAGS="--log-http /tmp/mock-http.jsonl"` |
| cmd may read credentials from real `~/.commandcode` | `llm-mock-run-commandcode` sets `HOME=` to isolated temp dir |
| Port conflicts between concurrent mock instances | Clean up orphans with `pkill -9 -f "__serve__\|llm-mock"` before each run |

## Dry-run log

| Step | Timestamp | Result | Evidence |
|------|-----------|--------|----------|
| 1 Build | 12:15 | OK | `/tmp/tty-watch` and `/tmp/llm-mock-run-commandcode` built |
| 2 Deploy | 12:15 | OK | No deploy needed; local binaries in `/tmp/` |
| 3 Run | 12:15 | OK | Session `dryrun-repro` started; sent "hello" + Enter via `tty-watch send` |
| 4 Inspect | 12:16 | **REPRO PASS** | `grep -qE "Connection Issue\|Retrying"` matched; scrollback: `Connection Issue. Retrying (attempt 7)` |
