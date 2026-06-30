# agentstorage Tests

Doc-style tests for `github.com/xhd2015/agent-pro/pkgs/agentstorage`, covering
file-backed session storage under `AGENT_RUN_HOME`.

# DSN (Domain Specific Notion)

`FileStore` is the default implementation of `Store`. All durable state lives
under a single home directory (default `~/.agent-run/`, overridden in tests via
`AGENT_RUN_HOME`).

```
AGENT_RUN_HOME/
  config.json              # {default_agent_runner, default_model, last_session}
  auth.token               # optional web auth token (not exercised here)
  sessions/<runner>/<session_id>/
    meta.json              # SessionMeta: runner, session_id, status, model, workspace, timestamps
    events.jsonl           # NDJSON lines of types.AgentEvent
    messages.jsonl         # queued user follow-ups {id, text, session_id, created_at}
  cursors/
    web-ui.json
    tui.json
```

`Store` methods map to paths: `Config`/`SaveConfig` ↔ `config.json`;
`CreateSession`/`GetSession`/`UpdateSessionStatus`/`ListSessions` ↔ `meta.json`;
`AppendEvent`/`ReadEvents` ↔ `events.jsonl` (offset = byte position for resume);
`AppendMessage`/`ListMessages`/`PopMessages` ↔ `messages.jsonl` (FIFO pop).

Tests call the package directly (no CLI). Each run uses an isolated temp home
`filepath.Join(t.TempDir(), ".agent-run")` with `AGENT_RUN_HOME` set accordingly.

## Version

0.0.2

## Decision Tree

```
pkgs/agentstorage/tests/
├── DOCTEST.md
├── SETUP.md
├── home/
│   ├── env-override-home-path/     AGENT_RUN_HOME overrides explicit Home path
│   └── creates-session-dirs-on-first-write/
│                                     first write creates sessions/<runner>/<id>/
├── config/
│   ├── load-missing-returns-empty-defaults/
│   └── save-and-reload-roundtrip/
├── session/
│   ├── create-get-update-status/
│   ├── list-filters-by-runner/
│   ├── get-missing-error/
│   └── workspace-roundtrip/          CreateSession workspace → GetSession preserves path
├── events/
│   ├── append-and-read-from-start/
│   ├── read-after-offset-skips-prior/
│   └── read-empty-session/
├── messages/
│   ├── append-list-roundtrip/
│   ├── pop-fifo-order/
│   └── pop-empty-queue/
└── isolation/
    └── writes-stay-under-home/     no files written outside AGENT_RUN_HOME
```

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `home/env-override-home-path` | `AGENT_RUN_HOME` wins over constructor home argument |
| 2 | `home/creates-session-dirs-on-first-write` | First `AppendEvent` creates session directory tree |
| 3 | `config/load-missing-returns-empty-defaults` | Missing `config.json` yields empty `Config` defaults |
| 4 | `config/save-and-reload-roundtrip` | `SaveConfig` then `Config` returns identical values |
| 5 | `session/create-get-update-status` | Create, get, update status, get again |
| 6 | `session/list-filters-by-runner` | `ListSessions` returns only matching runner |
| 7 | `session/get-missing-error` | `GetSession` on unknown id returns error |
| 7b | `session/workspace-roundtrip` | `CreateSession` with `workspace` → `GetSession` returns same path |
| 8 | `events/append-and-read-from-start` | Appended events readable from offset 0 |
| 9 | `events/read-after-offset-skips-prior` | `ReadEvents(afterOffset)` skips earlier lines |
| 10 | `events/read-empty-session` | New session returns zero events and offset 0 |
| 11 | `messages/append-list-roundtrip` | `AppendMessage` then `ListMessages` preserves order |
| 12 | `messages/pop-fifo-order` | `PopMessages` drains oldest-first |
| 13 | `messages/pop-empty-queue` | `PopMessages` on empty queue returns nil/empty slice |
| 14 | `isolation/writes-stay-under-home` | All tracked writes remain under home prefix |

## How to Run

```sh
doctest vet ./pkgs/agentstorage/tests
doctest test -v ./pkgs/agentstorage/tests
doctest test -v ./pkgs/agentstorage/tests/session/create-get-update-status
```

```go
import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
)

type Request struct {
	Home        string
	TempDir     string
	Env         []string
	Operation   string // home | config | session | events | messages | isolation
	Action      string // leaf-specific action within Operation
	Runner      string
	SessionID   string
	RunnerSess  string
	Status      string
	Model       string
	Workspace   string
	Config      agentstorage.Config
	Events      []types.AgentEvent
	MessageText []string
	AfterOffset int64
	OtherRunner string
	OtherSessID string
}

type Response struct {
	Store        agentstorage.Store
	Config       agentstorage.Config
	Sessions     []agentstorage.SessionMeta
	Session      *agentstorage.Session
	Events       []types.AgentEvent
	EventsOffset int64
	Messages     []agentstorage.Message
	Err          error
	FilesWritten []string
	ResolvedHome string
}

func Run(t *testing.T, req *Request) (*Response, error) {
	applyEnv(req.Env)
	defer restoreEnv(req.Env)

	store, home, err := openStore(t, req)
	if err != nil {
		return &Response{Err: err, ResolvedHome: home}, err
	}

	resp := &Response{Store: store, ResolvedHome: home}
	tracker := newFileTracker(home)

	switch req.Operation {
	case "home":
		switch req.Action {
		case "env_override":
			resp.ResolvedHome = store.Home()
		case "creates_dirs":
			ev := sampleEvent("bootstrap")
			if err := store.AppendEvent(req.Runner, req.SessionID, ev); err != nil {
				resp.Err = err
				return resp, err
			}
			scanFileTracker(tracker)
			resp.FilesWritten = tracker.paths
		default:
			return nil, fmt.Errorf("unknown home action %q", req.Action)
		}

	case "config":
		switch req.Action {
		case "load_missing":
			cfg, err := store.Config()
			if err != nil {
				resp.Err = err
				return resp, err
			}
			resp.Config = cfg
		case "save_reload":
			if err := store.SaveConfig(req.Config); err != nil {
				resp.Err = err
				return resp, err
			}
			scanFileTracker(tracker)
			cfg, err := store.Config()
			if err != nil {
				resp.Err = err
				return resp, err
			}
			resp.Config = cfg
			resp.FilesWritten = tracker.paths
		default:
			return nil, fmt.Errorf("unknown config action %q", req.Action)
		}

	case "session":
		switch req.Action {
		case "create_get_update":
			meta := agentstorage.SessionMeta{
				Runner:          req.Runner,
				SessionID:       req.SessionID,
				RunnerSessionID: req.RunnerSess,
				Status:          "running",
				Model:           req.Model,
			}
			if err := store.CreateSession(req.Runner, req.SessionID, meta); err != nil {
				resp.Err = err
				return resp, err
			}
			sess, err := store.GetSession(req.Runner, req.SessionID)
			if err != nil {
				resp.Err = err
				return resp, err
			}
			resp.Session = sess
			if err := store.UpdateSessionStatus(req.Runner, req.SessionID, req.Status); err != nil {
				resp.Err = err
				return resp, err
			}
			sess2, err := store.GetSession(req.Runner, req.SessionID)
			if err != nil {
				resp.Err = err
				return resp, err
			}
			resp.Session = sess2
		case "list_by_runner":
			m1 := agentstorage.SessionMeta{Runner: req.Runner, SessionID: req.SessionID, Status: "running"}
			m2 := agentstorage.SessionMeta{Runner: req.OtherRunner, SessionID: req.OtherSessID, Status: "running"}
			if err := store.CreateSession(req.Runner, req.SessionID, m1); err != nil {
				resp.Err = err
				return resp, err
			}
			if err := store.CreateSession(req.OtherRunner, req.OtherSessID, m2); err != nil {
				resp.Err = err
				return resp, err
			}
			list, err := store.ListSessions(req.Runner)
			if err != nil {
				resp.Err = err
				return resp, err
			}
			resp.Sessions = list
		case "get_missing":
			sess, err := store.GetSession(req.Runner, req.SessionID)
			resp.Session = sess
			resp.Err = err
		case "workspace_roundtrip":
			meta := agentstorage.SessionMeta{
				Runner:    req.Runner,
				SessionID: req.SessionID,
				Status:    "running",
				Workspace: req.Workspace,
			}
			if err := store.CreateSession(req.Runner, req.SessionID, meta); err != nil {
				resp.Err = err
				return resp, err
			}
			sess, err := store.GetSession(req.Runner, req.SessionID)
			if err != nil {
				resp.Err = err
				return resp, err
			}
			resp.Session = sess
		default:
			return nil, fmt.Errorf("unknown session action %q", req.Action)
		}

	case "events":
		switch req.Action {
		case "append_read_start":
			for _, ev := range req.Events {
				if err := store.AppendEvent(req.Runner, req.SessionID, ev); err != nil {
					resp.Err = err
					return resp, err
				}
			}
			events, offset, err := store.ReadEvents(req.Runner, req.SessionID, 0)
			if err != nil {
				resp.Err = err
				return resp, err
			}
			resp.Events = events
			resp.EventsOffset = offset
		case "read_after_offset":
			for _, ev := range req.Events {
				if err := store.AppendEvent(req.Runner, req.SessionID, ev); err != nil {
					resp.Err = err
					return resp, err
				}
			}
			_, firstOffset, err := store.ReadEvents(req.Runner, req.SessionID, 0)
			if err != nil {
				resp.Err = err
				return resp, err
			}
			events, offset, err := store.ReadEvents(req.Runner, req.SessionID, firstOffset)
			if err != nil {
				resp.Err = err
				return resp, err
			}
			resp.Events = events
			resp.EventsOffset = offset
		case "read_empty":
			events, offset, err := store.ReadEvents(req.Runner, req.SessionID, 0)
			if err != nil {
				resp.Err = err
				return resp, err
			}
			resp.Events = events
			resp.EventsOffset = offset
		default:
			return nil, fmt.Errorf("unknown events action %q", req.Action)
		}

	case "messages":
		switch req.Action {
		case "append_list":
			for _, text := range req.MessageText {
				if _, err := store.AppendMessage(req.Runner, req.SessionID, text); err != nil {
					resp.Err = err
					return resp, err
				}
			}
			msgs, err := store.ListMessages(req.Runner, req.SessionID)
			if err != nil {
				resp.Err = err
				return resp, err
			}
			resp.Messages = msgs
		case "pop_fifo":
			for _, text := range req.MessageText {
				if _, err := store.AppendMessage(req.Runner, req.SessionID, text); err != nil {
					resp.Err = err
					return resp, err
				}
			}
			msgs, err := store.PopMessages(req.Runner, req.SessionID)
			if err != nil {
				resp.Err = err
				return resp, err
			}
			resp.Messages = msgs
		case "pop_empty":
			msgs, err := store.PopMessages(req.Runner, req.SessionID)
			if err != nil {
				resp.Err = err
				return resp, err
			}
			resp.Messages = msgs
		default:
			return nil, fmt.Errorf("unknown messages action %q", req.Action)
		}

	case "isolation":
		switch req.Action {
		case "writes_under_home":
			meta := agentstorage.SessionMeta{Runner: req.Runner, SessionID: req.SessionID, Status: "running"}
			if err := store.CreateSession(req.Runner, req.SessionID, meta); err != nil {
				resp.Err = err
				return resp, err
			}
			if err := store.SaveConfig(req.Config); err != nil {
				resp.Err = err
				return resp, err
			}
			if err := store.AppendEvent(req.Runner, req.SessionID, sampleEvent("iso")); err != nil {
				resp.Err = err
				return resp, err
			}
			if _, err := store.AppendMessage(req.Runner, req.SessionID, "hello"); err != nil {
				resp.Err = err
				return resp, err
			}
			scanFileTracker(tracker)
			resp.FilesWritten = tracker.paths
		default:
			return nil, fmt.Errorf("unknown isolation action %q", req.Action)
		}

	default:
		return nil, fmt.Errorf("unknown operation %q", req.Operation)
	}

	return resp, nil
}

func sampleEvent(text string) types.AgentEvent {
	return types.AgentEvent{
		Type: types.ActionMessage,
		Text: text,
	}
}
```