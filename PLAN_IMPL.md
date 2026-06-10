# Doctest Implementation Plan

This plan builds the doctest CLI and its integration test foundation in phases.
The work should proceed phase by phase, keeping old command entrypoints working
while moving reusable implementation into importable packages under
`agents/doctest/libdoc`.

## Phase 1: Scriptable fake-codex

Goal: make `cmd/fake-codex` return deterministic, expected response sequences
for integration tests.

Implementation:

- Add a scripted mode to `fake-codex exec`.
- Support `--script <path>` and `FAKE_CODEX_SCRIPT=<path>`.
- Keep the existing random generator as the fallback when no script is given.
- Reuse `pkgs/fake-agent.Event` for script events so fake output matches the
  existing Codex JSON event format.
- Allow scripts to define process exit behavior for error-path tests.

Expected script shape:

```json
{
  "events": [
    {
      "type": "item.completed",
      "item": {
        "id": "m1",
        "type": "message",
        "text": "Done",
        "status": "completed"
      }
    }
  ],
  "exit_code": 0,
  "stderr": ""
}
```

Tests:

- Scripted success prints the exact scripted events.
- Scripted failure prints scripted events and exits nonzero.
- `--script` overrides `FAKE_CODEX_SCRIPT`.
- Invalid script files fail with clear errors.
- Existing seeded random behavior still works.

## Phase 2: Agent TUI supports Codex

Goal: make `agentui` run either OpenCode or Codex.

Implementation:

- Introduce an agent backend abstraction for the runtime.
- Preserve OpenCode as the default backend.
- Implement Codex backend using `agent/cli/codex`.
- Make `--agent-runner codex` work.
- Allow tests to point Codex at `cmd/fake-codex` through the existing CLI path
  configuration or environment.
- Preserve session logging, stream updates, question handling, and plain mode.

Tests:

- `--agent-runner codex` routes to the Codex backend.
- fake Codex is invoked when configured.
- backend errors surface through CLI/TUI state.
- OpenCode behavior remains compatible.

## Phase 3: TUI testability

Goal: make agent TUI flows assertable without fragile screenshots or manual
terminal interaction.

Implementation:

- Split orchestration state from terminal rendering where needed.
- Add a test driver for feeding input and observing state transitions.
- Support forcing interactive mode in tests.
- Keep production rendering behavior unchanged.

Assertable outcomes:

- pending questions are captured and displayed in state
- supplied answers are recorded
- backend resumes after answers
- final output is recorded
- errors and cancellation are deterministic

## Phase 4: Doc-style integration tests

Goal: create real executable doc-style integration tests under
`agents/doctest/tests`.

Coverage:

- top-level and subcommand help
- invalid arguments and exit codes
- `agent generate` using fake Codex
- `agent fill-code` using fake Codex and then build
- `validate`
- `build`
- `test`
- `skill --list`
- `skill doc-spec show/install`
- `skill code-spec show/install`
- TUI clarification flow

Assertions:

- stdout/stderr contain required text
- exit code matches expected behavior
- generated files exist and contain expected sections/code blocks
- fake Codex was used instead of a real backend
- fill-code runs before build
- TUI state transitions are observable

## Phase 5: Run and stabilize

Run targeted tests first, then the doctest integration suite.

```sh
go test ./cmd/fake-codex/...
go test ./pkgs/fake-agent/...
go test ./agent/agentui/...
go test ./agents/doctest/...
go test ./agents/test-case-tree-runner/...
go test ./agents/test-case-tree-design-expert/...
```

Once the `doctest` CLI exists:

```sh
doctest validate agents/doctest/tests
doctest build agents/doctest/tests
doctest test agents/doctest/tests
```
