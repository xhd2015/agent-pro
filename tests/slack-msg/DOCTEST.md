# slack-msg CLI (send | history | listen) Doctests

Doc-style tests for the unified first-class CLI at `cmd/slack-msg` with subcommands
`send`, `history`, and `listen`. Ports contracts from `tests/slack-send-cli` and
`tests/slack-listen` (updated argv and `--token` instead of `--bot-token`), and
adds history coverage (oldest→newest, `--json`, `--thread`). Uses `pkgs/slackutil`
for config/token/channel resolution and `slacktest` + `SLACK_API_URL` for unit paths.

# DSN (Domain Specific Notion)

**Participants**

- **Caller** — user or doctest harness invoking the `slack-msg` binary.
- **slack-msg CLI** — root dispatcher for `send` | `history` | `listen`; top-level
  `-h`/`--help` lists the three commands; unknown command → stderr + exit 1.
- **send** — posts via `chat.postMessage`; requires exactly one positional MESSAGE;
  flags `--token`, `--channel`, `--config`, `--thread`, `-h`/`--help`; token/channel
  from CLI → env → config; prints Sending / Using config / OK lines on success.
- **history** — fetches `conversations.history` or `conversations.replies` (`--thread`);
  prints human lines oldest→newest (`[ts] user: text`) or `--json` document in the
  same order; flags `--token`, `--channel`, `--config`, `--limit`, `--json`, `--thread`.
- **listen** — Socket Mode inbound bridge → filter → agent-run → `chat.postMessage`
  thread reply; flags use `--token` (bot) and `--app-token`; singleton lock file.
- **SlackConfig JSON** — optional file with `botToken`, `appToken`, `defaultChannelId`,
  `knownChannels`; loaded only when `--config` or `SLACK_CONFIG` is set (no cwd walk).
- **Slack Web API** — real `slack.com` in integration; `slacktest` when `SLACK_API_URL`
  is set (`conversations.list`, `chat.postMessage`, `conversations.history`,
  `conversations.replies`, Socket Mode).
- **agent-run** — external runner for listen; mocked via `SLACK_LISTEN_AGENT_RUN`.
- **Singleton lock** — second listen instance exits non-zero with
  `another slack-msg is already running`.
- **Test harness** — builds `./cmd/slack-msg` once per session; quick-exit CLI runs
  for send/history/root; daemon probes for listen unit paths.

**Behaviors**

- `slack-msg -h|--help` — lists `send`, `history`, `listen`; exit 0; trailing `\n`.
- `slack-msg <unknown>` — stderr error; exit 1.
- `slack-msg send [options] MESSAGE` — message required (reject 0 or 2+); missing
  token → `bot token required`; missing channel → `channel required`; success
  stdout ends with `OK ts=... channel=...` and trailing `\n`; API failure →
  `send failed:` prefix, exit 1, no `OK`.
- `slack-msg history [options] [CHANNEL]` — channel via flag/positional/env/config;
  human or JSON oldest→newest; API failure → `history failed:` prefix, exit 1.
- `slack-msg listen [options]` — missing bot/app token exit 1 before connect;
  config `(none)` vs absolute path logging; filter bot-self / DM / mention /
  allowFrom / channel; thread vs stateless session routing; lock singleton;
  reply with `thread_ts` and optional `--reply-prefix`.

## Version

0.0.2

## Decision Tree

```
tests/slack-msg/
├── DOCTEST.md
├── SETUP.md                           # build binary, slacktest, mock agent, daemon helpers
├── testdata/
│   ├── valid-config.json
│   ├── empty-token-config.json
│   ├── empty-app-token-config.json
│   └── default-channel-name.json
├── help/                              # top-level -h / --help
│   ├── SETUP.md
│   ├── short-flag/
│   └── long-flag/
├── unknown-command/
│   ├── SETUP.md
│   └── not-recognized/
├── send/
│   ├── SETUP.md
│   ├── help/
│   ├── message-errors/
│   ├── token-errors/
│   ├── channel-errors/
│   ├── config-errors/
│   ├── config-none/
│   ├── config-with-default/
│   ├── channel-resolve/               # label: unit
│   ├── send-success/                  # label: unit (incl. --thread)
│   ├── send-errors/                   # label: unit
│   └── integration/                   # label: integration, slow
├── history/
│   ├── SETUP.md
│   ├── help/
│   ├── token-errors/
│   ├── channel-errors/
│   ├── order-oldest-first/            # label: unit
│   ├── json-output/                   # label: unit
│   ├── limit/                         # label: unit
│   ├── thread-replies/                # label: unit
│   ├── channel-resolve/               # label: unit
│   └── history-errors/                # label: unit
└── listen/
    ├── SETUP.md
    ├── help/
    ├── token-errors/                  # --token / --app-token
    ├── lock/                          # label: unit
    ├── config/
    ├── filter/                        # label: unit
    ├── session-routing/               # label: unit
    ├── reply/                         # label: unit
    └── integration/                   # label: integration, slow
```

Parameter ranking (most → least significant):

1. **Command / outcome** — top-level help vs unknown vs send vs history vs listen
2. **Within command** — help vs validation-error vs success/unit path vs integration
3. **Credential / channel source** — CLI flags vs env vs `--config` JSON
4. **Backend** — slacktest (`SLACK_API_URL`) vs live Slack vs no network

## Test Index

| # | Leaf | Labels | Description |
|---|------|--------|-------------|
| 1 | `help/short-flag` | (default) | Top-level `-h` lists send/history/listen; exit 0 |
| 2 | `help/long-flag` | (default) | Top-level `--help` same contract |
| 3 | `unknown-command/not-recognized` | (default) | Unknown subcommand → stderr; exit 1 |
| 4 | `send/help/short-flag` | (default) | `send -h` usage; exit 0 |
| 5 | `send/help/long-flag` | (default) | `send --help` same as `-h` |
| 6 | `send/message-errors/missing-message` | (default) | No MESSAGE → `message required` |
| 7 | `send/message-errors/multiple-positionals` | (default) | Two positionals → `exactly one message required` |
| 8 | `send/token-errors/missing-token` | (default) | No bot token → `bot token required` |
| 9 | `send/channel-errors/missing-channel` | (default) | No channel → `channel required` |
| 10 | `send/config-errors/bad-config-path` | (default) | Missing config file → `failed to load config` |
| 11 | `send/config-errors/empty-bot-token` | (default) | Empty config botToken → `botToken is empty in` |
| 12 | `send/config-none/stdout-line` | (default) | CLI flags only → `Using config from: (none)` |
| 13 | `send/channel-resolve/api-name-with-hash` | unit | `#general` via conversations.list |
| 14 | `send/channel-resolve/api-name-without-hash` | unit | `general` normalized via API |
| 15 | `send/channel-resolve/direct-channel-id` | unit | `C...` passthrough |
| 16 | `send/channel-resolve/direct-dm-id` | unit | `D...` passthrough |
| 17 | `send/channel-resolve/direct-group-id` | unit | `G...` passthrough |
| 18 | `send/channel-resolve/config-known-channels` | unit | knownChannels fast path |
| 19 | `send/send-success/cli-flags` | unit | `--token --channel` + message |
| 20 | `send/send-success/channel-by-id` | unit | Channel ID + custom message |
| 21 | `send/send-success/multi-word-message` | unit | Single multi-word MESSAGE |
| 22 | `send/send-success/env-token` | unit | `SLACK_BOT_TOKEN` fallback |
| 23 | `send/send-success/env-channel` | unit | `SLACK_CHANNEL` fallback |
| 24 | `send/send-success/thread-ts` | unit | `--thread TS` post in thread |
| 25 | `send/send-errors/channel-not-found` | unit | Unknown name → `channel not found` |
| 26 | `send/send-errors/api-post-failed` | unit | PostMessage error → `send failed:` |
| 27 | `send/config-with-default/message-only` | unit | `--config` + message uses defaults |
| 28 | `send/config-with-default/override-channel` | unit | CLI `--channel` wins over config |
| 29 | `send/config-with-default/default-channel-name` | unit | Name-shaped defaultChannelId |
| 30 | `send/integration/live-explicit-config` | integration, slow | Live send via repo config |
| 31 | `history/help/short-flag` | (default) | `history -h`; exit 0 |
| 32 | `history/help/long-flag` | (default) | `history --help` |
| 33 | `history/token-errors/missing-token` | (default) | No bot token → `bot token required` |
| 34 | `history/channel-errors/missing-channel` | (default) | No channel → `channel required` |
| 35 | `history/order-oldest-first/multi-message` | unit | Human lines oldest→newest |
| 36 | `history/json-output/chronological` | unit | JSON messages oldest→newest |
| 37 | `history/limit/limits-results` | unit | `--limit 2` trims to 2 chronological lines |
| 38 | `history/thread-replies/prints-replies` | unit | `--thread` uses replies path |
| 39 | `history/channel-resolve/direct-channel-id` | unit | Direct channel ID history |
| 40 | `history/channel-resolve/api-name-with-hash` | unit | `#general` resolve then history |
| 41 | `history/history-errors/channel-not-found` | unit | Unknown channel → `history failed:` + not found |
| 42 | `history/history-errors/api-failed` | unit | API error → `history failed:` |
| 43 | `listen/help/short-flag` | (default) | `listen -h`; `--token` not `--bot-token` |
| 44 | `listen/help/long-flag` | (default) | `listen --help` |
| 45 | `listen/token-errors/missing-bot-token` | (default) | No bot token → `bot token required` |
| 46 | `listen/token-errors/missing-app-token` | (default) | No app token → `app token required` |
| 47 | `listen/lock/already-running` | unit | Second instance: `another slack-msg is already running` |
| 48 | `listen/config/none-stdout-line` | unit | `Using config from: (none)` |
| 49 | `listen/config/explicit-path` | unit | Absolute `--config` path logged |
| 50 | `listen/config/bad-config-path` | (default) | Missing config → `failed to load config` |
| 51 | `listen/filter/ignore-bot-self` | unit | Bot-authored message ignored |
| 52 | `listen/filter/dm-always-processed` | unit | DM without mention processed |
| 53 | `listen/filter/channel-requires-mention` | unit | Channel msg without mention ignored |
| 54 | `listen/filter/channel-no-require-mention` | unit | `--no-require-mention` processes channel |
| 55 | `listen/filter/allow-from-blocked` | unit | User not in allowFrom ignored |
| 56 | `listen/filter/allow-from-wildcard` | unit | Default allow processes any user |
| 57 | `listen/filter/channel-filter-excludes` | unit | `--channel` filter drops others |
| 58 | `listen/session-routing/thread-first-run` | unit | First msg → `run --keep-tty --session` |
| 59 | `listen/session-routing/thread-follow-up-send` | unit | Follow-up → `send <session-id>` |
| 60 | `listen/session-routing/stateless-each-run` | unit | Stateless → every msg `run` |
| 61 | `listen/reply/posts-in-thread` | unit | PostMessage includes `thread_ts` |
| 62 | `listen/reply/reply-prefix` | unit | `--reply-prefix` applied |
| 63 | `listen/integration/live-socket-reply` | integration, slow | Live Socket Mode connect probe |

## How to Run

```sh
# Structure validation
doctest vet ./tests/slack-msg

# Default CI — help + validation (no labels; no network)
doctest test ./tests/slack-msg

# Unit suite — slacktest + mock agent-run (RED until cmd/slack-msg lands)
doctest test --label unit ./tests/slack-msg

# Live Slack
doctest test --label integration ./tests/slack-msg
doctest test --label slow ./tests/slack-msg

# Single leaf
doctest test -v ./tests/slack-msg/help/short-flag
doctest test -v ./tests/slack-msg/history/order-oldest-first/multi-message
```

**Implementer note (RED until done):** Requires `cmd/slack-msg` with `send` |
`history` | `listen`, unified `--token` bot flag, `SLACK_API_URL` →
`slack.OptionAPIURL`, `SLACK_LISTEN_AGENT_RUN` mock path, chronological history
output, and lock message `another slack-msg is already running`.

```go
import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/slacktest"
)

const (
	slackTestToken         = "xoxb-slacktest-token"
	slackTestBotToken      = slackTestToken
	slackTestAppToken      = "xapp-slacktest-token"
	slackTestChannelID     = "C0ALE44K5J6"
	slackTestDMChannelID   = "D024BE91L"
	slackTestUserID        = "W012A3CDE"
	slackTestOtherUserID   = "W0OTHERUSR"
	slackTestBotUserID     = "U023BECGF"
	slackTestTeamID        = "T024BE7LD"
	defaultAgentReply      = "mock agent reply"
	envAgentRun            = "SLACK_LISTEN_AGENT_RUN"
	envAgentLog            = "SLACK_LISTEN_AGENT_LOG"
)

// History fixture: API returns newest-first; CLI must print oldest→newest.
var historyMessagesNewestFirst = []map[string]any{
	{"type": "message", "user": "U_NEWER", "text": "second message", "ts": "1710000002.000200"},
	{"type": "message", "user": "U_OLDER", "text": "first message", "ts": "1710000001.000100"},
	{"type": "message", "user": "U_NEWEST", "text": "third message", "ts": "1710000003.000300"},
}

// Thread replies fixture (newest-first in mock; CLI chronological).
var threadRepliesNewestFirst = []map[string]any{
	{"type": "message", "user": "U_R2", "text": "reply two", "ts": "1710001000.000300", "thread_ts": "1710001000.000100"},
	{"type": "message", "user": "U_R1", "text": "reply one", "ts": "1710001000.000200", "thread_ts": "1710001000.000100"},
	{"type": "message", "user": "U_PARENT", "text": "parent", "ts": "1710001000.000100", "thread_ts": "1710001000.000100"},
}

type InjectedEvent struct {
	Kind     string // message | app_mention | dm
	Channel  string
	User     string
	Text     string
	TS       string
	ThreadTS string
}

type CapturedPost struct {
	Channel  string
	Text     string
	ThreadTS string
}

type Request struct {
	RepoRoot      string
	WorkDir       string
	Bin           string
	Args          []string
	ConfigPath    string
	ConfigFixture string
	ConfigInline  string
	SlackAPIURL   string
	Env           []string
	ClearSlackEnv bool

	// Listen-oriented fields (also used when ListenMode).
	ListenMode     bool
	BotToken       string
	AppToken       string
	LockFile       string
	Daemon         bool
	InjectEvents   []InjectedEvent
	WantAgentCalls int // -1 => len(InjectEvents)
	AgentLogPath   string
	MockAgentPath  string
	ObserveTimeout time.Duration
	Posts          *[]CapturedPost
	SecondInstance bool

	// History server variants: default, historyFail.
	HistoryAPIFail bool
}

type Response struct {
	ExitCode         int
	Stdout           string
	Stderr           string
	AgentInvocations []string
	PostMessages     []CapturedPost
	SecondExitCode   int
	SecondStderr     string
	SecondStdout     string
}

var slackTestChannels = []slack.Channel{
	{
		GroupConversation: slack.GroupConversation{
			Conversation: slack.Conversation{ID: "C0ALE44K5J6"},
			Name:         "general",
		},
	},
	{
		GroupConversation: slack.GroupConversation{
			Conversation: slack.Conversation{ID: "C0OTHERCHAN"},
			Name:         "random",
		},
	},
}

func findModuleRoot() (string, error) {
	start := DOCTEST_ROOT
	for dir := start; ; dir = filepath.Dir(dir) {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		if filepath.Dir(dir) == dir {
			return "", fmt.Errorf("go.mod not found above %s", start)
		}
	}
}

var (
	buildOnce sync.Once
	builtBin  string
	buildErr  error
)

func sessionCacheDir() string {
	return filepath.Join(os.TempDir(), "slack-msg-doctest-"+DOCTEST_SESSION_ID)
}

func withFileLock(lockPath string, fn func() error) error {
	if err := os.MkdirAll(filepath.Dir(lockPath), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		return err
	}
	defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
	return fn()
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func buildSlackMsg(t *testing.T) (string, error) {
	t.Helper()
	buildOnce.Do(func() {
		repoRoot, err := findModuleRoot()
		if err != nil {
			buildErr = err
			return
		}
		cacheDir := sessionCacheDir()
		if err := os.MkdirAll(cacheDir, 0o755); err != nil {
			buildErr = err
			return
		}
		bin := filepath.Join(cacheDir, "slack-msg")
		ready := filepath.Join(cacheDir, "bin.ready")
		lock := filepath.Join(cacheDir, "build.lock")
		buildErr = withFileLock(lock, func() error {
			if fileExists(ready) && fileExists(bin) {
				builtBin = bin
				return nil
			}
			ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, "go", "build", "-o", bin, "./cmd/slack-msg")
			cmd.Dir = repoRoot
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("go build ./cmd/slack-msg: %w\n%s", err, stderr.String())
			}
			if err := os.WriteFile(ready, []byte("ok"), 0o644); err != nil {
				return err
			}
			builtBin = bin
			return nil
		})
	})
	return builtBin, buildErr
}

func resolveFixturePath(name string) string {
	candidates := []string{
		filepath.Join(DOCTEST_ROOT, "testdata", name),
		filepath.Join(DOCTEST_ROOT, name),
	}
	for _, c := range candidates {
		if fileExists(c) {
			return c
		}
	}
	return filepath.Join(DOCTEST_ROOT, "testdata", name)
}

func materializeConfig(t *testing.T, req *Request) error {
	t.Helper()
	if req.ConfigInline == "" && req.ConfigFixture == "" {
		return nil
	}
	path := req.ConfigPath
	if path == "" {
		if req.WorkDir == "" {
			req.WorkDir = t.TempDir()
		}
		path = filepath.Join(req.WorkDir, "slack-config.json")
	}
	var data []byte
	var err error
	if req.ConfigInline != "" {
		data = []byte(req.ConfigInline)
	} else {
		data, err = os.ReadFile(resolveFixturePath(req.ConfigFixture))
		if err != nil {
			return fmt.Errorf("read fixture %s: %w", req.ConfigFixture, err)
		}
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write config %s: %w", path, err)
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	req.ConfigPath = abs
	return nil
}

func mergeEnv(base []string, extra ...string) []string {
	seen := make(map[string]struct{})
	out := make([]string, 0, len(base)+len(extra))
	for _, e := range base {
		if i := strings.IndexByte(e, '='); i > 0 {
			seen[e[:i]] = struct{}{}
		}
		out = append(out, e)
	}
	for _, e := range extra {
		if i := strings.IndexByte(e, '='); i > 0 {
			k := e[:i]
			if _, ok := seen[k]; ok {
				continue
			}
			seen[k] = struct{}{}
		}
		out = append(out, e)
	}
	return out
}

func withoutEnvKeys(env []string, keys ...string) []string {
	keySet := make(map[string]struct{}, len(keys))
	for _, k := range keys {
		keySet[k] = struct{}{}
	}
	out := make([]string, 0, len(env))
	for _, e := range env {
		if i := strings.IndexByte(e, '='); i > 0 {
			if _, drop := keySet[e[:i]]; drop {
				continue
			}
		}
		out = append(out, e)
	}
	return out
}

type slackTestMode int

const (
	slackTestDefault slackTestMode = iota
	slackTestPostFail
	slackTestHistoryFail
)

var (
	// Per-process slacktest servers only. Do not cache API URLs on disk across
	// go-test packages — that reuses dead servers after the creating process exits.
	// slacktest.NewTestServer registers into an internal hub so the listener stays live.
	slackTestOnce            sync.Once
	slackTestURL             string
	slackTestErr             error
	slackTestPostFailOnce    sync.Once
	slackTestPostFailURL     string
	slackTestPostFailErr     error
	slackTestHistoryFailOnce sync.Once
	slackTestHistoryFailURL  string
	slackTestHistoryFailErr  error
)

func conversationsListHandler(channels []slack.Channel) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := struct {
			slack.SlackResponse
			Channels         []slack.Channel `json:"channels"`
			ResponseMetadata struct {
				NextCursor string `json:"next_cursor"`
			} `json:"response_metadata"`
		}{
			SlackResponse: slack.SlackResponse{Ok: true},
			Channels:      channels,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}
}

func limitFromRequest(r *http.Request) int {
	_ = r.ParseForm()
	limitStr := r.Form.Get("limit")
	if limitStr == "" {
		return 0
	}
	n, err := strconv.Atoi(limitStr)
	if err != nil || n <= 0 {
		return 0
	}
	return n
}

func applyLimitNewestFirst(msgs []map[string]any, limit int) []map[string]any {
	if limit <= 0 || limit >= len(msgs) {
		return msgs
	}
	// msgs are newest-first; take the first `limit` (newest N).
	return msgs[:limit]
}

func historyHandler(fail bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if fail {
			_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "error": "internal_error"})
			return
		}
		msgs := applyLimitNewestFirst(historyMessagesNewestFirst, limitFromRequest(r))
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":       true,
			"messages": msgs,
			"has_more": false,
		})
	}
}

func repliesHandler(fail bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if fail {
			_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "error": "internal_error"})
			return
		}
		msgs := applyLimitNewestFirst(threadRepliesNewestFirst, limitFromRequest(r))
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":       true,
			"messages": msgs,
			"has_more": false,
		})
	}
}

func startSlackTestServer(mode slackTestMode) (string, error) {
	sts := slacktest.NewTestServer(func(s slacktest.Customize) {
		s.Handle("/conversations.list", conversationsListHandler(slackTestChannels))
		s.Handle("/conversations.history", historyHandler(mode == slackTestHistoryFail))
		s.Handle("/conversations.replies", repliesHandler(mode == slackTestHistoryFail))
		if mode == slackTestPostFail {
			s.Handle("/chat.postMessage", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]any{
					"ok":    false,
					"error": "invalid_auth",
				})
			})
		}
	})
	sts.Start()
	return sts.GetAPIURL(), nil
}

func ensureSlackTestServer(t *testing.T) (string, error) {
	t.Helper()
	slackTestOnce.Do(func() {
		url, err := startSlackTestServer(slackTestDefault)
		if err != nil {
			slackTestErr = err
			return
		}
		slackTestURL = url
	})
	return slackTestURL, slackTestErr
}

func ensureSlackTestServerPostFail(t *testing.T) (string, error) {
	t.Helper()
	slackTestPostFailOnce.Do(func() {
		url, err := startSlackTestServer(slackTestPostFail)
		if err != nil {
			slackTestPostFailErr = err
			return
		}
		slackTestPostFailURL = url
	})
	return slackTestPostFailURL, slackTestPostFailErr
}

func ensureSlackTestServerHistoryFail(t *testing.T) (string, error) {
	t.Helper()
	slackTestHistoryFailOnce.Do(func() {
		url, err := startSlackTestServer(slackTestHistoryFail)
		if err != nil {
			slackTestHistoryFailErr = err
			return
		}
		slackTestHistoryFailURL = url
	})
	return slackTestHistoryFailURL, slackTestHistoryFailErr
}

func writeMockAgent(t *testing.T, dir, logPath string) string {
	t.Helper()
	path := filepath.Join(dir, "mock-agent-run")
	script := fmt.Sprintf(`#!/bin/sh
echo "INVOCATION $*" >> %q
printf %%s %q
`, logPath, defaultAgentReply)
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("write mock agent: %v", err)
	}
	return path
}

func newListenSlackTestServer(t *testing.T, posts *[]CapturedPost) (*slacktest.Server, string) {
	t.Helper()
	sts := slacktest.NewTestServer(func(s slacktest.Customize) {
		s.Handle("/chat.postMessage", func(w http.ResponseWriter, r *http.Request) {
			data, _ := io.ReadAll(r.Body)
			values, _ := url.ParseQuery(string(data))
			if posts != nil {
				*posts = append(*posts, CapturedPost{
					Channel:  values.Get("channel"),
					Text:     values.Get("text"),
					ThreadTS: values.Get("thread_ts"),
				})
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"ok":      true,
				"channel": values.Get("channel"),
				"ts":      fmt.Sprintf("%d.000001", time.Now().Unix()),
				"text":    values.Get("text"),
			})
		})
	})
	sts.Start()
	t.Cleanup(sts.Stop)
	return sts, sts.GetAPIURL()
}

func socketModeEnvelope(envelopeID string, eventType string, inner map[string]any) (string, error) {
	inner["type"] = eventType
	payload := map[string]any{
		"token":      slackTestBotToken,
		"team_id":    slackTestTeamID,
		"type":       "event_callback",
		"event":      inner,
		"event_id":   "Ev" + envelopeID,
		"event_time": time.Now().Unix(),
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	envelope := map[string]any{
		"envelope_id":              envelopeID,
		"payload":                  json.RawMessage(payloadBytes),
		"type":                     "events_api",
		"accepts_response_payload": false,
	}
	b, err := json.Marshal(envelope)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func injectEvent(sts *slacktest.Server, ev InjectedEvent) error {
	channel := ev.Channel
	user := ev.User
	if user == "" {
		user = slackTestUserID
	}
	ts := ev.TS
	if ts == "" {
		ts = fmt.Sprintf("%d.000100", time.Now().Unix())
	}
	kind := ev.Kind
	if kind == "" {
		kind = "message"
	}
	if kind == "dm" {
		channel = slackTestDMChannelID
		kind = "message"
	}
	inner := map[string]any{
		"user":    user,
		"text":    ev.Text,
		"ts":      ts,
		"channel": channel,
	}
	if ev.ThreadTS != "" {
		inner["thread_ts"] = ev.ThreadTS
	}
	eventType := "message"
	if kind == "app_mention" {
		eventType = "app_mention"
	}
	envelopeID := "Env-" + strings.ReplaceAll(ts, ".", "")
	msg, err := socketModeEnvelope(envelopeID, eventType, inner)
	if err != nil {
		return err
	}
	sts.SendToWebsocket(msg)
	return nil
}

func readAgentInvocations(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			out = append(out, line)
		}
	}
	return out, nil
}

func waitForAgentLog(path string, wantMin int, timeout time.Duration) ([]string, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		lines, err := readAgentInvocations(path)
		if err != nil {
			return nil, err
		}
		if len(lines) >= wantMin {
			return lines, nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	lines, _ := readAgentInvocations(path)
	return lines, fmt.Errorf("timeout waiting for agent log %s (got %d, want %d)", path, len(lines), wantMin)
}

func waitForPosts(posts *[]CapturedPost, wantMin int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if posts != nil && len(*posts) >= wantMin {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	n := 0
	if posts != nil {
		n = len(*posts)
	}
	return fmt.Errorf("timeout waiting for post messages (got %d, want %d)", n, wantMin)
}

func isSubcommand(s string) bool {
	switch s {
	case "send", "history", "listen":
		return true
	default:
		return false
	}
}

func insertConfigAfterSubcommand(args []string, configPath string) []string {
	cfg := []string{"--config", configPath}
	if len(args) > 0 && isSubcommand(args[0]) {
		out := make([]string, 0, len(args)+2)
		out = append(out, args[0])
		out = append(out, cfg...)
		out = append(out, args[1:]...)
		return out
	}
	return append(cfg, args...)
}

func defaultListenArgs(req *Request) []string {
	args := []string{"listen"}
	if req.BotToken != "" {
		args = append(args, "--token", req.BotToken)
	}
	if req.AppToken != "" {
		args = append(args, "--app-token", req.AppToken)
	}
	if req.ConfigPath != "" {
		args = append(args, "--config", req.ConfigPath)
	}
	if req.LockFile != "" {
		args = append(args, "--lock-file", req.LockFile)
	}
	args = append(args, req.Args...)
	return args
}

func runSimple(t *testing.T, req *Request) (*Response, error) {
	if req.WorkDir == "" {
		req.WorkDir = t.TempDir()
	}
	if err := materializeConfig(t, req); err != nil {
		return nil, err
	}
	if req.HistoryAPIFail {
		apiURL, err := ensureSlackTestServerHistoryFail(t)
		if err != nil {
			return nil, err
		}
		req.SlackAPIURL = apiURL
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, req.Bin, req.Args...)
	cmd.Dir = req.WorkDir

	env := os.Environ()
	if req.ClearSlackEnv {
		env = withoutEnvKeys(env, "SLACK_BOT_TOKEN", "SLACK_APP_TOKEN", "SLACK_CHANNEL", "SLACK_CONFIG", "SLACK_API_URL", envAgentRun, envAgentLog)
	}
	if req.SlackAPIURL != "" {
		env = withoutEnvKeys(env, "SLACK_API_URL")
		env = append(env, "SLACK_API_URL="+req.SlackAPIURL)
	}
	env = mergeEnv(env, req.Env...)
	cmd.Env = env

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	resp := &Response{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}
	if err == nil {
		resp.ExitCode = 0
		return resp, nil
	}
	if ctx.Err() != nil {
		return resp, ctx.Err()
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		resp.ExitCode = exitErr.ExitCode()
		return resp, nil
	}
	return resp, err
}

func runListenQuick(t *testing.T, req *Request) (*Response, error) {
	if req.WorkDir == "" {
		req.WorkDir = t.TempDir()
	}
	if err := materializeConfig(t, req); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, req.Bin, defaultListenArgs(req)...)
	cmd.Dir = req.WorkDir
	env := os.Environ()
	if req.ClearSlackEnv {
		env = withoutEnvKeys(env, "SLACK_BOT_TOKEN", "SLACK_APP_TOKEN", "SLACK_CONFIG", "SLACK_API_URL", envAgentRun, envAgentLog)
	}
	if req.SlackAPIURL != "" {
		env = withoutEnvKeys(env, "SLACK_API_URL")
		env = append(env, "SLACK_API_URL="+req.SlackAPIURL)
	}
	if req.MockAgentPath != "" {
		env = mergeEnv(env, envAgentRun+"="+req.MockAgentPath)
	}
	if req.AgentLogPath != "" {
		env = mergeEnv(env, envAgentLog+"="+req.AgentLogPath)
	}
	env = mergeEnv(env, req.Env...)
	cmd.Env = env
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	resp := &Response{Stdout: stdout.String(), Stderr: stderr.String()}
	if err == nil {
		resp.ExitCode = 0
		return resp, nil
	}
	if ctx.Err() != nil {
		return resp, ctx.Err()
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		resp.ExitCode = exitErr.ExitCode()
		return resp, nil
	}
	return resp, err
}

func runDaemon(t *testing.T, req *Request) (*Response, error) {
	if req.WorkDir == "" {
		req.WorkDir = t.TempDir()
	}
	if err := materializeConfig(t, req); err != nil {
		return nil, err
	}
	if req.AgentLogPath == "" {
		req.AgentLogPath = filepath.Join(req.WorkDir, "agent.log")
	}
	if req.MockAgentPath == "" {
		req.MockAgentPath = writeMockAgent(t, req.WorkDir, req.AgentLogPath)
	}
	var posts []CapturedPost
	if req.Posts != nil {
		posts = *req.Posts
	}
	sts, apiURL := newListenSlackTestServer(t, &posts)
	req.SlackAPIURL = apiURL
	timeout := req.ObserveTimeout
	if timeout == 0 {
		timeout = 8 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout+10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, req.Bin, defaultListenArgs(req)...)
	cmd.Dir = req.WorkDir
	env := os.Environ()
	if req.ClearSlackEnv {
		env = withoutEnvKeys(env, "SLACK_BOT_TOKEN", "SLACK_APP_TOKEN", "SLACK_CONFIG", "SLACK_API_URL", envAgentRun, envAgentLog)
	}
	env = mergeEnv(env,
		"SLACK_API_URL="+req.SlackAPIURL,
		envAgentRun+"="+req.MockAgentPath,
		envAgentLog+"="+req.AgentLogPath,
	)
	env = mergeEnv(env, req.Env...)
	cmd.Env = env
	stdoutPipe, _ := cmd.StdoutPipe()
	stderrPipe, _ := cmd.StderrPipe()
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	var stdoutBuf, stderrBuf bytes.Buffer
	done := make(chan struct{})
	go func() {
		io.Copy(&stdoutBuf, stdoutPipe)
		io.Copy(&stderrBuf, stderrPipe)
		close(done)
	}()
	time.Sleep(500 * time.Millisecond)
	for i, ev := range req.InjectEvents {
		if err := injectEvent(sts, ev); err != nil {
			_ = cmd.Process.Kill()
			<-done
			return nil, err
		}
		if i < len(req.InjectEvents)-1 {
			time.Sleep(200 * time.Millisecond)
		}
	}
	wantAgent := req.WantAgentCalls
	if wantAgent < 0 {
		wantAgent = len(req.InjectEvents)
	}
	var invocations []string
	var agentErr error
	if wantAgent > 0 {
		invocations, agentErr = waitForAgentLog(req.AgentLogPath, wantAgent, timeout)
	} else {
		time.Sleep(800 * time.Millisecond)
		invocations, _ = readAgentInvocations(req.AgentLogPath)
	}
	if req.SecondInstance {
		secondCtx, secondCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer secondCancel()
		second := exec.CommandContext(secondCtx, req.Bin, defaultListenArgs(req)...)
		second.Dir = req.WorkDir
		second.Env = env
		var sOut, sErr bytes.Buffer
		second.Stdout = &sOut
		second.Stderr = &sErr
		err := second.Run()
		resp := &Response{
			Stdout:           stdoutBuf.String(),
			Stderr:           stderrBuf.String(),
			AgentInvocations: invocations,
			PostMessages:     posts,
			SecondStdout:     sOut.String(),
			SecondStderr:     sErr.String(),
		}
		if err == nil {
			resp.SecondExitCode = 0
		} else {
			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) {
				resp.SecondExitCode = exitErr.ExitCode()
			}
		}
		_ = cmd.Process.Signal(syscall.SIGTERM)
		<-done
		return resp, nil
	}
	if wantAgent > 0 && agentErr != nil {
		_ = cmd.Process.Kill()
		<-done
		return &Response{Stdout: stdoutBuf.String(), Stderr: stderrBuf.String(), AgentInvocations: invocations, PostMessages: posts}, agentErr
	}
	if wantAgent > 0 {
		_ = waitForPosts(&posts, 1, 3*time.Second)
	}
	_ = cmd.Process.Signal(syscall.SIGTERM)
	<-done
	_ = cmd.Wait()
	return &Response{
		Stdout:           stdoutBuf.String(),
		Stderr:           stderrBuf.String(),
		AgentInvocations: invocations,
		PostMessages:     posts,
	}, nil
}

func assertExitCode(t *testing.T, resp *Response, want int) {
	t.Helper()
	if resp.ExitCode != want {
		t.Fatalf("exit code = %d, want %d\nstdout:\n%s\nstderr:\n%s", resp.ExitCode, want, resp.Stdout, resp.Stderr)
	}
}

func Run(t *testing.T, req *Request) (*Response, error) {
	if req.Bin == "" {
		return nil, fmt.Errorf("req.Bin not set; root Setup must build slack-msg")
	}
	if req.ListenMode || req.Daemon || req.SecondInstance {
		if req.BotToken == "" && !req.ClearSlackEnv {
			req.BotToken = slackTestBotToken
		}
		if req.AppToken == "" && !req.ClearSlackEnv && req.Daemon {
			req.AppToken = slackTestAppToken
		}
		if req.Daemon || req.SecondInstance {
			return runDaemon(t, req)
		}
		return runListenQuick(t, req)
	}
	return runSimple(t, req)
}

func _dsnSlackEventsUnused() {
	_ = slackevents.AppMentionEvent{}
}
```
