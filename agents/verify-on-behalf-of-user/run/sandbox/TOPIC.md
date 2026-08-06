---
name: verify-on-behalf-of-user/sandbox
description: >-
  Default sandbox mode: paths, enter-sandbox.sh setup, forbidden writes, and
  optional strict sandbox-exec mode for verify-on-behalf-of-user. For opt-in
  outside-sandbox runs, see topic host.
---

# Sandbox

**Mode id:** `sandbox` — this is the **default** and only automatic mode.

Verification runs in a **persistent sandbox** that does not touch real user data.
For opt-in real-env verification, see topic **`host`** (only when the user
explicitly asks for host / outside sandbox).

## Paths

| Path | Purpose |
|------|---------|
| `~/.sandbox/default-home` | `HOME` during verify (`.tsk`, `.config`, etc.) |
| `~/.sandbox/bin` | Built binaries — **never** `~/go/bin` |
| `~/.sandbox/transcripts/` | Markdown transcripts (primary on-disk deliverable; also used when Mode is host) |

## enter-sandbox.sh

After `agent-pro skill --install verify-on-behalf-of-user`:

```sh
VERIFY_SKILL_ROOT=".agents/skills/verify-on-behalf-of-user"
source "${VERIFY_SKILL_ROOT}/scripts/enter-sandbox.sh"
```

From an agent-pro checkout (before install):

```sh
source agents/verify-on-behalf-of-user/run/scripts/enter-sandbox.sh
```

Default: **reset** sandbox data dirs (`.tsk`, `.config`) at start. Pass `--no-reset`
to keep prior sandbox state.

## Forbidden during verify (sandbox mode)

- Writing to real `~/.tsk`, `~/.config`, or `~/go/bin`
- `go install` without `-o "$SANDBOX_BIN/..."` (or equivalent sandbox-only output)
- Scenario or smoke commands without sourcing `enter-sandbox.sh` first
- Switching to host install paths without an explicit user opt-in (see `host`)

## Optional strict mode

Set `VERIFY_SANDBOX=strict` and use `scripts/sandbox-verify.sb` with `sandbox-exec`
(OS-level write restrictions). Default is env-only sandbox.
