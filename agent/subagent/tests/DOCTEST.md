# Subagent Library Tests

Verify the `subagent` package in `github.com/xhd2015/agent-pro/agent/subagent` after migration from `github.com/xhd2015/doctest/libdoc/subagent`.

Focus areas: configurable session base directory (R7), generic environment variable naming (R9), session management operations (R10), and Logf behavior (R8).

## Decision Tree

```
external/agent-pro/agent/subagent/tests/
├── DOCTEST.md                         # This file
├── SETUP.md                           # Root: Request/Response types, stub Run()
│
├── config-env-vars/                   # === Configurable env var / meta field names ===
│   ├── SETUP.md                       # Run() constructs Config with custom field names
│   ├── session-env-var/
│   │   ├── SETUP.md
│   │   ├── custom-env-var/            # Custom SessionEnvVar → env var read for session ID
│   │   └── default-env-var/           # SessionEnvVar empty → default env var used
│   ├── session-meta-field/
│   │   ├── SETUP.md
│   │   ├── custom-meta-field/        # Custom SessionMetaField → stored in meta.json, matched
│   │   └── default-meta-field/       # SessionMetaField empty → default meta field used
│   └── debug-session-env/
│       ├── SETUP.md
│       ├── custom-debug-env/          # Custom DebugSessionEnv → overrides session base
│       └── default-debug-env/         # DebugSessionEnv empty → default env var checked
│
├── session-base/                      # === R7: Configurable sessions base ===
│   ├── SETUP.md                       # Run() calls ListSessions, captures stdout
│   ├── explicit-session-base/         # SessionBase="/tmp/custom" → sessions listed from there
│   ├── default-base/                  # SessionBase="" → ~/.agent-pro/subagent/<role>/sessions/
│   └── env-override/                  # AGENT_PRO_SUBAGENT_DEBUG_SESSION_HOME overrides SessionBase
│
├── session-id-resolution/             # === R9: Generic session env var naming ===
│   ├── SETUP.md                       # Run() resolves session ID via Run(Status)
│   ├── from-env-var/                  # AGENT_PRO_SUBAGENT_TESTROLE_SESSION_ID → used as session ID
│   ├── flag-overrides-env/            # --session-id flag takes priority over env var
│   ├── codex-thread-id-fallback/      # CODEX_THREAD_ID used as fallback when no flag/env
│   └── no-session-id-error/           # No flag, no env, no CODEX_THREAD_ID → error
│
├── logf/                              # === R8: Logf output behavior ===
│   ├── SETUP.md                       # Run() captures stdout, calls subagent.Logf()
│   ├── without-trailing-newline/      # Message without \n → one \n appended
│   ├── with-trailing-newline/         # Message with \n → exactly one \n
│   ├── empty-message/                 # Empty string → timestamp + \n
│   ├── format-verbs/                  # Format verbs (%s, %d) with args
│   └── special-chars/                 # Multiline content preserved
│
└── session-operations/                # === R10: Session management features ===
    ├── SETUP.md                       # Run() creates session dirs, calls Run(option)
    ├── list-sessions/
    │   ├── SETUP.md
    │   ├── empty/                     # No sessions → "No sessions found"
    │   └── with-sessions/            # Sessions exist → listed sorted by time
    ├── show-status/
    │   ├── SETUP.md
    │   ├── session-not-found/         # Session not found → stderr error
    │   ├── finished-session/          # Session with meta + events → status output
    │   └── running-session/           # Session with pid → "running (PID ...)"
    └── trace-session/
        ├── SETUP.md
        ├── no-events/                 # No events file → "(no events yet)"
        └── with-events/              # Events present → formatted event lines
```

## Test Index

### config-env-vars — 6 leaves
| Leaf | Description |
|------|-------------|
| `session-env-var/custom-env-var` | `Config.SessionEnvVar` set → custom env var read for session ID |
| `session-env-var/default-env-var` | `Config.SessionEnvVar` empty → default env var read |
| `session-meta-field/custom-meta-field` | `Config.SessionMetaField` set → stored in meta.json, used for matching |
| `session-meta-field/default-meta-field` | `Config.SessionMetaField` empty → default meta field used |
| `debug-session-env/custom-debug-env` | `Config.DebugSessionEnv` set → custom env var overrides session base |
| `debug-session-env/default-debug-env` | `Config.DebugSessionEnv` empty → default env var checked |

### session-base (R7) — 3 leaves
| Leaf | Description |
|------|-------------|
| `explicit-session-base` | `Options.SessionBase` set → sessions stored/listed from that path |
| `default-base` | `SessionBase` empty → defaults to `~/.agent-pro/subagent/<role>/sessions/` |
| `env-override` | `AGENT_PRO_SUBAGENT_DEBUG_SESSION_HOME` set → overrides `SessionBase` |

### session-id-resolution (R9) — 4 leaves
| Leaf | Description |
|------|-------------|
| `from-env-var` | `AGENT_PRO_SUBAGENT_<ROLE>_SESSION_ID` env var → session ID from env |
| `flag-overrides-env` | `--session-id` flag set → takes priority over env var |
| `codex-thread-id-fallback` | No flag/env → uses `CODEX_THREAD_ID` env var |
| `no-session-id-error` | Nothing set → error with generated session ID |

### logf (R8) — 5 leaves
| Leaf | Description |
|------|-------------|
| `without-trailing-newline` | Message without `\n` → `\n` appended |
| `with-trailing-newline` | Message with `\n` → exactly one `\n` |
| `empty-message` | Empty string → timestamp `[YYYY-MM-DDTHH:MM:SS]` + `\n` |
| `format-verbs` | Format verbs with args → properly formatted |
| `special-chars` | Multiline, special chars → preserved verbatim |

### session-operations (R10) — 7 leaves

#### list-sessions (2 leaves)
| Leaf | Description |
|------|-------------|
| `empty` | No session dirs → "No sessions found" |
| `with-sessions` | Multiple sessions → sorted by creation time descending |

#### show-status (3 leaves)
| Leaf | Description |
|------|-------------|
| `session-not-found` | Unknown session ID → stderr error |
| `finished-session` | Session with meta.json + events.jsonl → formatted status |
| `running-session` | Session with pid file matching live process → "running (PID ...)" |

#### trace-session (2 leaves)
| Leaf | Description |
|------|-------------|
| `no-events` | No events.jsonl → "(no events yet)" header, Done footer |
| `with-events` | events.jsonl exists → formatted event lines, Done footer |

Total: **25 leaves** across **5 feature areas**.

## How to Run

```sh
doctest test -v ./external/agent-pro/agent/subagent/tests/
```
