# Agent Runner Detection — Ancestor Walk

Verify the `autoDetectAgentRunner` function in `github.com/xhd2015/agent-pro/agent/subagent` correctly identifies which agent runner (opencode, pi, codex, crush, grok) is hosting the current process.

This test covers **ancestor walk** enhancements: when the immediate parent is a shell (bash, zsh, etc.), walk up one more level to detect pi or grok as the grandparent.

## Detection Chain (Priority Order)

```
1. AGENT_PRO_SUBAGENT_<ROLE>_AGENT_RUNNER env override  →  returns env value
2. CODEX_THREAD_ID env  →  returns "codex"
3. PI_CODING_AGENT env  →  returns "pi"
4. Parent process detection:
   a. getProcessName(ppid)   →  match opencode/pi/codex/crush/grok
   b. if no match: getProcessName(pppid)  →  match "pi" or "grok"
```

## Decision Tree

```
agent-runner-detection/
├── DOCTEST.md                           # This file
├── SETUP.md                             # Root: Request/Response types, Run()
│
├── env-override/                        # === Priority 1: AgentRunnerEnv ===
│   ├── SETUP.md                         # AgentRunnerEnv = "TEST_AGENT_RUNNER"
│   ├── returns-specified-value/         # TEST_AGENT_RUNNER=pi → "pi", true
│   │   ├── SETUP.md
│   │   └── ASSERT.md
│   └── env-var-absent/                  # TEST_AGENT_RUNNER not set → falls through
│       ├── SETUP.md
│       └── ASSERT.md
│
├── codex-thread-id/                     # === Priority 2: CODEX_THREAD_ID ===
│   ├── SETUP.md                         # Ensure no env override, no process hook
│   └── returns-codex/                   # CODEX_THREAD_ID=abc123 → "codex", true
│       ├── SETUP.md
│       └── ASSERT.md
│
├── pi-coding-agent/                     # === Priority 3: PI_CODING_AGENT ===
│   ├── SETUP.md                         # Ensure no CODEX_THREAD_ID
│   └── returns-pi/                      # PI_CODING_AGENT=1 → "pi", true
│       ├── SETUP.md
│       └── ASSERT.md
│
├── priority-precedence/                 # === Cross-priority interactions ===
│   ├── SETUP.md
│   ├── override-beats-codex/            # P1 env override set + CODEX_THREAD_ID → P1 wins
│   │   ├── SETUP.md
│   │   └── ASSERT.md
│   └── codex-beats-pi-env/              # CODEX_THREAD_ID + PI_CODING_AGENT → P2 (codex) wins
│       ├── SETUP.md
│       └── ASSERT.md
│
└── parent-process/                      # === Priority 4: Process tree ===
    ├── SETUP.md                         # Installs TestProcessNameFunc hook
    ├── direct-ppid/                     # ppid matches a known agent
    │   ├── SETUP.md
    │   ├── pi/                          # ppid="pi" → "pi", true
    │   │   ├── SETUP.md
    │   │   └── ASSERT.md
    │   ├── opencode/                    # ppid="opencode" → "opencode", true
    │   │   ├── SETUP.md
    │   │   └── ASSERT.md
    │   ├── codex/                       # ppid="codex" → "codex", true
    │   │   ├── SETUP.md
    │   │   └── ASSERT.md
    │   ├── crush/                       # ppid="crush" → "crush", true
    │   │   ├── SETUP.md
    │   │   └── ASSERT.md
    │   └── grok/                        # ppid="grok" → "grok", true
    │       ├── SETUP.md
    │       └── ASSERT.md
    ├── grandparent-pi/                  # ppid not agent, pppid = "pi"
    │   ├── SETUP.md                     # First call returns non-agent ppid name
    │   ├── bash-ppid/                   # ppid="bash", pppid="pi" → "pi", true
    │   │   ├── SETUP.md
    │   │   └── ASSERT.md
    │   └── zsh-ppid/                    # ppid="zsh", pppid="pi" → "pi", true
    │       ├── SETUP.md
    │       └── ASSERT.md
    ├── grandparent-grok/                # ppid not agent, pppid = "grok"
    │   ├── SETUP.md                     # First call returns non-agent ppid name
    │   ├── bash-ppid/                   # ppid="bash", pppid="grok" → "grok", true
    │   │   ├── SETUP.md
    │   │   └── ASSERT.md
    │   └── zsh-ppid/                    # ppid="zsh", pppid="grok" → "grok", true
    │       ├── SETUP.md
    │       └── ASSERT.md
    └── grandparent-non-pi/              # ppid not agent, pppid not pi/grok
        ├── SETUP.md                     # First call returns non-agent ppid name
        ├── non-agent-ancestry/          # ppid="bash", pppid="bash" → "", false
        │   ├── SETUP.md
        │   └── ASSERT.md
        ├── codex-grandparent/           # ppid="bash", pppid="codex" → "", false
        │   ├── SETUP.md
        │   └── ASSERT.md
        └── opencode-grandparent/        # ppid="bash", pppid="opencode" → "", false
            ├── SETUP.md
            └── ASSERT.md
```

## Test Index

### env-override (Priority 1) — 2 leaves
| Leaf | Description |
|------|-------------|
| `returns-specified-value` | AgentRunnerEnv configured and env var set → returns that value |
| `env-var-absent` | AgentRunnerEnv configured but env var not set → falls through to lower priorities |

### codex-thread-id (Priority 2) — 1 leaf
| Leaf | Description |
|------|-------------|
| `returns-codex` | CODEX_THREAD_ID set → returns "codex", true |

### pi-coding-agent (Priority 3) — 1 leaf
| Leaf | Description |
|------|-------------|
| `returns-pi` | PI_CODING_AGENT set → returns "pi", true |

### priority-precedence — 2 leaves
| Leaf | Description |
|------|-------------|
| `override-beats-codex` | Both P1 env override and CODEX_THREAD_ID set → P1 wins |
| `codex-beats-pi-env` | Both CODEX_THREAD_ID and PI_CODING_AGENT set → P2 (codex) wins over P3 (pi) |

### parent-process (Priority 4) — 12 leaves

#### direct-ppid (5 leaves)
| Leaf | Description |
|------|-------------|
| `pi` | ppid="pi" → returns "pi", true |
| `opencode` | ppid="opencode" → returns "opencode", true |
| `codex` | ppid="codex" → returns "codex", true |
| `crush` | ppid="crush" → returns "crush", true |
| `grok` | ppid="grok" → returns "grok", true |

#### grandparent-pi (2 leaves)
| Leaf | Description |
|------|-------------|
| `bash-ppid` | ppid="bash", pppid="pi" → grandparent walk finds pi (bash shell) |
| `zsh-ppid` | ppid="zsh", pppid="pi" → grandparent walk finds pi (zsh shell variant) |

#### grandparent-grok (2 leaves)
| Leaf | Description |
|------|-------------|
| `bash-ppid` | ppid="bash", pppid="grok" → grandparent walk finds grok (bash shell) |
| `zsh-ppid` | ppid="zsh", pppid="grok" → grandparent walk finds grok (zsh shell variant) |

#### grandparent-non-pi (3 leaves)
| Leaf | Description |
|------|-------------|
| `non-agent-ancestry` | ppid="bash", pppid="bash" → no agent in ancestry, not detected |
| `codex-grandparent` | ppid="bash", pppid="codex" → grandparent walk is pi/grok-only, not detected |
| `opencode-grandparent` | ppid="bash", pppid="opencode" → grandparent walk is pi/grok-only, not detected |

Total: **18 leaves** across **5 feature areas**.

## How to Run

```sh
# Validate tree structure
doctest vet ./external/agent-pro/agent/subagent/tests/agent-runner-detection/

# Run tests
doctest test -v ./external/agent-pro/agent/subagent/tests/agent-runner-detection/
```

```go
import (

    "sync"
    "testing"

    "github.com/xhd2015/agent-pro/agent/subagent"
    "github.com/xhd2015/doctest/session"
)


type Request struct {
    Env            []string
    AgentRunnerEnv string
    ProcessNames   []string
}

type Response struct {
    Runner   string
    Detected bool
}

// envLookup: force-unset detection keys, apply req.Env; request-scoped (no Setenv).
// splitEnv is defined in SETUP.md.
func envLookup(env []string, forceUnset ...string) func(string) (string, bool) {
    overlay := map[string]*string{}
    for _, k := range forceUnset {
        overlay[k] = nil
    }
    for _, e := range env {
        parts := splitEnv(e)
        if len(parts) != 2 {
            continue
        }
        v := parts[1]
        overlay[parts[0]] = &v
    }
    return func(key string) (string, bool) {
        p, ok := overlay[key]
        if !ok {
            return "", false
        }
        if p == nil {
            return "", true
        }
        return *p, true
    }
}

// processNameHookMu serializes TestProcessNameFunc package-global mutation
// (same class of residual as product Test hooks under t.Parallel).
var processNameHookMu sync.Mutex

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
    if len(req.ProcessNames) > 0 {
        processNameHookMu.Lock()
        defer processNameHookMu.Unlock()
        callIdx := 0
        subagent.TestProcessNameFunc = func(pid int) string {
            if callIdx < len(req.ProcessNames) {
                name := req.ProcessNames[callIdx]
                callIdx++
                return name
            }
            return ""
        }
        defer func() { subagent.TestProcessNameFunc = nil }()
    }

    runner, detected := subagent.TestExported_autoDetectAgentRunner(subagent.Config{
        AgentRunnerEnv: req.AgentRunnerEnv,
        EnvLookup:      envLookup(req.Env, "CODEX_THREAD_ID", "PI_CODING_AGENT"),
    })

    return &Response{Runner: runner, Detected: detected}, nil
}
```
