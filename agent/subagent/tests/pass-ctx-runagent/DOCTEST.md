# Pass Context Through runAgent

Verify that `runAgent` accepts `ctx context.Context` as its first parameter (replacing the hardcoded 30-second timeout), and that the context from the caller propagates to `runner.Agent.Ask`.

## Decision Tree

```
pass-ctx-runagent/
├── DOCTEST.md                # This file
├── SETUP.md                  # Root: Package under test = agent/subagent
│                              #   Run() calls runAgent(ctx, ...) directly
│
└── ctx-propagation/          # Context propagation behavior
    ├── SETUP.md              # Sets AgentRunner = "fake-codex"
    ├── canceled/             # Pre-canceled context → error
    │   ├── SETUP.md          # CancelCtx = true
    │   └── ASSERT.md         # err != nil
    └── not-canceled/         # Normal context → output
        ├── SETUP.md          # Prompt = "write hello world"
        └── ASSERT.md         # err == nil && output != ""
```

Branches:
- **Root**: Split on context state (canceled vs normal).
- **Group**: `ctx-propagation` narrows to cancellation behavior.
- **Leaves**: One for each context state.

## Test Index

| Leaf | Description |
|------|-------------|
| `ctx-propagation/canceled` | Pre-canceled context passed to `runAgent` → returns error |
| `ctx-propagation/not-canceled` | Normal `context.Background()` passed to `runAgent` → returns output |

## How to Run

```sh
# Validate tree structure
doctest vet ./external/agent-pro/agent/subagent/tests/pass-ctx-runagent/

# Run tests (expect RED — compilation fails until runAgent signature is updated)
doctest test -v ./external/agent-pro/agent/subagent/tests/pass-ctx-runagent/
```
