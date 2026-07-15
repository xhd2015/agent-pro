---
name: verify-on-behalf-of-user/sandbox
description: >-
  Sandbox paths, enter-sandbox.sh setup, forbidden writes, and optional strict
  sandbox-exec mode for verify-on-behalf-of-user.
---

# Sandbox

Verification runs in a **persistent sandbox** that does not touch real user data.

## Paths

| Path | Purpose |
|------|---------|
| `~/.sandbox/default-home` | `HOME` during verify (`.tsk`, `.config`, etc.) |
| `~/.sandbox/bin` | Built binaries — **never** `~/go/bin` |
| `~/.sandbox/transcripts/` | Markdown transcripts (primary deliverable) |

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

## Forbidden during verify

- Writing to real `~/.tsk`, `~/.config`, or `~/go/bin`
- `go install` without `-o "$SANDBOX_BIN/..."`
- Smoke commands without sourcing `enter-sandbox.sh` first

## Optional strict mode

Set `VERIFY_SANDBOX=strict` and use `scripts/sandbox-verify.sb` with `sandbox-exec`
(OS-level write restrictions). Default is env-only sandbox.