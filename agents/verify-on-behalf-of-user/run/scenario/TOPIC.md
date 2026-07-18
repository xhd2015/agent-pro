---
name: verify-on-behalf-of-user/scenario
description: >-
  Depth labels, surface gate, real integration scenarios, browser-agent UI rules,
  and FAIL when the UI path is unavailable.
---

# Scenario

Human-shaped verify is **scenario-first**: stand up what a user would run, then
exercise a real journey. Smoke is allowed only as a **labeled** downgrade.

## Depth (always label)

Every transcript header and final reply must include:

```text
Depth: smoke | scenario | full
Reason: <one line>
```

| Depth | Meaning |
|-------|---------|
| `smoke` | Minimal commands that still exercise the claim (e.g. pure CLI). Labeled downgrade only. |
| `scenario` | **Default.** Real bring-up + at least one user journey with observables. |
| `full` | Multiple scenarios and/or full recipe matrix + teardown scrutiny. |

Never omit the depth label. If using smoke, the reason must explain why scenario
is unnecessary (typically CLI-only, no server/UI).

## Surface gate

Classify before running:

| Surface | Runtime bring-up | UI tool | Min depth |
|---------|------------------|---------|-----------|
| CLI-only | No | — | smoke OK if labeled |
| Interactive CLI / TTY | Yes (as needed) | **tty-watch** (see `tty` topic) | scenario |
| Server / HTTP | Yes | optional API checks | scenario |
| Frontend / UI | Yes | **browser-agent** | scenario |
| Session lifecycle | Yes | browser-agent if UI; **tty-watch** if interactive shell | scenario |
| Multi-binary | Yes | as needed | scenario |

If server **or** frontend/UI **or** session lifecycle is in scope → depth ≥ **scenario**.

## Interactive CLI / TTY

When the claim needs a **real TTY** (not pipe-only):

- Follow topic **`tty`**: `agent-pro skill --show verify-on-behalf-of-user/tty`
- **Start** with `tty-watch run --detach` (non-blocking)
- Drive input with `tty-watch send`; assert via `tty-watch snapshot`
- **Always** end with `tty-watch kill` to reclaim registry/PTY resources
- Missing `tty-watch` when TTY is required → **FAIL** (not skip)

Do not treat pipe/`SkipTTYCheck` library Attach alone as proof of interactive CLI.

## What a scenario is

A scenario is a short user story with **observables**, not a help dump:

```text
S1 — happy path
  Given: sandbox build + server ready at http://127.0.0.1:<port>
  When:  user opens session page / runs claim-related CLI
  Then:  expected status/DOM/text/exit codes hold
```

Prefer `docs/verify-recipes/<feature>.md` steps when present.

## UI: browser-agent only

For any frontend/UI surface:

1. Ensure `browser-agent` is on `PATH`
2. Bootstrap: `browser-agent session new` (do **not** run `browser-agent serve` as agent bootstrap)
3. Export `BROWSER_AGENT_SESSION_ID`
4. Open the app (create-tab / navigate) to the real local URL
5. Assert via `session eval`, screenshots, logs as needed
6. Keep the session control page (`/go?session=`) open

**Anti-patterns for this skill:**

```text
# WRONG — headless doctest tooling, not human-shaped verify
playwright-debug ...
npx playwright ...

# WRONG — blocks; session new already ensures the daemon
browser-agent serve
```

`playwright-debug` and similar headless runners stay in **doctests/CI**. This skill
replaces them with a live **browser-agent** path for pre-commit human review.

## FAIL when UI path is broken

If surface requires UI and any of the following holds, verdict is **FAIL**:

- `browser-agent` missing from PATH
- `session new` fails
- Chrome/extension not connected / no capturable tab
- Cannot open or assert on the app URL

Do **not** skip UI, do **not** mark PASS with CLI-only evidence, do **not** use
BLOCKED as a soft pass. Record the failure details in the transcript and still
write + inline the partial transcript.

## Smoke as labeled downgrade

Allowed when:

- Surface is CLI-only, and
- Depth is explicitly `smoke` with reason, and
- Commands still exercise the claimed behavior (not only `-h` unless the claim is help text)

## Teardown hygiene

Record and run teardown after scenarios:

- Kill leftover server/serve PIDs from this verify run
- Leave sandbox home for inspection (`~/.sandbox/default-home`)
- Note screenshot and log paths under the transcript or `/tmp/verify-<slug>/`
