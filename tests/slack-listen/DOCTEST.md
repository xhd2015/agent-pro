# slack-listen CLI (Socket Mode → agent-run → reply) Doctests

Doc-style tests for `script/debug/slack-listen listen`: a foreground Slack inbound
bridge using Socket Mode, filtering inbound events, dispatching to `agent-run`, and
replying via `chat.postMessage` in thread. Uses `pkgs/slackutil` for config/tokens/channel
resolution and `slacktest` + `SLACK_API_URL` for unit paths.

# DSN (Domain Specific Notion)

**Participants**

- **Caller** — user or doctest harness invoking `slack-listen listen`.
- **slack-listen CLI** — parses flags (`--bot-token`, `--app-token`, `--config`,
  `--channel`, `--require-mention` / `--no-require-mention`, `--allow-from`,
  `--session-mode`, `--idle-timeout`, `--agent-runner`, `--agent-runner-config-home`,
  `--reply-prefix`, `--lock-file`, `-h`/`--help`); acquires singleton lock; loads
  JSON config only when `--config` (or `SLACK_CONFIG`) is set; connects Socket Mode;
  filters events; dispatches agent; posts thread replies.
- **SlackConfig JSON** — optional file with `botToken`, `appToken`, `knownChannels`;
  never auto-discovered.
- **Socket Mode WebSocket** — `apps.connections.open` then event stream; ack immediately.
- **Slack Web API** — `chat.postMessage` for replies; `slacktest` fake server when
  `SLACK_API_URL` is set.
- **agent-run** — external runner (`run --keep-tty --session <id>` or `send <id>` for
  thread mode; `run` per message for stateless). Mocked via `SLACK_LISTEN_AGENT_RUN` in unit tests.
- **Singleton lock file** — prevents concurrent listeners; second instance exits with
  `another slack-listen is already running`.
- **Test harness** — builds binary once per session, runs quick-exit or daemon probes,
  injects Socket Mode `events_api` envelopes via `slacktest`, captures agent invocations
  and PostMessage calls.

**Behaviors**

- `slack-listen listen [options]` — foreground daemon; SIGINT/SIGTERM graceful stop.
- `-h` / `--help` → usage, exit 0, no Socket Mode connect.
- Missing `botToken` or `appToken` → dedicated stderr, exit 1 before connect.
- No `--config` → startup log `Using config from: (none)`; no cwd walk.
- `--config PATH` → load JSON; startup log `Using config from: <absolute-path>`.
- Ignore bot's own messages; DMs always processed (ignore requireMention); channels
  with requireMention only on `app_mention` or `@bot`; `allowFrom` `*` or user IDs;
  optional `--channel` filter.
- Thread mode (default): first msg → `agent-run run --keep-tty --session <id>`;
  follow-ups → `agent-run send <id>`; session id `slack-{channel}-{thread_ts}`.
- Stateless mode: every msg → `agent-run run`.
- Reply: PostMessage in thread (`thread_ts`); optional `--reply-prefix`.
- Lock: second instance exits non-zero with `another slack-listen is already running`.

## Version

0.0.2

## Decision Tree

```
tests/slack-listen/
├── DOCTEST.md
├── SETUP.md                           # build binary, mock agent, slacktest, daemon helpers
├── testdata/
│   ├── valid-config.json
│   └── empty-app-token-config.json
├── help/
│   ├── SETUP.md
│   ├── short-flag/                    # -h
│   └── long-flag/                     # --help
├── token-errors/
│   ├── SETUP.md
│   ├── missing-bot-token/
│   └── missing-app-token/
├── lock/
│   ├── SETUP.md
│   └── already-running/
├── config/
│   ├── SETUP.md
│   ├── none-stdout-line/
│   ├── explicit-path/
│   └── bad-config-path/
├── filter/                            # label: unit — event gating via injected Socket Mode events
│   ├── SETUP.md
│   ├── ignore-bot-self/
│   ├── dm-always-processed/
│   ├── channel-requires-mention/
│   ├── channel-no-require-mention/
│   ├── allow-from-blocked/
│   ├── allow-from-wildcard/
│   └── channel-filter-excludes/
├── session-routing/                   # label: unit — mock agent-run argv recording
│   ├── SETUP.md
│   ├── thread-first-run/
│   ├── thread-follow-up-send/
│   └── stateless-each-run/
├── reply/                             # label: unit — PostMessage thread_ts + prefix
│   ├── SETUP.md
│   ├── posts-in-thread/
│   └── reply-prefix/
└── integration/                       # label: integration, slow
    ├── SETUP.md
    └── live-socket-reply/
```

Parameter ranking (most → least significant):

1. **Outcome** — help vs token-error vs lock vs config vs filter vs session-routing vs reply vs integration
2. **Session mode** — thread (default) vs stateless (session-routing group)
3. **Filter factor** — mention gate, DM bypass, allowFrom, bot-self, channel filter
4. **Backend** — slacktest (`SLACK_API_URL`) vs live Socket Mode

## Test Index

| # | Leaf | Labels | Description |
|---|------|--------|-------------|
| 1 | `help/short-flag` | (default) | `-h` prints usage; exit 0; no connect |
| 2 | `help/long-flag` | (default) | `--help` identical contract to `-h` |
| 3 | `token-errors/missing-bot-token` | (default) | No bot token → `bot token required` |
| 4 | `token-errors/missing-app-token` | (default) | Bot only, no app token → `app token required` |
| 5 | `lock/already-running` | unit | Second instance exits with singleton message |
| 6 | `config/none-stdout-line` | unit | Startup logs `Using config from: (none)` |
| 7 | `config/explicit-path` | unit | `--config` logs absolute path |
| 8 | `config/bad-config-path` | (default) | Missing config file → `failed to load config` |
| 9 | `filter/ignore-bot-self` | unit | Bot-authored message ignored |
| 10 | `filter/dm-always-processed` | unit | DM processed without mention |
| 11 | `filter/channel-requires-mention` | unit | Channel message without mention ignored |
| 12 | `filter/channel-no-require-mention` | unit | `--no-require-mention` processes channel message |
| 13 | `filter/allow-from-blocked` | unit | User not in allowFrom ignored |
| 14 | `filter/allow-from-wildcard` | unit | `*` allowFrom processes any user |
| 15 | `filter/channel-filter-excludes` | unit | `--channel` filter drops other channels |
| 16 | `session-routing/thread-first-run` | unit | First thread msg → `run --keep-tty --session slack-...` |
| 17 | `session-routing/thread-follow-up-send` | unit | Second thread msg → `send <session-id>` |
| 18 | `session-routing/stateless-each-run` | unit | Every msg → `run` (no send) |
| 19 | `reply/posts-in-thread` | unit | PostMessage includes `thread_ts` |
| 20 | `reply/reply-prefix` | unit | Reply text prefixed |
| 21 | `integration/live-socket-reply` | integration, slow | Live Socket Mode + agent reply |

## How to Run

```sh
# Structure validation
doctest vet ./tests/slack-listen

# Default CI — help + validation + config error (no labels; no network/daemon)
doctest test ./tests/slack-listen

# Unit suite — slacktest + mock agent-run (RED until P1 lands)
doctest test --label unit ./tests/slack-listen

# Live Slack Socket Mode (requires repo slack-config.json with botToken + appToken)
doctest test --label integration ./tests/slack-listen/integration
doctest test --label slow ./tests/slack-listen/integration/live-socket-reply

# Single leaf
doctest test -v ./tests/slack-listen/filter/dm-always-processed
```

**Implementer note (RED until done):** Requires `script/debug/slack-listen` with `listen`
subcommand, `pkgs/slackutil`, `SLACK_API_URL` → `slack.OptionAPIURL`, env
`SLACK_LISTEN_AGENT_RUN` for mock agent path, startup log line for config source,
and singleton lock file behavior.

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
	slackTestBotToken      = "xoxb-slacktest-token"
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

	BotToken string
	AppToken string
	LockFile string

	Daemon         bool
	InjectEvents   []InjectedEvent
	WantAgentCalls int // -1 => len(InjectEvents)
	AgentLogPath   string
	MockAgentPath  string
	ObserveTimeout time.Duration
	Posts          *[]CapturedPost

	SecondInstance bool
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

var (
	buildOnce sync.Once
	builtBin  string
	buildErr  error
)

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

func sessionCacheDir() string {
	return filepath.Join(os.TempDir(), "slack-listen-doctest-"+DOCTEST_SESSION_ID)
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

func buildSlackListen(t *testing.T) (string, error) {
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
		bin := filepath.Join(cacheDir, "slack-listen")
		ready := filepath.Join(cacheDir, "bin.ready")
		lock := filepath.Join(cacheDir, "build.lock")
		buildErr = withFileLock(lock, func() error {
			if fileExists(ready) && fileExists(bin) {
				builtBin = bin
				return nil
			}
			ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, "go", "build", "-o", bin, "./script/debug/slack-listen")
			cmd.Dir = repoRoot
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("go build ./script/debug/slack-listen: %w\n%s", err, stderr.String())
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

func newSlackTestServer(t *testing.T, posts *[]CapturedPost) (*slacktest.Server, string) {
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

func defaultListenArgs(req *Request) []string {
	args := []string{"listen"}
	if req.BotToken != "" {
		args = append(args, "--bot-token", req.BotToken)
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

func runQuick(t *testing.T, req *Request) (*Response, error) {
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
	sts, apiURL := newSlackTestServer(t, &posts)
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
		return nil, fmt.Errorf("req.Bin not set; root Setup must build slack-listen")
	}
	if req.BotToken == "" && !req.ClearSlackEnv {
		req.BotToken = slackTestBotToken
	}
	if req.AppToken == "" && !req.ClearSlackEnv && req.Daemon {
		req.AppToken = slackTestAppToken
	}
	if req.Daemon || req.SecondInstance {
		return runDaemon(t, req)
	}
	return runQuick(t, req)
}

func _dsnSlackEventsUnused() {
	_ = slackevents.AppMentionEvent{}
}
```