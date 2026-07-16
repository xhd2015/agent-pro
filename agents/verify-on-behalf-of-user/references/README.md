# Verify recipes (consumer repos)

This skill is generic. Per-project scenario steps live in **consumer repositories**:

```
<your-repo>/docs/verify-recipes/<feature>.md
```

## Recipe schema

Each recipe should include:

1. **Scope** — what claim the recipe covers
2. **Surface** — CLI-only | server | frontend/UI | session lifecycle | multi-binary
3. **Depth** — recommended `smoke` | `scenario` | `full` (agents always label the depth they actually use)
4. **Bring-up** — build targets, start commands, ready checks, env (`HOME` sandbox, ports)
5. **Scenarios** — named journeys (S1, S2, …) with steps and **observables**
6. **UI** — if frontend/UI: browser-agent steps (never playwright-debug in this skill)
7. **Teardown** — kill PIDs, temp dirs
8. **Expected** — exit codes, stdout/stderr, HTTP/DOM checks, on-disk artifacts

### Example skeleton

```markdown
# Verify recipe: my-feature

## Surface
server + frontend

## Depth
scenario

## Bring-up
- go build -o "$SANDBOX_BIN/..." ...
- start server; wait for http://127.0.0.1:<port>/health

## Scenarios
### S1 — happy path
Steps: ...
Observables: ...

## UI (browser-agent)
- browser-agent session new
- open http://127.0.0.1:<port>/...
- eval / screenshot assertions

## Teardown
- kill server PID
```

The agent copies recipe commands into the verify transcript, always labels depth,
and after writing the transcript file **inlines the full transcript** in the reply.
