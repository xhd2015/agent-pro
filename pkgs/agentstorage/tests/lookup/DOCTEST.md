# agentstorage runner_session_id lookup + lazy cache

In-process L2 doctests for finding agent-run sessions by provider
`runner_session_id`, with a lazy per-UUID on-disk cache under the store home.

Classic TDD: lookup API and cache do not exist yet — leaves go RED until
`pkgs/agentstorage` exports `IsGrokRunner`, `ListByRunnerSessionID`, and
`FindByGrokSessionID`, and bumps store generation on write.

# DSN (Domain Specific Notion)

Lookup maps a provider UUID to agent-run session metas without creating
sessions. Cache is a rebuildable hint; `sessions/*/meta.json` stays
authoritative.

**Participants**

- **Caller** — library consumer invoking Find / List / IsGrokRunner.
- **Store** — `NewFileStore(explicitHome)`; seeds via `CreateSession` and mutates
  via `UpdateSessionRunnerSessionID` / `CreateSession` / `ClearAllSessions`.
- **SessionMeta** — `runner`, `session_id`, `runner_session_id` on disk.
- **FindByGrokSessionID** — List filtered to `grok` + `grok-tty`, then
  cardinality 0 / 1 / 2+.
- **ListByRunnerSessionID** — all metas matching trimmed UUID; optional exact
  runner filter.
- **IsGrokRunner** — exact trim match for `grok` or `grok-tty` only.
- **Lazy index** — `index/generation`, `index/by-runner-session/.gen`, and
  per-UUID `*.json` entry lists rebuilt on cold/stale Find/List.

**Behaviors**

- Empty / whitespace query id → error `--grok-session-id requires a non-empty value`.
- Find: 0 → not found; 1 → that meta; 2+ → ambiguous (session ids asc, comma-space).
- List with runners filters exact trimmed runner names; no runners → all matches.
- Cold/stale (gens unequal or index missing) → full `ListSessions` scan; write one
  cache file per non-empty `runner_session_id`; set `.gen` to store gen.
- Warm (gens equal) → read UUID file if present; missing file → no matches (no scan).
- Writes bump `index/generation` on CreateSession, non-empty
  UpdateSessionRunnerSessionID, and ClearAllSessions (which also drops the index
  mapping dir). Next Find/List rebuilds; bind does not eagerly rewrite UUID files.

## Version

0.0.2

## Decision Tree

```
pkgs/agentstorage/tests/lookup/
├── DOCTEST.md
├── SETUP.md
├── is-grok-runner/
│   └── classify/                      # exact grok|grok-tty vs non-matches
├── find-list/
│   ├── unique/
│   │   ├── grok-tty/                  # Find unique runner=grok-tty
│   │   └── legacy-grok/               # Find unique runner=grok
│   ├── missing/
│   │   ├── unknown-uuid/              # Find not-found for never-seeded UUID
│   │   └── empty-rsid-meta-not-hit/   # empty meta.runner_session_id is not a hit
│   ├── empty-query-id/                # Find+List reject blank/whitespace id
│   ├── ambiguous-two-grok/            # Find ambiguous; ListBy grok filter len=2
│   ├── ignores-non-grok/              # same UUID on codex+grok → Find returns grok
│   └── list-unfiltered/               # ListBy no runners → grok+codex len=2
└── cache/
    ├── cold-rebuild-writes-all-uuids/ # first Find writes queried + sibling UUID files
    ├── warm-hit/                      # second Find; .gen unchanged; correct meta
    ├── warm-known-miss-no-file/       # post-rebuild unknown UUID → not-found, no new file
    ├── stale-after-bind-update/       # UpdateSessionRunnerSessionID bumps gen; next Find sees bind
    ├── stale-second-match-visible/    # CreateSession same UUID → next Find ambiguous
    └── clear-all-sessions/            # ClearAllSessions drops index; Find not-found
```

Parameter ranking (most → least significant):

1. **Concern** — predicate (`IsGrokRunner`) vs Find/List cardinality vs cache lifecycle
2. **Outcome class** — unique / missing / empty / ambiguous / filter
3. **Cache phase** — cold rebuild / warm hit-or-miss / stale after write / clear

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `is-grok-runner/classify` | `grok`/`grok-tty` true; `codex-tty`/`GROK`/`grok-foo`/empty false |
| 2 | `find-list/unique/grok-tty` | Unique `grok-tty` → Find returns that SessionID |
| 3 | `find-list/unique/legacy-grok` | Unique legacy `runner=grok` → Find returns that meta |
| 4 | `find-list/missing/unknown-uuid` | Find error contains `not found` and the UUID |
| 5 | `find-list/missing/empty-rsid-meta-not-hit` | Seeded empty `runner_session_id` does not match a query UUID |
| 6 | `find-list/empty-query-id` | Empty/whitespace → `--grok-session-id requires a non-empty value` (Find+List) |
| 7 | `find-list/ambiguous-two-grok` | Two grok-tty same UUID → Find ambiguous (ids asc); List len=2 |
| 8 | `find-list/ignores-non-grok` | codex-tty + grok-tty same UUID → Find returns only grok-tty |
| 9 | `find-list/list-unfiltered` | ListBy with no runner filter returns grok and codex (len=2) |
| 10 | `cache/cold-rebuild-writes-all-uuids` | Cold Find(A) creates `A.json`, `B.json`, and `.gen` |
| 11 | `cache/warm-hit` | Second Find same UUID; meta correct; `.gen` unchanged |
| 12 | `cache/warm-known-miss-no-file` | After rebuild, unknown UUID not-found; no cache file created |
| 13 | `cache/stale-after-bind-update` | After warm, UpdateSessionRunnerSessionID; next Find sees new mapping |
| 14 | `cache/stale-second-match-visible` | After unique warm, CreateSession same UUID → Find ambiguous |
| 15 | `cache/clear-all-sessions` | ClearAllSessions removes index mapping; Find not-found |

## How to Run

```sh
doctest vet ./pkgs/agentstorage/tests/lookup
doctest test ./pkgs/agentstorage/tests/lookup
doctest test -v ./pkgs/agentstorage/tests/lookup/find-list/ambiguous-two-grok
doctest test -v ./pkgs/agentstorage/tests/lookup/cache/cold-rebuild-writes-all-uuids
```

```go
import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
	"github.com/xhd2015/doctest/session"
)

// SeedMeta is one session written via CreateSession before the asserted op.
type SeedMeta struct {
	SessionID       string
	Runner          string
	RunnerSessionID string
	Status          string
}

// RunnerCase is one IsGrokRunner input/expectation.
type RunnerCase struct {
	Name   string
	Runner string
	Want   bool
}

// MutateOp is an optional store write after an optional warm lookup.
type MutateOp struct {
	Kind            string // update_rsid | create | clear_all
	SessionID       string
	Runner          string
	RunnerSessionID string
	Status          string
}

// CacheSnap captures lazy-index filesystem facts for asserts.
type CacheSnap struct {
	IndexDirExists bool
	ByRunnerExists bool
	Generation     string // index/generation contents (trimmed)
	ByRunnerGen    string // index/by-runner-session/.gen contents (trimmed)
	UUIDFiles      []string // basenames without .json, sorted
}

type Request struct {
	TempDir string
	Home    string

	Seeds []SeedMeta

	// WarmQueryID when set: perform WarmOp lookup before Mutate (populates cache).
	WarmQueryID string
	WarmOp      string // find | list | "" (default find when WarmQueryID set)

	Mutate *MutateOp

	Op string // is_grok | find | list | find_and_list

	QueryID string
	Runners []string // ListByRunnerSessionID filter; empty = no filter

	RunnerCases []RunnerCase
}

type Response struct {
	Meta    agentstorage.SessionMeta
	Metas   []agentstorage.SessionMeta
	FindErr error
	ListErr error
	// Err is the primary product error for single-op leaves (Find or List).
	Err error

	IsGrok []bool

	WarmGen    string
	CacheAfter CacheSnap
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	if req.Home == "" {
		return nil, fmt.Errorf("req.Home must be set by root Setup")
	}
	if req.Op == "" {
		return nil, fmt.Errorf("req.Op must be set by leaf Setup")
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

	resp := &Response{}

	if req.WarmQueryID != "" {
		wop := req.WarmOp
		if wop == "" {
			wop = "find"
		}
		switch wop {
		case "find":
			_, _ = agentstorage.FindByGrokSessionID(store, req.WarmQueryID)
		case "list":
			_, _ = agentstorage.ListByRunnerSessionID(store, req.WarmQueryID)
		default:
			return nil, fmt.Errorf("unknown WarmOp %q", wop)
		}
		warm := snapshotCache(req.Home)
		resp.WarmGen = warm.ByRunnerGen
	}

	if req.Mutate != nil {
		if err := applyMutate(store, req.Mutate); err != nil {
			return nil, fmt.Errorf("mutate: %w", err)
		}
	}

	switch req.Op {
	case "is_grok":
		if len(req.RunnerCases) == 0 {
			return nil, fmt.Errorf("req.RunnerCases required for is_grok")
		}
		resp.IsGrok = make([]bool, len(req.RunnerCases))
		for i, c := range req.RunnerCases {
			resp.IsGrok[i] = agentstorage.IsGrokRunner(c.Runner)
		}
	case "find":
		meta, err := agentstorage.FindByGrokSessionID(store, req.QueryID)
		resp.Meta = meta
		resp.FindErr = err
		resp.Err = err
	case "list":
		metas, err := agentstorage.ListByRunnerSessionID(store, req.QueryID, req.Runners...)
		resp.Metas = metas
		resp.ListErr = err
		resp.Err = err
	case "find_and_list":
		meta, ferr := agentstorage.FindByGrokSessionID(store, req.QueryID)
		resp.Meta = meta
		resp.FindErr = ferr
		resp.Err = ferr
		metas, lerr := agentstorage.ListByRunnerSessionID(store, req.QueryID, req.Runners...)
		resp.Metas = metas
		resp.ListErr = lerr
	default:
		return nil, fmt.Errorf("unknown Op %q", req.Op)
	}

	resp.CacheAfter = snapshotCache(req.Home)
	return resp, nil
}

func applyMutate(store agentstorage.Store, m *MutateOp) error {
	switch m.Kind {
	case "update_rsid":
		return store.UpdateSessionRunnerSessionID(m.SessionID, m.RunnerSessionID)
	case "create":
		status := m.Status
		if status == "" {
			status = "finished"
		}
		return store.CreateSession(m.SessionID, agentstorage.SessionMeta{
			Runner:          m.Runner,
			SessionID:       m.SessionID,
			RunnerSessionID: m.RunnerSessionID,
			Status:          status,
		})
	case "clear_all":
		return store.ClearAllSessions()
	default:
		return fmt.Errorf("unknown Mutate.Kind %q", m.Kind)
	}
}

func snapshotCache(home string) CacheSnap {
	var snap CacheSnap
	indexDir := filepath.Join(home, "index")
	byDir := filepath.Join(indexDir, "by-runner-session")
	if st, err := os.Stat(indexDir); err == nil && st.IsDir() {
		snap.IndexDirExists = true
	}
	if st, err := os.Stat(byDir); err == nil && st.IsDir() {
		snap.ByRunnerExists = true
	}
	snap.Generation = readTrimmed(filepath.Join(indexDir, "generation"))
	snap.ByRunnerGen = readTrimmed(filepath.Join(byDir, ".gen"))
	entries, err := os.ReadDir(byDir)
	if err != nil {
		return snap
	}
	for _, ent := range entries {
		name := ent.Name()
		if ent.IsDir() || !strings.HasSuffix(name, ".json") {
			continue
		}
		snap.UUIDFiles = append(snap.UUIDFiles, strings.TrimSuffix(name, ".json"))
	}
	sort.Strings(snap.UUIDFiles)
	return snap
}

func readTrimmed(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}
```
