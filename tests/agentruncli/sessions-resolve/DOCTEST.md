# agent-run sessions resolve (L2)

In-process doctests for the read-only CLI verb that prints an agent-run
`session_id` for a Grok provider `runner_session_id`. Uses P1
`agentstorage.FindByGrokSessionID`. Never creates a session.

Classic TDD: `agentruncli.RunSessions` (injected store + writers) and the
`resolve` verb do not exist yet — leaves go RED until the implementer exports
them. Nested root — no inheritance from `tests/agentruncli` Handle/stdio harness.

# DSN (Domain Specific Notion)

**Participants**

- **Caller** — harness or `Handle` after stripping the `sessions` token; passes
  remaining args into `RunSessions`.
- **RunSessions** — injected CLI entry: args, `Store`, stdout/stderr writers;
  returns product `error` (no `agent-run: ` prefix; main adds that outside L2).
- **Store** — `NewFileStore(explicitHome)` under an isolated temp home; seeds via
  `CreateSession` only in Setup/Run.
- **FindByGrokSessionID** — P1 library lookup; resolve delegates here only.
- **resolve verb** — reserved first remaining arg; handled before treating text
  as a bare session id.
- **Stdout** — human: bare `session_id` + `\n`; `--json`: one JSON object + `\n`.

**Behaviors**

- `resolve --grok-session-id ID` → Find; unique → print session id; 0/2+ → error.
- `resolve --json --grok-session-id ID` → same lookup; JSON with
  `session_id`, `runner`, `runner_session_id`, `status`.
- Flag omitted → `sessions resolve requires --grok-session-id`.
- Flag present but empty/whitespace → `--grok-session-id requires a non-empty value`.
- `resolve --help` / `-h` → nil error; documents Usage, `--grok-session-id`, `--json`.
- `RunSessions([]string{"-h"})` (sessions help) → mentions `resolve`.
- Read-only: no create-on-miss; errors leave stdout empty (no JSON error body required).

## Version

0.0.2

## Decision Tree

```
tests/agentruncli/sessions-resolve/
├── DOCTEST.md
├── SETUP.md
├── help/
│   ├── resolve-help/                 # resolve --help documents Usage + flags
│   └── sessions-help/                # sessions -h mentions resolve
└── resolve/
    ├── success/
    │   ├── human-grok-tty/           # stdout bare session_id\n
    │   ├── human-legacy-grok/        # runner=grok unique
    │   └── json/                     # --json object with four keys
    └── error/
        ├── not-found/                # library not-found + UUID; stdout empty
        ├── omitted-flag/             # sessions resolve requires --grok-session-id
        ├── empty-flag/               # --grok-session-id requires a non-empty value
        ├── ambiguous/                # two grok-tty; both session ids
        └── json-not-found/           # --json + miss still errors; no JSON body
```

Parameter ranking (most → least significant):

1. **Outcome class** — help vs resolve success vs resolve error
2. **Output mode** — human vs `--json` (success / json-not-found)
3. **Error kind** — omitted flag / empty flag / not-found / ambiguous
4. **Runner variant** — `grok-tty` vs legacy `grok` on the unique human path

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `help/resolve-help` | `resolve --help` → nil; stdout has Usage, `--grok-session-id`, `--json` |
| 2 | `help/sessions-help` | `RunSessions(-h)` → nil; stdout mentions `resolve` |
| 3 | `resolve/success/human-grok-tty` | Unique grok-tty → stdout `session_id\n`; err nil |
| 4 | `resolve/success/human-legacy-grok` | Unique `runner=grok` → same human stdout |
| 5 | `resolve/success/json` | `--json` → object with four keys; err nil |
| 6 | `resolve/error/not-found` | Missing UUID → not-found error; stdout empty |
| 7 | `resolve/error/omitted-flag` | No flag → `sessions resolve requires --grok-session-id` |
| 8 | `resolve/error/empty-flag` | Empty flag value → library empty-id message |
| 9 | `resolve/error/ambiguous` | Two grok-tty same UUID → ambiguous; both ids |
| 10 | `resolve/error/json-not-found` | `--json` + miss → error; stdout not a JSON object |

## How to Run

```sh
doctest vet ./tests/agentruncli/sessions-resolve
doctest test ./tests/agentruncli/sessions-resolve
doctest test -v ./tests/agentruncli/sessions-resolve/resolve/success/human-grok-tty
```

```go
import (
	"bytes"
	"fmt"
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/agentruncli"
	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
	"github.com/xhd2015/doctest/session"
)

// SeedMeta is one session written via CreateSession before RunSessions.
type SeedMeta struct {
	SessionID       string
	Runner          string
	RunnerSessionID string
	Status          string
}

type Request struct {
	TempDir string
	Home    string
	Seeds   []SeedMeta
	// Args are passed to RunSessions (after the sessions token).
	Args []string
}

type Response struct {
	Stdout string
	Stderr string
	Err    error
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	if req.Home == "" {
		return nil, fmt.Errorf("req.Home must be set by root Setup")
	}
	if req.Args == nil {
		return nil, fmt.Errorf("req.Args must be set by leaf Setup (use empty slice only for intentional empty argv)")
	}

	store, err := agentstorage.NewFileStore(req.Home)
	if err != nil {
		return nil, fmt.Errorf("NewFileStore: %w", err)
	}
	for _, s := range req.Seeds {
		status := s.Status
		if status == "" {
			status = "finished"
		}
		meta := agentstorage.SessionMeta{
			Runner:          s.Runner,
			SessionID:       s.SessionID,
			RunnerSessionID: s.RunnerSessionID,
			Status:          status,
		}
		if err := store.CreateSession(s.SessionID, meta); err != nil {
			return nil, fmt.Errorf("CreateSession(%s): %w", s.SessionID, err)
		}
	}

	var stdout, stderr bytes.Buffer
	runErr := agentruncli.RunSessions(req.Args, store, &stdout, &stderr)
	return &Response{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
		Err:    runErr,
	}, nil
}
```
