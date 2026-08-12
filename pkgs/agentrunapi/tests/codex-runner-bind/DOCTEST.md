# agentrunapi — EnsureCodexRunnerBound (codex-tty runner_session_id persist)

Classic TDD (**RED** until implementer lands bind persist). L2 in-process library
tests: when an agent-run **codex-tty** session is unbound
(`meta.runner_session_id` empty) and a Codex session UUID is discoverable,
**persist** it via `store.UpdateSessionRunnerSessionID` so status / resume /
pollers see `runner: bound`.

No live Codex, no live TTY, no process-global env/cwd. Fixtures use injected
`AGENT_RUN_HOME` (file store) and `CODEX_HOME` (rollouts) under `t.TempDir()`,
plus injectable scrollback.

**Out of scope:** evaluation package; L3 e2e with real Codex; turn-complete
semantics; CLI status stdout formatting.

# DSN (Domain Specific Notion)

**Participants**

- **Caller** — status / LifecycleProbe / AutoSendOrResume / open discovery wire
  (production). Tests call the library API directly.
- **`EnsureCodexRunnerBound`** — best-effort bind of `meta.runner_session_id` for
  **codex** runners only. Returns updated meta + whether a non-empty
  `runner_session_id` is present after the call.
- **Session store** — `agentstorage.Store` under injected home
  (`sessions/<id>/meta.json`). Persist path:
  `UpdateSessionRunnerSessionID(sessionID, codexUUID)`.
- **Session meta** — `runner`, `session_id`, `workspace`, `created_at`,
  `runner_session_id` (empty = unbound).
- **CODEX_HOME (injected)** — `sessions/YYYY/MM/DD/rollout-*-<uuid>.jsonl`.
  First `session_meta` line yields `payload.session_id` (or `id`), `payload.cwd`,
  timestamp. Match **cwd == meta.workspace** and **timestamp >= not-before**
  (`meta.created_at` or opts); pick **newest**.
- **Scrollback snapshot (injected)** — text that may contain
  `To continue this session, run codex resume <uuid>` (same extraction as
  `agenttty.FindCodexResumeSessionID`). Tests never open live registry TCP.

**Behaviors**

```
EnsureCodexRunnerBound(store, meta, opts):

  1. runner_session_id already non-empty
       -> no-op; return meta as-is; bound=true
  2. runner not codex (codex-tty | codex | codex-*)
       -> no-op; bound=false
  3. try scrollback: opts.SnapshotScrollback() if non-nil
       -> FindCodexResumeSessionID(text) -> id Y
  4. else/miss: rollout discovery under opts.CodexHome
       -> cwd match + timestamp not-before + newest -> id X
  5. on id found: store.UpdateSessionRunnerSessionID(meta.session_id, id)
       return meta with RunnerSessionID=id; bound=true
  6. miss: return meta unchanged; bound=false
       never hard-fail caller solely because bind missed
```

**Correct association** (real-machine evidence shape; fixtures use same fields):

| agent-run session | workspace cwd | Codex id |
|-------------------|---------------|----------|
| (fixture session) | injected workspace | rollout / resume uuid |

Not the agent-run session id string.

### Public API (Classic TDD — locked for implementer)

```go
// Package: github.com/xhd2015/agent-pro/pkgs/agentrunapi

// CodexRunnerBindOpts configures EnsureCodexRunnerBound.
// All paths and scrollback are injectable — no t.Setenv / t.Chdir / process
// globals required for product correctness under test.
type CodexRunnerBindOpts struct {
	// CodexHome is the CODEX_HOME root for rollout discovery.
	// Tests always set this; production may fall back to agenttty.CodexHome().
	CodexHome string

	// SnapshotScrollback, when non-nil, supplies TTY scrollback for resume-footer
	// bind without live registry/TCP. Nil means "no injected scrollback" (tests
	// for discovery-only leave this nil; production may resolve live TTY).
	SnapshotScrollback func() (string, error)

	// NotBefore, when non-zero, is the lower bound for rollout session timestamps.
	// When zero, implementer derives from meta.CreatedAt (RFC3339 / RFC3339Nano).
	NotBefore time.Time
}

// EnsureCodexRunnerBound ensures meta.runner_session_id is bound for codex runners
// when a Codex session id is discoverable (scrollback resume footer and/or
// CODEX_HOME rollout discovery). Best-effort: does not return an error solely
// because bind missed; store update errors are soft (still try return value).
//
// Rules:
//  1. If meta.RunnerSessionID already set → no-op, bound=true (never overwrite)
//  2. Only codex runners (codex-tty / codex / prefix codex)
//  3. Prefer resume footer from SnapshotScrollback when available
//  4. Else/also try rollout discovery: cwd == workspace, ts >= not-before, newest
//  5. On success: store.UpdateSessionRunnerSessionID + return updated meta
func EnsureCodexRunnerBound(store agentstorage.Store, meta agentstorage.SessionMeta, opts CodexRunnerBindOpts) (updated agentstorage.SessionMeta, bound bool)
```

Existing helpers implementer may reuse (do not break):

- `tryBindRunnerSessionFromZombie` (zombie_bind.go) — scrollback-only, live TTY
- `agenttty.FindCodexResumeSessionID` — footer parse
- `scanActiveCodexTranscripts` (agenttty, may need export or thin wrapper)

## Version

0.0.2

## Decision Tree

```
pkgs/agentrunapi/tests/codex-runner-bind/
├── DOCTEST.md
├── SETUP.md
├── discovery-bind/       # unbound codex + matching rollout → persist id
├── scrollback-bind/      # unbound codex + resume footer → persist id
├── already-bound/        # runner_session_id set → no overwrite
├── wrong-cwd/            # rollout other workspace → stay unbound
└── non-codex/            # grok-tty → no bind attempt side effect
```

Parameter ranking (most → least significant):

1. **Gate** — already bound / non-codex runner (short-circuit)
2. **Bind source** — scrollback footer vs rollout discovery
3. **Discovery filter** — cwd match + timestamp not-before

## Test Index

| # | Leaf | Description | Expect |
|---|------|-------------|--------|
| 1 | `discovery-bind` | Unbound codex-tty; CODEX_HOME rollout cwd=workspace, ts≥created_at → persist Codex id | RED |
| 2 | `scrollback-bind` | Unbound codex-tty; injected scrollback `codex resume Y` → persist Y | RED |
| 3 | `already-bound` | Bound id A; discovery/scrollback offer B → stay A; store not overwritten | RED |
| 4 | `wrong-cwd` | Rollout cwd ≠ workspace → stay unbound | RED |
| 5 | `non-codex` | Runner grok-tty + matching rollout/footer → no bind | RED |

## How to Run

```sh
cd .../external/agent-pro-master-2026-08-11
doctest vet ./pkgs/agentrunapi/tests/codex-runner-bind
doctest test ./pkgs/agentrunapi/tests/codex-runner-bind
doctest test -v ./pkgs/agentrunapi/tests/codex-runner-bind/discovery-bind
```

Expect: **RED** until implementer exports `EnsureCodexRunnerBound` / `CodexRunnerBindOpts`
and persists `meta.runner_session_id` (compile fail or assertion fail).

```go
import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/agent-pro/pkgs/agentrunapi"
	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
	"github.com/xhd2015/doctest/session"
)

// Request configures one EnsureCodexRunnerBound scenario.
// Paths are absolute under t.TempDir — never process-global env/cwd.
type Request struct {
	// Store home (AGENT_RUN_HOME layout). Root Setup defaults under TempDir.
	Home string
	// CodexHome root for rollout fixtures. Root Setup defaults under TempDir.
	CodexHome string
	// Workspace for meta + matching rollouts (absolute).
	Workspace string

	// Session seed
	SessionID       string
	Runner          string // default codex-tty
	RunnerSessionID string // pre-bound id; empty = unbound
	CreatedAt       string // RFC3339; seed into meta.CreatedAt
	Status          string // default running

	// SeedSession creates the session in the file store before Run.
	SeedSession bool

	// ScrollbackText, when non-empty, installs SnapshotScrollback inject.
	ScrollbackText string

	// NotBefore optional override for opts.NotBefore (zero => API uses meta.CreatedAt).
	NotBefore time.Time

	// Rollouts to seed under CodexHome before Run.
	// Empty slice = no rollouts.
	Rollouts []RolloutSeed
}

// RolloutSeed writes one rollout-*.jsonl with a session_meta line.
type RolloutSeed struct {
	// CodexSessionID is payload.session_id and path suffix.
	CodexSessionID string
	// Cwd is payload.cwd (absolute workspace path expected for match).
	Cwd string
	// Timestamp is session_meta timestamp (RFC3339 / RFC3339Nano).
	Timestamp string
	// DayPath optional YYYY/MM/DD under sessions/; default 2026/08/11.
	DayPath string
}

// Response captures EnsureCodexRunnerBound results + store re-read.
type Response struct {
	// Updated meta returned by the API.
	Meta agentstorage.SessionMeta
	// Bound flag returned by the API.
	Bound bool
	// StoredRunnerSessionID is meta.runner_session_id after re-read from store
	// (empty if session missing). Proves persist, not only return value.
	StoredRunnerSessionID string
}

// Run seeds store + CODEX_HOME fixtures, then calls EnsureCodexRunnerBound.
// Classic TDD: missing product symbols fail RED until implementer lands them.
func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	_ = d

	store, err := agentstorage.NewFileStore(req.Home)
	if err != nil {
		return nil, err
	}

	meta := agentstorage.SessionMeta{
		Runner:          req.Runner,
		SessionID:       req.SessionID,
		RunnerSessionID: req.RunnerSessionID,
		Status:          req.Status,
		Workspace:       req.Workspace,
		CreatedAt:       req.CreatedAt,
	}
	if req.SeedSession {
		if err := store.CreateSession(req.SessionID, meta); err != nil {
			return nil, err
		}
		// CreateSession may rewrite CreatedAt; re-apply fixed CreatedAt for
		// not-before discovery when leaf set an explicit value.
		if req.CreatedAt != "" {
			if err := rewriteMetaCreatedAt(t, store, req.SessionID, req.CreatedAt); err != nil {
				return nil, err
			}
			got, gerr := store.GetSession(req.SessionID)
			if gerr != nil {
				return nil, gerr
			}
			meta = got.Meta
		} else {
			got, gerr := store.GetSession(req.SessionID)
			if gerr != nil {
				return nil, gerr
			}
			meta = got.Meta
		}
	}

	if err := seedRollouts(t, req); err != nil {
		return nil, err
	}

	opts := agentrunapi.CodexRunnerBindOpts{
		CodexHome: req.CodexHome,
		NotBefore: req.NotBefore,
	}
	if req.ScrollbackText != "" {
		text := req.ScrollbackText
		opts.SnapshotScrollback = func() (string, error) {
			return text, nil
		}
	}

	updated, bound := agentrunapi.EnsureCodexRunnerBound(store, meta, opts)

	resp := &Response{
		Meta:  updated,
		Bound: bound,
	}
	if sess, err := store.GetSession(req.SessionID); err == nil {
		resp.StoredRunnerSessionID = strings.TrimSpace(sess.Meta.RunnerSessionID)
	}
	return resp, nil
}

func seedRollouts(t *testing.T, req *Request) error {
	t.Helper()
	for _, r := range req.Rollouts {
		if err := writeRollout(t, req.CodexHome, r); err != nil {
			return err
		}
	}
	return nil
}

func writeRollout(t *testing.T, codexHome string, r RolloutSeed) error {
	t.Helper()
	day := r.DayPath
	if day == "" {
		day = "2026/08/11"
	}
	id := strings.TrimSpace(r.CodexSessionID)
	if id == "" {
		return fmt.Errorf("rollout CodexSessionID required")
	}
	ts := strings.TrimSpace(r.Timestamp)
	if ts == "" {
		ts = "2026-08-11T12:00:00Z"
	}
	// Filename embeds uuid suffix the way real Codex rollouts do.
	base := "rollout-2026-08-11T12-00-00-" + id + ".jsonl"
	path := filepath.Join(codexHome, "sessions", filepath.FromSlash(day), base)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	line, err := json.Marshal(map[string]any{
		"timestamp": ts,
		"type":      "session_meta",
		"payload": map[string]any{
			"session_id": id,
			"cwd":        r.Cwd,
			"timestamp":  ts,
		},
	})
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(line, '\n'), 0o644)
}

// rewriteMetaCreatedAt overwrites meta.created_at after CreateSession so
// discovery not-before uses the leaf's fixed clock (CreateSession stamps now).
func rewriteMetaCreatedAt(t *testing.T, store agentstorage.Store, sessionID, createdAt string) error {
	t.Helper()
	sess, err := store.GetSession(sessionID)
	if err != nil {
		return err
	}
	sess.Meta.CreatedAt = createdAt
	data, err := json.Marshal(sess.Meta)
	if err != nil {
		return err
	}
	path := filepath.Join(store.Home(), "sessions", sessionID, "meta.json")
	return os.WriteFile(path, data, 0o644)
}
```
