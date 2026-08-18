# Scenario

**Feature**: `agent-run run --open` opens a keep-alive TTY session, auto-attaches
silently, then prints the terminal session id after detach

```
# validation
agent-run run --open --agent-runner fake-codex "x" -> error (non-TTY)
agent-run run --open --json --agent-runner grok-tty "x" -> error
agent-run run --agent-runner grok-tty -> prompt is required
agent-run run --agent-runner grok-tty --open -> empty prompt allowed

# open lifecycle (TTY)
agent-run run --agent-runner grok-tty --open ["prompt"]
  -> silent start + auto-attach
  -> on attach exit: stderr once "grok-tty: <id>"
  -> registry/PTY kept alive
```

## Preconditions

- Repository contains `cmd/agent-run` and `cmd/fake-codex`.
- Each leaf uses isolated `AGENT_RUN_HOME` under `t.TempDir()`.
- Non-TTY reject leaves use `fake-codex` on `PATH`.
- TTY lifecycle leaves set `AGENT_RUN_GROK_TTY_COMMAND` (or codex hook) to a fake
  interactive TUI that prints `GROK_TTY_BANNER` then reads/holds.
- Lifecycle leaves set `AGENT_RUN_OPEN_ATTACH_INSTANT=1` so auto-attach returns
  without a real controlling TTY (test contract; implementer must honor).
- **attach-without-banner** leaves deliberately **do not** set
  `AGENT_RUN_OPEN_ATTACH_INSTANT` — they exercise the production open readiness
  path (attach-first; no hard banner fail). Fake TUI prints no banner/OpenReady
  markers, holds long enough for attach, then exits so `AttachWriter` can return
  without an interactive controlling TTY.
- Session-scoped build cache may share compiled binaries across parallel leaves.
- Discovery hang / “Resolve session id…” **fixing** is out of scope; open mode
  must simply not print that progress to the screen.

## Steps

1. Root `Setup` resolves repo root, creates temp home, builds `agent-run` and
   `fake-codex` (session cache), sets `AGENT_RUN_HOME` + `PATH`.
2. Grouping `Setup` narrows outcome class (`help` / `reject` / `prompt-policy` /
   `tty-lifecycle` / `attach-without-banner`) and runner.
3. Leaf `Setup` finalizes flags, prompt, TTY hooks, and open-attach instant env
   (INSTANT only for tty-lifecycle; attach-without-banner forces INSTANT off).
4. `Run` executes `agent-run` with `req.Args` (optional registry post-read).
5. Leaf `Assert` checks exit code, error text, silence, session id line, registry,
   banner-policy, or argv/stdin inject probe (no-double-inject).

## Context

- Nested DOCTEST root: does not inherit parent `cmd/agent-run/tests` Setup/Run.
- Repo root from this tree: `d.DOCTEST_ROOT/../../../../..`.
- Session cache dir: `$TMPDIR/agent-run-run-open-doctest-<d.DOCTEST_SESSION_ID>/`.
- Test hook env: `AGENT_RUN_OPEN_ATTACH_INSTANT=1` — auto-attach returns
  immediately so lifecycle leaves complete without interactive stdin.

```go
import (
	"runtime"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

const (
	grokTTYBannerMarker     = "GROK_TTY_BANNER"
	envOpenAttachInstant    = "AGENT_RUN_OPEN_ATTACH_INSTANT"
	envGrokTTYCommand       = "AGENT_RUN_GROK_TTY_COMMAND"
	envCodexTTYCommand      = "AGENT_RUN_CODEX_TTY_COMMAND"
)

func sessionCacheDir(d *session.Doctest) string {
	return filepath.Join(os.TempDir(), "agent-run-run-open-doctest-"+d.DOCTEST_SESSION_ID)
}

func withFileLock(t *testing.T, lockPath string, fn func() error) error {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(lockPath), 0755); err != nil {
		return err
	}
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0644)
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

func buildOnce(t *testing.T, d *session.Doctest) (agentRun, fakeCodex string, err error) {
	t.Helper()
	cache := sessionCacheDir(d)
	agentRun = filepath.Join(cache, "agent-run")
	fakeCodex = filepath.Join(cache, "fake-codex")
	lock := filepath.Join(cache, "build.lock")
	ready := filepath.Join(cache, "binaries.ready")
	repoRoot, err := findAgentProRoot(d.DOCTEST_ROOT)
	if err != nil {
		return "", "", err
	}
	err = withFileLock(t, lock, func() error {
		if fileExists(ready) && fileExists(agentRun) && fileExists(fakeCodex) {
			return nil
		}
		if err := os.MkdirAll(cache, 0755); err != nil {
			return err
		}
		build := exec.Command(runtime.GOROOT()+"/bin/go", "build", "-C", "cmd", "-o", agentRun, "./agent-run")
		build.Dir = repoRoot
		if out, err := build.CombinedOutput(); err != nil {
			return fmt.Errorf("build agent-run: %w\n%s", err, string(out))
		}
		build2 := exec.Command(runtime.GOROOT()+"/bin/go", "build", "-C", "cmd", "-o", fakeCodex, "./fake-codex")
		build2.Dir = repoRoot
		if out, err := build2.CombinedOutput(); err != nil {
			return fmt.Errorf("build fake-codex: %w\n%s", err, string(out))
		}
		return os.WriteFile(ready, []byte("ok"), 0644)
	})
	return agentRun, fakeCodex, err
}

func fakeTUIRespondHi() string {
	return `sh -c 'printf "GROK_TTY_BANNER\nGrok › "; read line; echo "Response: $line"'`
}

func fakeTUIHoldSeconds(sec int) string {
	return fmt.Sprintf(`sh -c 'printf "GROK_TTY_BANNER\nGrok › "; sleep %d'`, sec)
}

// fakeTUIDelayedBanner delays legacy banner so non-open inject wait can be proven.
func fakeTUIDelayedBanner() string {
	return `sh -c 'sleep 0.3; printf "GROK_TTY_BANNER\nGrok › "; read line; echo "Response: $line"'`
}

// writeFakeTUIDelayedBannerProbe delays banner, then records the first stdin line
// (injected prompt) into probePath — proves non-open hard-wait inject without
// relying on grok session discovery / capture streaming.
// Prefer writeFakeTUIBannerArgvStdinProbe for new-session no-double-inject leaves
// (argv + timed read so absence of inject can complete).
func writeFakeTUIDelayedBannerProbe(t *testing.T, dir, probePath string) string {
	t.Helper()
	path := filepath.Join(dir, "fake-tui-delayed-banner-probe.sh")
	script := fmt.Sprintf(`#!/bin/sh
sleep 0.3
printf "GROK_TTY_BANNER\nGrok › "
if read -r line; then
  echo "STDIN=$line" > %q
  echo "Response: $line"
else
  echo "STDIN_COUNT=0" > %q
fi
`, probePath, probePath)
	if err := os.WriteFile(path, []byte(script), 0755); err != nil {
		t.Fatalf("write delayed-banner probe TUI: %v", err)
	}
	return path
}

// writeFakeTUIBannerArgvStdinProbe paints GROK_TTY_BANNER (optional delay), records
// trailing argv / PROMPT_ARG, then timed-reads PTY stdin for inject proof.
// Non-open new-session leaves use this so:
//   - banner hard-wait can succeed
//   - argv prompt is visible (PROMPT_ARG)
//   - re-inject absence is visible (STDIN_COUNT=0) without blocking forever
func writeFakeTUIBannerArgvStdinProbe(t *testing.T, dir, probePath string, delaySec float64, readTimeoutSec, holdAfterSec int) string {
	t.Helper()
	if readTimeoutSec <= 0 {
		readTimeoutSec = 5
	}
	if holdAfterSec < 0 {
		holdAfterSec = 0
	}
	if delaySec < 0 {
		delaySec = 0
	}
	path := filepath.Join(dir, "fake-tui-banner-argv-stdin-probe.sh")
	// delay via sleep; shell accepts fractional seconds on bash/sh used in CI.
	script := fmt.Sprintf(`#!/bin/sh
{
  echo "ARGV=$*"
  echo "PROMPT_ARG=${1-}"
} > %q
sleep %g
printf "GROK_TTY_BANNER\nGrok › "
if read -t %d line; then
  echo "STDIN=$line" >> %q
  echo "STDIN_COUNT=1" >> %q
  echo "Response: $line"
else
  echo "STDIN_COUNT=0" >> %q
fi
sleep %d
`, probePath, delaySec, readTimeoutSec, probePath, probePath, probePath, holdAfterSec)
	if err := os.WriteFile(path, []byte(script), 0755); err != nil {
		t.Fatalf("write banner argv/stdin probe TUI: %v", err)
	}
	return path
}

// writeFakeTUINoBannerHold writes a script that never prints legacy banner or
// modern OpenReady chrome (no GROK_TTY_BANNER / Grok › / ❯+Grok). Holds then
// exits so AttachWriter can complete via terminal-exit without INSTANT.
func writeFakeTUINoBannerHold(t *testing.T, dir string, holdSec int) string {
	t.Helper()
	if holdSec <= 0 {
		holdSec = 12
	}
	path := filepath.Join(dir, "fake-tui-no-banner-hold.sh")
	script := fmt.Sprintf("#!/bin/sh\nprintf 'booting\\n'\nsleep %d\n", holdSec)
	if err := os.WriteFile(path, []byte(script), 0755); err != nil {
		t.Fatalf("write no-banner hold TUI: %v", err)
	}
	return path
}

// writeFakeTUINoBannerArgvStdinProbe records trailing argv (new-session prompt)
// and any PTY stdin inject. Prints only non-ready "booting" text.
func writeFakeTUINoBannerArgvStdinProbe(t *testing.T, dir, probePath string, readTimeoutSec, holdAfterSec int) string {
	t.Helper()
	if readTimeoutSec <= 0 {
		readTimeoutSec = 5
	}
	if holdAfterSec <= 0 {
		holdAfterSec = 2
	}
	path := filepath.Join(dir, "fake-tui-no-banner-probe.sh")
	script := fmt.Sprintf(`#!/bin/sh
{
  echo "ARGV=$*"
  echo "PROMPT_ARG=${1-}"
} > %q
printf 'booting\n'
if read -t %d line; then
  echo "STDIN=$line" >> %q
  echo "STDIN_COUNT=1" >> %q
else
  echo "STDIN_COUNT=0" >> %q
fi
sleep %d
`, probePath, readTimeoutSec, probePath, probePath, probePath, holdAfterSec)
	if err := os.WriteFile(path, []byte(script), 0755); err != nil {
		t.Fatalf("write no-banner probe TUI: %v", err)
	}
	return path
}

func clearOpenInstantAttach(req *Request) {
	req.OpenInstantAttach = false
	// Force off even if the parent process exported INSTANT=1 (req.Env overrides
	// os.Environ duplicates by being appended last in execCmd).
	setEnvKV(req, envOpenAttachInstant, "0")
}

func hasBannerNotDetected(combined string) bool {
	lower := strings.ToLower(combined)
	return strings.Contains(lower, "banner not detected") ||
		strings.Contains(lower, "tui banner not detected")
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

func setEnvKV(req *Request, key, value string) {
	req.Env = withoutEnvKey(req.Env, key)
	req.Env = append(req.Env, key+"="+value)
}

func setGrokTTYCommand(req *Request, cmd string) {
	req.GrokTTYCommand = cmd
	setEnvKV(req, envGrokTTYCommand, cmd)
}

func setCodexTTYCommand(req *Request, cmd string) {
	req.CodexTTYCommand = cmd
	setEnvKV(req, envCodexTTYCommand, cmd)
}

func applyOpenInstantAttach(req *Request) {
	if !req.OpenInstantAttach {
		return
	}
	setEnvKV(req, envOpenAttachInstant, "1")
}

func execCmd(t *testing.T, command string, args []string, dir string, env []string, timeout time.Duration) (*Response, error) {
	t.Helper()
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	resp := &Response{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
		Err:    err,
	}
	if err == nil {
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

func runAgentRun(t *testing.T, req *Request, args ...string) (*Response, error) {
	t.Helper()
	if len(args) == 0 {
		args = req.Args
	}
	applyOpenInstantAttach(req)
	resp, err := execCmd(t, req.AgentRun, args, req.TempDir, req.Env, req.ExecTimeout)
	if err != nil {
		return resp, err
	}
	if resp != nil {
		if id, ok := parsePrefixedSessionID(resp.Stderr, "grok-tty"); ok {
			resp.SessionID = id
		} else if id, ok := parsePrefixedSessionID(resp.Stderr, "codex-tty"); ok {
			resp.SessionID = id
		}
		if req.Mode == "open-registry-after" && resp.SessionID != "" && resp.ExitCode == 0 {
			runner := req.Runner
			if runner == "" {
				runner = "grok-tty"
			}
			entry, rerr := readRegistryEntryOptional(req.Home, runner, resp.SessionID)
			if rerr == nil {
				resp.RegistryEntry = entry
			}
		}
	}
	return resp, nil
}

func assertSuccess(t *testing.T, resp *Response) {
	t.Helper()
	if resp.ExitCode != 0 {
		t.Fatalf("exit code = %d, stderr:\n%s\nstdout:\n%s", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
}

func assertContains(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("missing %q in:\n%s", want, got)
	}
}

// parsePrefixedSessionID extracts "<runner>: <id>" where id is a single token.
// Skips multi-word diagnostic lines (e.g. "grok-tty: grok session …").
func parsePrefixedSessionID(stderr, runner string) (string, bool) {
	prefix := runner + ":"
	for _, line := range strings.Split(stderr, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, prefix) {
			continue
		}
		rest := strings.TrimSpace(strings.TrimPrefix(line, prefix))
		if rest == "" || strings.Contains(rest, " ") {
			continue
		}
		if matched, _ := regexp.MatchString(`^[a-zA-Z0-9][a-zA-Z0-9._-]*$`, rest); matched {
			return rest, true
		}
	}
	return "", false
}

func countPrefixedSessionIDLines(stderr, runner string) int {
	prefix := runner + ":"
	n := 0
	for _, line := range strings.Split(stderr, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, prefix) {
			continue
		}
		rest := strings.TrimSpace(strings.TrimPrefix(line, prefix))
		if rest == "" || strings.Contains(rest, " ") {
			continue
		}
		if matched, _ := regexp.MatchString(`^[a-zA-Z0-9][a-zA-Z0-9._-]*$`, rest); matched {
			n++
		}
	}
	return n
}

func registryDir(home, runner string) string {
	return filepath.Join(home, runner+"-registry")
}

func registryPath(home, runner, sessionID string) string {
	return filepath.Join(registryDir(home, runner), sessionID+".json")
}

func readRegistryEntryOptional(home, runner, sessionID string) (*RegistryEntry, error) {
	path := registryPath(home, runner, sessionID)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var entry RegistryEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, err
	}
	if entry.SessionID == "" {
		entry.SessionID = sessionID
	}
	return &entry, nil
}

func portOpen(addr string) bool {
	conn, err := net.DialTimeout("tcp", addr, 500*time.Millisecond)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

// forbiddenOpenNoise reports human-facing discovery/event stream noise that
// must not appear on stdout/stderr during --open.
func forbiddenOpenNoise(combined string) []string {
	checks := []string{
		"Resolve session id",
		"💭",
		"💬",
		"[done]",
		`"type":"think"`,
		`"type":"message"`,
		`"type":"done"`,
	}
	var found []string
	for _, c := range checks {
		if strings.Contains(combined, c) {
			found = append(found, c)
		}
	}
	return found
}

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	repoRoot, err := findAgentProRoot(d.DOCTEST_ROOT)
	if err != nil {
		return err
	}
	req.RepoRoot = repoRoot
	req.TempDir = t.TempDir()
	req.Home = filepath.Join(req.TempDir, ".agent-run")
	if err := os.MkdirAll(req.Home, 0755); err != nil {
		return fmt.Errorf("mkdir home: %w", err)
	}
	agentRun, fakeCodex, err := buildOnce(t, d)
	if err != nil {
		return err
	}
	binDir := filepath.Join(req.TempDir, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("mkdir bin: %w", err)
	}
	req.AgentRun = filepath.Join(binDir, "agent-run")
	req.FakeCodex = filepath.Join(binDir, "fake-codex")
	if out, err := exec.Command("cp", agentRun, req.AgentRun).CombinedOutput(); err != nil {
		return fmt.Errorf("cp agent-run: %w\n%s", err, string(out))
	}
	if out, err := exec.Command("cp", fakeCodex, req.FakeCodex).CombinedOutput(); err != nil {
		return fmt.Errorf("cp fake-codex: %w\n%s", err, string(out))
	}
	if err := os.Chmod(req.AgentRun, 0755); err != nil {
		return err
	}
	if err := os.Chmod(req.FakeCodex, 0755); err != nil {
		return err
	}
	req.Env = append(req.Env,
		"AGENT_RUN_HOME="+req.Home,
		"PATH="+binDir+":"+os.Getenv("PATH"),
	)
	req.Args = []string{"run"}
	return nil
}

func findAgentProRoot(start string) (string, error) {
	if start == "" {
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		start = wd
	}
	for dir := start; ; dir = filepath.Dir(dir) {
		data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
		if err == nil {
			for _, line := range strings.Split(string(data), "\n") {
				if strings.TrimSpace(line) == "module github.com/xhd2015/agent-pro" {
					return dir, nil
				}
			}
		}
		if filepath.Dir(dir) == dir {
			return "", fmt.Errorf("could not find agent-pro module root above %s", start)
		}
	}
}

```
