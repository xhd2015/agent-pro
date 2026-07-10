# slack-send SDK Refactor Doctests

Doc-style tests for `script/debug/slack-send`: preserve identical CLI behavior
while refactoring the send path to `github.com/slack-go/slack`. Covers help,
config errors, channel resolution (stdout contract), isolated send success via
`slacktest`, send failures, and optional live Slack integration.

# DSN (Domain Specific Notion)

**Participants**

- **Caller** — user or doctest harness invoking the `slack-send` binary (or `go run ./script/debug/slack-send`).
- **slack-send CLI** — parses args (`-h`/`--help`, optional `[channel] [text]`), discovers `slack-config.json` by walking cwd upward until `go.mod`, loads `SlackConfig`, resolves channel names/IDs, prints `Sending to...` and `Using config from...`, then posts via Slack Web API (`chat.postMessage`).
- **slack-config.json** — JSON file with `botToken`, `defaultChannelId`, `knownChannels`, etc.
- **Slack Web API** — real `slack.com` in production; `slacktest` fake server in unit send tests when `SLACK_API_URL` is set (implementer hook, not CLI-visible).
- **Test harness** — builds `slack-send` once per session, runs in isolated temp dirs with fixture configs, captures stdout/stderr/exit code.

**Behaviors**

- No args → `defaultChannelId` + text `"Hello slack"`.
- `[channel]` → `resolveChannel` (known map, `#` prefix, direct `C`/`D`/`G` IDs, unknown passthrough).
- `[channel] [text...]` → joined text.
- `-h` / `--help` → print usage block (hardcoded defaults), exit 0, no config load, no send.
- Config missing/unreadable → stderr `failed to load config ...`, exit 1.
- Empty `botToken` → stderr `botToken is empty in ...`, exit 1.
- Send failure → stderr `send failed: ...` (SDK) or `slack error: ...` (legacy raw path), exit 1, no `OK` line.
- Send success → stdout three lines ending with `OK ts=... channel=...\n`, exit 0.

## Version

0.0.2

## Decision Tree

```
tests/slack-send-sdk/
├── DOCTEST.md
├── SETUP.md                           # build binary, workdir helpers, slacktest session cache
├── testdata/
│   ├── valid-config.json
│   └── empty-token-config.json
├── help/
│   ├── SETUP.md
│   ├── short-flag/                    # -h, unit
│   └── long-flag/                   # --help, unit
├── config-errors/
│   ├── SETUP.md
│   ├── missing-config/              # unit
│   └── empty-bot-token/             # unit
├── channel-resolve/                 # unit — assert Sending line; send fails (fake token)
│   ├── SETUP.md
│   ├── known-name-with-hash/
│   ├── known-name-without-hash/
│   ├── direct-channel-id/
│   ├── direct-dm-id/
│   ├── direct-group-id/
│   └── unknown-passthrough/
├── send-success/                    # unit — slacktest via SLACK_API_URL (RED until SDK + hook)
│   ├── SETUP.md
│   ├── default-no-args/
│   ├── channel-by-name/
│   ├── channel-by-id-custom-text/
│   └── multi-word-text/
├── send-errors/
│   ├── SETUP.md
│   └── invalid-channel-id/          # unit — slacktest returns API error
└── integration/                     # label: integration, slow — real Slack
    ├── SETUP.md
    └── live-default-send/
```

Parameter ranking (most → least significant):

1. **Outcome** — help vs config-error vs resolve-only vs send-success vs send-error vs live integration
2. **Backend** — slacktest (`SLACK_API_URL`) vs real Slack vs no network (help/config)
3. **Channel/text args** — default, by name, by ID, custom text, edge resolution
4. **Help variant** — `-h` vs `--help`

## Test Index

| # | Leaf | Labels | Description |
|---|------|--------|-------------|
| 1 | `help/short-flag` | (default) | `-h` prints full usage; exit 0; no config |
| 2 | `help/long-flag` | (default) | `--help` identical to `-h` |
| 3 | `config-errors/missing-config` | (default) | Isolated dir with go.mod, no config → load failure |
| 4 | `config-errors/empty-bot-token` | (default) | Config with empty token → exit 1 |
| 5 | `channel-resolve/known-name-with-hash` | unit | `#general` → resolved ID in Sending line |
| 6 | `channel-resolve/known-name-without-hash` | unit | `general` → resolved via `#general` map |
| 7 | `channel-resolve/direct-channel-id` | unit | `C...` ID used as-is |
| 8 | `channel-resolve/direct-dm-id` | unit | `D...` ID used as-is |
| 9 | `channel-resolve/direct-group-id` | unit | `G...` ID used as-is |
| 10 | `channel-resolve/unknown-passthrough` | unit | Unknown name passed through unchanged |
| 11 | `send-success/default-no-args` | unit | slacktest OK line with defaults |
| 12 | `send-success/channel-by-name` | unit | `#general` resolves and sends |
| 13 | `send-success/channel-by-id-custom-text` | unit | ID + custom message text |
| 14 | `send-success/multi-word-text` | unit | Multi-word text joined with spaces |
| 15 | `send-errors/invalid-channel-id` | unit | Invalid channel → send failed, no OK |
| 16 | `integration/live-default-send` | integration, slow | Real token from repo `slack-config.json` |

## How to Run

```sh
# Structure validation
doctest vet ./tests/slack-send-sdk

# Default CI — help + config-errors (no labels; no network)
doctest test ./tests/slack-send-sdk

# Unit suite — channel-resolve + slacktest send paths (RED until SDK + SLACK_API_URL)
doctest test --label unit ./tests/slack-send-sdk

# Live Slack (requires repo slack-config.json with valid botToken)
doctest test --label integration ./tests/slack-send-sdk/integration
doctest test --label slow ./tests/slack-send-sdk/integration/live-default-send

# Single leaf
doctest test -v ./tests/slack-send-sdk/help/short-flag
```

**Implementer note (RED until done):** `send-success` and `send-errors/invalid-channel-id` require
`github.com/slack-go/slack` plus a test-only `SLACK_API_URL` env override in `main.go`
(`slack.OptionAPIURL`) so the harness can point at `slacktest`. Channel-resolve leaves pass
against the current raw HTTP impl (assert Sending line before network failure).

```go
import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/slack-go/slack/slacktest"
)

type Request struct {
	RepoRoot      string
	WorkDir       string
	Bin           string
	Args          []string
	ConfigFixture string // basename under DOCTEST_ROOT/testdata/ or leaf testdata/
	ConfigInline  string // raw JSON; wins over ConfigFixture
	WriteGoMod    bool   // write minimal go.mod in WorkDir (default true except help)
	SlackAPIURL   string // SLACK_API_URL env for slacktest-backed sends
	Env           []string
	UseRepoConfig bool // integration: WorkDir=RepoRoot, no fixture write
}

type Response struct {
	ExitCode int
	Stdout   string
	Stderr   string
}

const minimalGoMod = "module slack-send-doctest\n\ngo 1.25\n"

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
	return filepath.Join(os.TempDir(), "slack-send-sdk-doctest-"+DOCTEST_SESSION_ID)
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

func writeWorkDir(t *testing.T, req *Request) error {
	t.Helper()
	if req.UseRepoConfig {
		req.WorkDir = req.RepoRoot
		return nil
	}
	if req.WorkDir == "" {
		req.WorkDir = t.TempDir()
	}
	if req.WriteGoMod {
		if err := os.WriteFile(filepath.Join(req.WorkDir, "go.mod"), []byte(minimalGoMod), 0o644); err != nil {
			return fmt.Errorf("write go.mod: %w", err)
		}
	}
	cfgPath := filepath.Join(req.WorkDir, "slack-config.json")
	if req.ConfigInline != "" {
		if err := os.WriteFile(cfgPath, []byte(req.ConfigInline), 0o644); err != nil {
			return fmt.Errorf("write inline config: %w", err)
		}
		return nil
	}
	if req.ConfigFixture != "" {
		src := resolveFixturePath(req.ConfigFixture)
		data, err := os.ReadFile(src)
		if err != nil {
			return fmt.Errorf("read fixture %s: %w", src, err)
		}
		if err := os.WriteFile(cfgPath, data, 0o644); err != nil {
			return fmt.Errorf("write config: %w", err)
		}
	}
	return nil
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

func withoutEnvKey(env []string, key string) []string {
	prefix := key + "="
	out := make([]string, 0, len(env))
	for _, e := range env {
		if strings.HasPrefix(e, prefix) {
			continue
		}
		out = append(out, e)
	}
	return out
}

var (
	slackTestOnce sync.Once
	slackTestURL  string
	slackTestErr  error
)

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
			sts := slacktest.NewTestServer()
			sts.Start()
			slackTestURL = sts.GetAPIURL()
			return os.WriteFile(urlFile, []byte(slackTestURL), 0o644)
		})
	})
	return slackTestURL, slackTestErr
}

func readFixture(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(resolveFixturePath(name))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return string(data)
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
	if err := writeWorkDir(t, req); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, req.Bin, req.Args...)
	cmd.Dir = req.WorkDir

	env := os.Environ()
	if req.SlackAPIURL != "" {
		env = withoutEnvKey(env, "SLACK_API_URL")
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