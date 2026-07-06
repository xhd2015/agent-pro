# Phase 6: Test Migration (Deferred)

**Status:** deferred  
**Depends on:** Phase 5  
**Blocks:** —

## Objective

Update agent-run doctest trees to match the new ttywatch-based behavior. Explicitly deferred until Phases 0–5 land.

## Scope

### Doctest directories to update

| Directory | Focus |
|-----------|-------|
| `cmd/agent-run/tests/tty/` | attach, send, snapshot, watch, status, help |
| `cmd/agent-run/tests/ttyrunner/` | registry, storage, resolution → migrate to agenttty |
| `cmd/agent-run/tests/grok-tty/` | headless run, attach, send |
| `cmd/agent-run/tests/codex-tty/` | headless run, attach, send |
| `cmd/agent-run/tests/web/` | terminal proxy, follow-up send, persistent lifecycle |
| `cmd/agent-run/tests/run/` | TTY runner integration tests |

### New tests to add

| Test | Assertion |
|------|-----------|
| `cmd/agent-run/tests/tty/snapshot/` | Sanitized scrollback printed |
| `cmd/agent-run/tests/tty/watch/` | Readonly stream, no stdin |
| `cmd/agent-run/tests/tty/send/injects-conditional-cr/` | `\r` appended only when missing |
| `cmd/agent-run/tests/tty/send/no-ctrl-u/` | No `\x15` in injected bytes |

### Send semantics test cases

| Input message | Expected injected bytes |
|---------------|---------------------------|
| `"follow up"` | `"follow up\r"` |
| `"follow up\r"` | `"follow up\r"` |
| `"a\nb"` | `"a\nb\r"` |
| `"done\r\n"` | `"done\r\n"` (no extra `\r`) |

### Behavioral changes tests must reflect

| Area | Old expectation | New expectation |
|------|-----------------|-----------------|
| send | Ctrl-U + trim + `\r` | verbatim + conditional `\r` |
| run | in-process registry PID = agent-run PID | registry PID = serve child PID |
| attach | ptyclient basic | ttywatch attach with detach |
| run stdout | agent events | agent events only (no change, but explicit) |
| web follow-up | `\x15` prefix | `SendMessage(suffixCR=true)` |

## Harness updates

`cmd/agent-run/tests/tty/SETUP.md` and `cmd/agent-run/tests/ttyrunner/SETUP.md` probe helpers likely need:

- `InjectedBytes` capture via fake ptywrap server HTTP inject endpoint (not WS)
- `prepare-inject` endpoint on fake server (match ttywatch server)
- Registry under `AGENT_RUN_HOME/{runner}-registry/` via ttywatch format

Consider reusing patterns from `script/tty-watch/tests/send/injects-verbatim/`.

## Execution strategy

1. Run existing suites to collect failures after Phase 5
2. Update shared SETUP.md probe helpers first
3. Fix send tests (largest semantic change)
4. Add snapshot/watch tests
5. Fix web terminal tests
6. Remove obsolete ttyrunner-specific tests or repoint to agenttty

## Acceptance Criteria

- [ ] `doctest test ./cmd/agent-run/tests/tty/...` passes
- [ ] `doctest test ./cmd/agent-run/tests/grok-tty/...` passes
- [ ] `doctest test ./cmd/agent-run/tests/codex-tty/...` passes
- [ ] `doctest test ./cmd/agent-run/tests/web/...` (terminal-related) passes
- [ ] New snapshot/watch/send semantics tests added
- [ ] Obsolete tests for deleted packages removed

## Subagent Prompt Template

```
Implement Phase 6 of agent-run TTY migration per:
  cmd/agent-run/ttywatch-migration/PHASE_6_TEST_MIGRATION.md

Update agent-run doctests for ttywatch-based behavior. Add snapshot/watch tests.
Fix send tests for conditional \r (no Ctrl-U). Depends on Phase 5 complete.
```