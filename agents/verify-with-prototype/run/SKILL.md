---
name: verify-with-prototype
description: >-
  Prototype-before-production workflow for risky host integrations (sudo,
  privileges, system config). Use when a feature touches the real OS, when
  behavior must be proven with a small script before full implementation, or when
  the user mentions prototype verification, POC script, or verify-with-prototype.
  Run after investigate/followup and before doctest-TDD integration.
---

You guide the agent through **prototype verification** before full implementation
and integration. The goal is to learn failure modes cheaply (false positives,
session caches, TTY requirements) in a throwaway script, then encode lessons in
a real package with fakes-based tests.

# When to use

- Host or privilege integration: `sudo`, `/etc/sudoers.d`, TUN, `networksetup`, etc.
- Persistent state vs session cache can be confused (`sudo -n`, timestamp cache)
- Production mistakes are hard to undo or dangerous to test in CI
- User wants "setup once, then never password again" or similar UX proofs

# When not to use

- Pure documentation or config-only changes
- Behavior already covered by isolated unit tests with no OS side effects
- User only wants root-cause explanation → use `investigate` instead

# Required workflow

## Phase 0 — Understand (often already done)

- Symptom, expected behavior, and constraints from `investigate` / `followup`
- Identify the **stand-in** for the expensive dependency (e.g. `hello.sh` instead of `sing-box`)

## Phase 1 — POC script

Create `script/<feature>-poc/` (Go preferred) with **small command surface**:

| Command | Purpose |
|---------|---------|
| `detect` | Report persistent state vs cache vs "works right now" **separately** |
| `auto-setup` | One-time install path (may prompt on TTY) |
| `remove-*` | Tear down + invalidate session cache if relevant |

Document a **manual test matrix** in the script header:

```
remove → auto-setup (password once) → auto-setup (no password) → remove → auto-setup (password again)
```

**POC rules**

- Do not infer persistent install from `sudo -n` alone — use drop-in file + local manifest (or equivalent)
- On remove, flush sudo timestamp cache (`sudo -k`) when testing "prompt again" behavior
- Distinguish **stdin** vs **stdout** TTY when password is required

## Phase 2 — Manual verification

User (or agent with user) runs the matrix in a real terminal. Capture lessons:

- False positives (e.g. warm cache after `remove-sudo`)
- Missing non-TTY guard
- Wrong detection signal

Fix the POC until the matrix passes. **Do not** proceed to production until manual proof succeeds.

## Phase 3 — Requirement doc

Write `REQUIREMENT-DESIGN-<slug>.md` with:

- Data model (`Config`, `Rule`, manifest paths, sudoers drop-in name)
- Approved scope (package only vs integration)
- Test tree outline (doctests with injectable `FS` + `Runner`, no real `/etc` or `sudo`)

Get explicit user approval before TDD implementation.

## Phase 4 — Extract package + TDD

- Extract core logic to `pkgs/<name>/` with injectable `FS` and `Runner`
- Doctest tree with **fakes only** — no real system modification in CI
- Thin CLIs: POC script + production caller(s) share the package
- Guards:
  - Under `go test`, panic if `FS`/`Runner` nil (production deps would touch the OS)
  - Before interactive `sudo`, error if stdin is not a TTY (already-installed path may skip)

## Phase 5 — Integration

- Wire production entrypoints (flags like `--no-setup-sudo` for opt-out of auto-setup)
- Hook layer in existing packages for test injection
- Bump / `replace` module deps as needed

# Response shape

1. **Risk summary** — what can go wrong on the real system
2. **POC plan** — commands, stand-in, manual matrix
3. **Detection model** — persistent vs cache vs live probe (table)
4. **After manual proof** — package API sketch + test groups
5. **Stop** — do not implement production code until POC matrix passes and user approves REQUIREMENT-DESIGN

# Case study (this repo)

VPN passwordless sudo (`master-2026-07-06-vpn-no-password` work):

- POC: `script/sudo-nopasswd-poc` (`detect`, `auto-setup`, `remove-sudo`)
- Lesson: `sudo -n` after `remove-sudo` succeeded via **timestamp cache**, not NOPASSWD
- Fix: manifest + `/etc/sudoers.d` stat for `IsInstalled`; `sudo -k` on remove
- Package: `pkgs/sudosetup` with fake FS/Runner doctests
- Integration: `ws-proxy vpn` auto-setup + `--no-setup-sudo`
- Guards: test-time production-deps panic; stdin-TTY check before first setup

# Related commands

```sh
agent-pro skill investigate --show
agent-pro skill followup --show
agent-pro skill verify-with-prototype --show
```

Install locally:

```sh
agent-pro skill verify-with-prototype install
```