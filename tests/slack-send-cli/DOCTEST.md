# slack-send CLI (less-flags) Doctests

Doc-style tests for `script/debug/slack-send` after refactor to `less-flags`:
required positional MESSAGE, explicit `--config` only (no auto-discovery), channel
name resolution via Slack API (`conversations.list`), and `slacktest` for unit sends.

# DSN (Domain Specific Notion)

**Participants**

- **Caller** — user or doctest harness invoking the `slack-send` binary.
- **slack-send CLI** — parses flags (`--token`, `--channel`, `--config`, `-h`/`--help`)
  with `less-flags` + `StopOnFirstArg`, requires exactly one positional MESSAGE,
  resolves token/channel from CLI → env → config precedence, loads JSON config only
  when `--config` (or `SLACK_CONFIG`) is set, resolves channel names to IDs via
  `knownChannels` map or `conversations.list`, prints three stdout lines on success,
  posts via `chat.postMessage`.
- **SlackConfig JSON** — optional file with `botToken`, `defaultChannelId`,
  `knownChannels`; never auto-discovered.
- **Slack Web API** — real `slack.com` in integration; `slacktest` fake server in
  unit tests when `SLACK_API_URL` is set (including custom `conversations.list`).
- **Test harness** — builds `slack-send` once per session, runs in isolated temp
  dirs, captures stdout/stderr/exit code.

**Behaviors**

- `slack-send [options] MESSAGE` — message required; reject 0 or 2+ positionals.
- No `--config` → stdout `Using config from: (none)`; no cwd walk for config.
- `--config PATH` → load JSON; stdout `Using config from: <absolute-path>`.
- Channel ID (`C`/`D`/`G` prefix) used as-is; names normalized and resolved.
- `-h` / `--help` → usage, exit 0, no send.
- Missing token/channel/message → dedicated stderr messages, exit 1.
- Send failure → stderr `send failed:` prefix, exit 1, no `OK` line.

## Version

0.0.2

## Decision Tree

```
tests/slack-send-cli/
├── DOCTEST.md
├── SETUP.md                           # build binary, env helpers, slacktest session cache
├── testdata/
│   ├── valid-config.json
│   ├── empty-token-config.json
│   └── default-channel-name.json
├── help/
│   ├── SETUP.md
│   ├── short-flag/                    # -h
│   └── long-flag/                     # --help
├── message-errors/
│   ├── SETUP.md
│   ├── missing-message/
│   └── multiple-positionals/
├── token-errors/
│   ├── SETUP.md
│   └── missing-token/
├── channel-errors/
│   ├── SETUP.md
│   └── missing-channel/
├── config-errors/
│   ├── SETUP.md
│   ├── bad-config-path/
│   └── empty-bot-token/
├── config-none/
│   ├── SETUP.md
│   └── stdout-line/
├── channel-resolve/                   # label: unit — name→ID via API or knownChannels
│   ├── SETUP.md
│   ├── api-name-with-hash/
│   ├── api-name-without-hash/
│   ├── direct-channel-id/
│   ├── direct-dm-id/
│   ├── direct-group-id/
│   └── config-known-channels/
├── send-success/                      # label: unit — slacktest isolated send
│   ├── SETUP.md
│   ├── cli-flags/
│   ├── channel-by-id/
│   ├── multi-word-message/
│   ├── env-token/
│   └── env-channel/
├── send-errors/                       # label: unit
│   ├── SETUP.md
│   ├── channel-not-found/
│   └── api-post-failed/
├── config-with-default/
│   ├── SETUP.md
│   ├── message-only/
│   ├── override-channel/
│   └── default-channel-name/
└── integration/                       # label: integration, slow
    ├── SETUP.md
    └── live-explicit-config/
```

Parameter ranking (most → least significant):

1. **Outcome** — help vs validation-error vs config-none vs resolve vs send-success vs send-error vs live
2. **Credential source** — CLI flags vs env vs `--config` JSON
3. **Channel resolution** — direct ID vs API list vs knownChannels map
4. **Backend** — slacktest (`SLACK_API_URL`) vs real Slack vs no network

## Test Index

| # | Leaf | Labels | Description |
|---|------|--------|-------------|
| 1 | `help/short-flag` | (default) | `-h` prints usage; exit 0; no send |
| 2 | `help/long-flag` | (default) | `--help` identical contract to `-h` |
| 3 | `message-errors/missing-message` | (default) | Flags only, no MESSAGE → `message required` |
| 4 | `message-errors/multiple-positionals` | (default) | Two positionals → `exactly one message required` |
| 5 | `token-errors/missing-token` | (default) | `--channel` + message, no token → `bot token required` |
| 6 | `channel-errors/missing-channel` | (default) | `--token` + message, no channel → `channel required` |
| 7 | `config-errors/bad-config-path` | (default) | `--config` missing file → `failed to load config` |
| 8 | `config-errors/empty-bot-token` | (default) | Config with empty token, no override → exit 1 |
| 9 | `config-none/stdout-line` | (default) | CLI flags only → `Using config from: (none)` |
| 10 | `channel-resolve/api-name-with-hash` | unit | `#general` resolved via conversations.list |
| 11 | `channel-resolve/api-name-without-hash` | unit | `general` normalized and resolved via API |
| 12 | `channel-resolve/direct-channel-id` | unit | `C...` ID passed through |
| 13 | `channel-resolve/direct-dm-id` | unit | `D...` ID passed through |
| 14 | `channel-resolve/direct-group-id` | unit | `G...` ID passed through |
| 15 | `channel-resolve/config-known-channels` | unit | `--config` knownChannels fast path |
| 16 | `send-success/cli-flags` | unit | `--token --channel` + message; config none |
| 17 | `send-success/channel-by-id` | unit | Channel ID + custom message |
| 18 | `send-success/multi-word-message` | unit | Single quoted multi-word MESSAGE |
| 19 | `send-success/env-token` | unit | `SLACK_BOT_TOKEN` env fallback |
| 20 | `send-success/env-channel` | unit | `SLACK_CHANNEL` env fallback |
| 21 | `send-errors/channel-not-found` | unit | Unknown name → `channel not found` |
| 22 | `send-errors/api-post-failed` | unit | PostMessage API error → `send failed:` |
| 23 | `config-with-default/message-only` | unit | `--config` + message uses defaults |
| 24 | `config-with-default/override-channel` | unit | `--config --channel` CLI wins |
| 25 | `config-with-default/default-channel-name` | unit | Config `defaultChannelId` name resolved |
| 26 | `integration/live-explicit-config` | integration, slow | Real Slack via `--config` + message |

## How to Run

```sh
# Structure validation
doctest vet ./tests/slack-send-cli

# Default CI — help + validation + config-none (no labels; no network)
doctest test ./tests/slack-send-cli

# Unit suite — slacktest send/resolve paths (RED until refactor lands)
doctest test --label unit ./tests/slack-send-cli

# Live Slack (requires repo slack-config.json with valid botToken)
doctest test --label integration ./tests/slack-send-cli/integration
doctest test --label slow ./tests/slack-send-cli/integration/live-explicit-config

# Single leaf
doctest test -v ./tests/slack-send-cli/help/short-flag
```

**Implementer note (RED until done):** Unit leaves require `less-flags` CLI, explicit
`--config` only, required MESSAGE, `conversations.list` channel resolution, and
`SLACK_API_URL` → `slack.OptionAPIURL` for slacktest.

```go
import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slacktest"
)

type Request struct {
	RepoRoot      string
	WorkDir       string
	Bin           string
	Args          []string
	ConfigPath    string // absolute path written by Setup when using --config
	ConfigFixture string // basename under DOCTEST_ROOT/testdata/
	ConfigInline  string // raw JSON; wins over ConfigFixture
	SlackAPIURL   string // SLACK_API_URL env for slacktest-backed runs
	Env           []string
	ClearSlackEnv bool // strip SLACK_BOT_TOKEN, SLACK_CHANNEL, SLACK_CONFIG
}

type Response struct {
	ExitCode int
	Stdout   string
	Stderr   string
}

const slackTestToken = "xoxb-slacktest-token"

var slackTestChannels = []slack.Channel{
	{
		GroupConversation: slack.GroupConversation{
			Conversation: slack.Conversation{
				ID: "C0ALE44K5J6",
			},
			Name: "general",
		},
	},
	{
		GroupConversation: slack.GroupConversation{
			Conversation: slack.Conversation{
				ID: "C0OTHERCHAN",
			},
			Name: "random",
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

func buildSlackSend(t *testing.T) (string, error) {
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
		bin := filepath.Join(cacheDir, "slack-send")
		ready := filepath.Join(cacheDir, "bin.ready")
		lock := filepath.Join(cacheDir, "build.lock")
		buildErr = withFileLock(lock, func() error {
			if fileExists(ready) && fileExists(bin) {
				builtBin = bin
				return nil
			}
			ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, "go", "build", "-o", bin, "./script/debug/slack-send")
			cmd.Dir = repoRoot
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("go build ./script/debug/slack-send: %w\n%s", err, stderr.String())
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

func sessionCacheDir() string {
	return filepath.Join(os.TempDir(), "slack-send-cli-doctest-"+DOCTEST_SESSION_ID)
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

func readFixture(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(resolveFixturePath(name))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return string(data)
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
)

var (
	slackTestOnce   sync.Once
	slackTestURL    string
	slackTestErr    error
	slackTestPostFailOnce sync.Once
	slackTestPostFailURL  string
	slackTestPostFailErr  error
)

func conversationsListHandler(channels []slack.Channel) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := struct {
			slack.SlackResponse
			Channels []slack.Channel `json:"channels"`
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

func startSlackTestServer(mode slackTestMode) (string, error) {
	sts := slacktest.NewTestServer(func(s slacktest.Customize) {
		s.Handle("/conversations.list", conversationsListHandler(slackTestChannels))
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
		cacheDir := sessionCacheDir()
		urlFile := filepath.Join(cacheDir, "slacktest.url")
		lock := filepath.Join(cacheDir, "slacktest.lock")
		slackTestErr = withFileLock(lock, func() error {
			if b, err := os.ReadFile(urlFile); err == nil && len(strings.TrimSpace(string(b))) > 0 {
				slackTestURL = strings.TrimSpace(string(b))
				return nil
			}
			url, err := startSlackTestServer(slackTestDefault)
			if err != nil {
				return err
			}
			slackTestURL = url
			return os.WriteFile(urlFile, []byte(slackTestURL), 0o644)
		})
	})
	return slackTestURL, slackTestErr
}

func ensureSlackTestServerPostFail(t *testing.T) (string, error) {
	t.Helper()
	slackTestPostFailOnce.Do(func() {
		cacheDir := sessionCacheDir()
		urlFile := filepath.Join(cacheDir, "slacktest-postfail.url")
		lock := filepath.Join(cacheDir, "slacktest-postfail.lock")
		slackTestPostFailErr = withFileLock(lock, func() error {
			if b, err := os.ReadFile(urlFile); err == nil && len(strings.TrimSpace(string(b))) > 0 {
				slackTestPostFailURL = strings.TrimSpace(string(b))
				return nil
			}
			url, err := startSlackTestServer(slackTestPostFail)
			if err != nil {
				return err
			}
			slackTestPostFailURL = url
			return os.WriteFile(urlFile, []byte(slackTestPostFailURL), 0o644)
		})
	})
	return slackTestPostFailURL, slackTestPostFailErr
}

func assertExitCode(t *testing.T, resp *Response, want int) {
	t.Helper()
	if resp.ExitCode != want {
		t.Fatalf("exit code = %d, want %d\nstdout:\n%s\nstderr:\n%s", resp.ExitCode, want, resp.Stdout, resp.Stderr)
	}
}

func Run(t *testing.T, req *Request) (*Response, error) {
	if req.Bin == "" {
		return nil, fmt.Errorf("req.Bin not set; root Setup must build slack-send")
	}
	if req.WorkDir == "" {
		req.WorkDir = t.TempDir()
	}
	if err := materializeConfig(t, req); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, req.Bin, req.Args...)
	cmd.Dir = req.WorkDir

	env := os.Environ()
	if req.ClearSlackEnv {
		env = withoutEnvKeys(env, "SLACK_BOT_TOKEN", "SLACK_CHANNEL", "SLACK_CONFIG")
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
```