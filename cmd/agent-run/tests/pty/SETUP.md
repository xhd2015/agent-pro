# Scenario

**Feature**: `agent-run pty` — stats, kill-orphans (PPID / kind / all filters), and open-path serve reclaim

```
# stats reads process table + optional sysctl/ptmx probes
agent-run pty stats -> human-readable PTY limit / masters / __serve counts

# kill-orphans: default PPID==1; --kind OR predicates; --all wins; --exe ANDed
agent-run pty kill-orphans [--dry-run] [--exe PATH] [--all]
                           [--kind=test-generated] [--kind=workdir-at-tmp]
  -> list or terminate selected serve PIDs (never self)

# open keep-alive leaves detached __serve; harness reclaim removes it
agent-run run --open (+ instant attach) -> keep-alive serve
  -> reclaim (ReclaimSessionID / kill serve PID) -> no leftover serve
```

## Preconditions

- Nested DOCTEST root: does not inherit parent `cmd/agent-run/tests` Setup/Run.
- Repo root from this tree: `d.DOCTEST_ROOT/../../../..`
  (`pty` → `tests` → `agent-run` → `cmd` → module root).
- Each leaf uses isolated `AGENT_RUN_HOME` under `t.TempDir()`.
- Session-scoped build cache: `$TMPDIR/agent-run-pty-doctest-<d.DOCTEST_SESSION_ID>/`
  shares the compiled `agent-run` binary across parallel leaves.
- Kill tests **must** pass `--exe` pointing at the test binary so host
  brainstorm/seatalk/`__serve*` processes are not destroyed (when a single
  binary path covers all spawned serves).
- Open-cleanup leaves set `AGENT_RUN_OPEN_ATTACH_INSTANT=1` and a fake TUI via
  `AGENT_RUN_GROK_TTY_COMMAND` (no real Grok CLI).
- No new durable store under `AGENT_RUN_HOME` for pty stats itself.
- `workdir-at-tmp` kind assumes macOS Go temp (`/var/folders/…/T/`); related
  leaves skip when `t.TempDir()` does not match that layout.

## Steps

1. Root `Setup` builds `agent-run` once per session, sets `AGENT_RUN_HOME`.
2. Grouping / leaf `Setup` sets `Mode`, `Args`, optional `SpawnPlan` / `SpawnServe`.
3. `Run` executes CLI and/or spawn+kill / open+reclaim flows (optional follow-up CLI).
4. Leaf `Assert` checks exit code, stdout keywords / trailing `\n`, and process
   liveness / filter selection.

## Context

- Serve matching: first argv is `IsServeSubcommand` (`__serve_<slug>__`) or
  command line is `agent-run __serve…`.
- Selection after serve match: `--all` → all; else any `--kind` → OR of kind
  predicates; else PPID == 1. Then `--exe` AND. Drop self PID.
- Kind predicates: `TestGenerated` in command/exe; `/var/folders/` and `/T/`
  in command/exe for workdir-at-tmp.
- Successful user-facing CLI stdout must end with trailing `\n`.

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
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
	"path"
)

const (
	envOpenAttachInstant = "AGENT_RUN_OPEN_ATTACH_INSTANT"
	envGrokTTYCommand    = "AGENT_RUN_GROK_TTY_COMMAND"
	defaultServeHoldSecs = 120
	defaultRegistryCreated = "2026-07-03T12:00:00Z"
)

func sessionCacheDir(d *session.Doctest) string {
	return filepath.Join(os.TempDir(), "agent-run-pty-doctest-"+d.DOCTEST_SESSION_ID)
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

func buildOnce(t *testing.T, d *session.Doctest) (agentRun string, err error) {
	t.Helper()
	cache := sessionCacheDir(d)
	agentRun = filepath.Join(cache, "agent-run")
	lock := filepath.Join(cache, "build.lock")
	ready := filepath.Join(cache, "binaries.ready")
	repoRoot, err := findAgentProRoot(d.DOCTEST_ROOT)
	if err != nil {
		return "", err
	}
	err = withFileLock(t, lock, func() error {
		if fileExists(ready) && fileExists(agentRun) {
			return nil
		}
		if err := os.MkdirAll(cache, 0755); err != nil {
			return err
		}
		// Satisfy //go:embed dist if frontend dist is gitignored / empty.
		for _, rel := range []string{"frontend-agent-run/dist", "frontend/dist"} {
			if err := ensureStubDist(filepath.Join(repoRoot, rel)); err != nil {
				return fmt.Errorf("ensure %s stub: %w", rel, err)
			}
		}
		build := exec.Command(runtime.GOROOT()+"/bin/go", "build", "-C", "cmd", "-o", agentRun, "./agent-run")
		build.Dir = repoRoot
		if out, err := build.CombinedOutput(); err != nil {
			return fmt.Errorf("build agent-run: %w\n%s", err, string(out))
		}
		return os.WriteFile(ready, []byte("ok"), 0644)
	})
	return agentRun, err
}

func ensureStubDist(distDir string) error {
	// DistComplete needs non-empty index.html and at least one assets/* file.
	// placeholder.txt alone is not enough; always ensure a minimal SPA shell.
	if err := os.MkdirAll(filepath.Join(distDir, "assets"), 0755); err != nil {
		return err
	}
	const shell = `<!doctype html>
<html lang="en">
<head><meta charset="UTF-8"><title>agent-run</title></head>
<body>
<div id="root"></div>
</body>
</html>
`
	indexPath := filepath.Join(distDir, "index.html")
	needIndex := true
	if data, err := os.ReadFile(indexPath); err == nil {
		s := string(data)
		if strings.Contains(s, `id="root"`) || strings.Contains(s, "id='root'") {
			needIndex = false
		}
	}
	if needIndex {
		if err := os.WriteFile(indexPath, []byte(shell), 0644); err != nil {
			return err
		}
	}
	assetPath := filepath.Join(distDir, "assets", "doctest-stub.js")
	if st, err := os.Stat(assetPath); err != nil || st.Size() == 0 {
		if err := os.WriteFile(assetPath, []byte("/* doctest stub */\n"), 0644); err != nil {
			return err
		}
	}
	return nil
}

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	repoRoot, err := findAgentProRoot(d.DOCTEST_ROOT)
	if err != nil {
		return err
	}
	req.RepoRoot = repoRoot
	req.TempDir = t.TempDir()
	req.Home = filepath.Join(req.TempDir, ".agent-run")
	agentRun, err := buildOnce(t, d)
	if err != nil {
		return err
	}
	// Copy or link into per-leaf bin/ so --exe path is leaf-isolated and
	// still a real agent-run binary for spawn + kill matching.
	req.AgentRun = filepath.Join(req.TempDir, "bin", "agent-run")
	if err := os.MkdirAll(filepath.Dir(req.AgentRun), 0755); err != nil {
		return fmt.Errorf("mkdir bin: %w", err)
	}
	if err := copyFile(agentRun, req.AgentRun); err != nil {
		return fmt.Errorf("install agent-run: %w", err)
	}
	if err := os.Chmod(req.AgentRun, 0755); err != nil {
		return err
	}
	req.Env = append(req.Env, "AGENT_RUN_HOME="+req.Home)
	if req.ExecTimeout <= 0 {
		req.ExecTimeout = 30 * time.Second
	}
	if req.ServeHoldSecs <= 0 {
		req.ServeHoldSecs = defaultServeHoldSecs
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, in, 0755)
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

func processAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	err := syscall.Kill(pid, 0)
	return err == nil
}

func terminatePID(pid int) {
	if pid <= 0 || pid == os.Getpid() || !processAlive(pid) {
		return
	}
	_ = syscall.Kill(pid, syscall.SIGTERM)
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) && processAlive(pid) {
		time.Sleep(50 * time.Millisecond)
	}
	if processAlive(pid) {
		_ = syscall.Kill(pid, syscall.SIGKILL)
	}
	// Best-effort process group (Setsid serve).
	_ = syscall.Kill(-pid, syscall.SIGKILL)
}

// tempDirLooksLikeWorkdirAtTmp reports whether path matches the workdir-at-tmp
// kind predicate (macOS Go t.TempDir layout: /var/folders/…/T/…).
func tempDirLooksLikeWorkdirAtTmp(path string) bool {
	return strings.Contains(path, "/var/folders/") && strings.Contains(path, "/T/")
}

// serveBinaryForMarker returns an agent-run path whose command/exe string
// carries the requested kind marker. Copies from req.AgentRun when needed.
func serveBinaryForMarker(t *testing.T, req *Request, marker string) string {
	t.Helper()
	marker = strings.TrimSpace(marker)
	switch marker {
	case "", "workdir-at-tmp":
		// Default leaf binary already lives under t.TempDir() → /var/folders/…/T/ on macOS.
		return req.AgentRun
	case "test-generated":
		dst := filepath.Join(req.TempDir, "TestGeneratedCase", "bin", "agent-run")
		if fileExists(dst) {
			return dst
		}
		if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
			t.Fatalf("mkdir TestGeneratedCase bin: %v", err)
		}
		if err := copyFile(req.AgentRun, dst); err != nil {
			t.Fatalf("copy TestGenerated agent-run: %v", err)
		}
		if err := os.Chmod(dst, 0755); err != nil {
			t.Fatalf("chmod TestGenerated agent-run: %v", err)
		}
		return dst
	default:
		t.Fatalf("unknown PathMarker %q", marker)
		return ""
	}
}

// spawnDetachedServe starts agent-run as a true orphan __serve_* process (PPID 1)
// running `sleep N` so kill-orphans can see a matching executable + serve token.
// Double-fork is required: a mere Setsid child of go test becomes a zombie after
// external SIGKILL (signal 0 still succeeds until parent Wait), which breaks
// kills-matching-exe liveness asserts.
func spawnDetachedServe(t *testing.T, req *Request) (pid int, sessionID string) {
	t.Helper()
	sessionID = strings.TrimSpace(req.ServeSessionID)
	if sessionID == "" {
		sessionID = fmt.Sprintf("pty-orphan-%d-%d", os.Getpid(), time.Now().UnixNano()%1e9)
		req.ServeSessionID = sessionID
	}
	return spawnOrphanServeAt(t, req, req.AgentRun, sessionID)
}

// spawnOrphanServeAt double-forks bin as __serve under PPID 1.
func spawnOrphanServeAt(t *testing.T, req *Request, bin, sessionID string) (pid int, sid string) {
	t.Helper()
	sid = strings.TrimSpace(sessionID)
	if sid == "" {
		sid = fmt.Sprintf("pty-orphan-%d-%d", os.Getpid(), time.Now().UnixNano()%1e9)
	}
	hold := req.ServeHoldSecs
	if hold <= 0 {
		hold = defaultServeHoldSecs
	}
	token := "__serve_sleep_" + strconv.Itoa(hold) + "__"
	pidFile := filepath.Join(req.TempDir, fmt.Sprintf("serve-%s.pid", sid))
	_ = os.Remove(pidFile)

	// perl double-fork → orphan under launchd/PID 1; write grandchild pid then exec.
	perl := `
use strict;
use warnings;
use POSIX qw(setsid);
my ($pidfile, @argv) = @ARGV;
my $pid = fork();
die "fork1: $!" unless defined $pid;
exit 0 if $pid;          # first parent
die "setsid: $!" unless defined setsid();
$pid = fork();
die "fork2: $!" unless defined $pid;
exit 0 if $pid;          # intermediate
open my $fh, ">", $pidfile or die "pidfile: $!";
print $fh $$;
close $fh;
exec { $argv[0] } @argv or die "exec: $!";
`
	args := []string{
		"-e", perl,
		pidFile,
		bin, token, sid, "sleep", strconv.Itoa(hold),
	}
	cmd := exec.Command("perl", args...)
	cmd.Dir = req.TempDir
	cmd.Env = append(os.Environ(), req.Env...)
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Start(); err != nil {
		t.Fatalf("spawn detached serve (perl): %v", err)
	}
	// Reap the first-level perl parent so it is not a zombie of go test.
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		// first parent should have exited immediately after fork
	}

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		b, err := os.ReadFile(pidFile)
		if err == nil {
			p, convErr := strconv.Atoi(strings.TrimSpace(string(b)))
			if convErr == nil && p > 0 && processAlive(p) {
				pid = p
				t.Cleanup(func() { terminatePID(pid) })
				time.Sleep(150 * time.Millisecond)
				return pid, sid
			}
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("spawned orphan serve did not write live pid to %s", pidFile)
	return 0, sid
}

// spawnChildServeAt starts bin as a direct child __serve (PPID = harness, ≠ 1).
// Caller must Wait via cleanup so SIGKILL does not leave a zombie for signal-0 checks.
func spawnChildServeAt(t *testing.T, req *Request, bin, sessionID string) (pid int, sid string) {
	t.Helper()
	sid = strings.TrimSpace(sessionID)
	if sid == "" {
		sid = fmt.Sprintf("pty-child-%d-%d", os.Getpid(), time.Now().UnixNano()%1e9)
	}
	hold := req.ServeHoldSecs
	if hold <= 0 {
		hold = defaultServeHoldSecs
	}
	token := "__serve_sleep_" + strconv.Itoa(hold) + "__"
	cmd := exec.Command(bin, token, sid, "sleep", strconv.Itoa(hold))
	cmd.Dir = req.TempDir
	cmd.Env = append(os.Environ(), req.Env...)
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Start(); err != nil {
		t.Fatalf("spawn child serve: %v", err)
	}
	pid = cmd.Process.Pid
	t.Cleanup(func() {
		terminatePID(pid)
		_ = cmd.Wait()
	})
	// Brief settle so ps can observe the process.
	time.Sleep(150 * time.Millisecond)
	if !processAlive(pid) {
		t.Fatalf("child serve pid %d died immediately", pid)
	}
	return pid, sid
}

// spawnServeSpec starts one ServeSpawnSpec and returns its PID and session id.
func spawnServeSpec(t *testing.T, req *Request, spec ServeSpawnSpec) (pid int, sessionID string) {
	t.Helper()
	bin := serveBinaryForMarker(t, req, spec.PathMarker)
	sessionID = strings.TrimSpace(spec.SessionID)
	if spec.Orphan {
		return spawnOrphanServeAt(t, req, bin, sessionID)
	}
	return spawnChildServeAt(t, req, bin, sessionID)
}


func execCmd(t *testing.T, command string, args []string, dir string, env []string, timeout time.Duration) (*Response, error) {
	t.Helper()
	if timeout <= 0 {
		timeout = 30 * time.Second
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
	return execCmd(t, req.AgentRun, args, req.TempDir, req.Env, req.ExecTimeout)
}

func runKillOrphansFlow(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	spawnPIDs := make(map[string]int)
	var primaryPID int
	var primarySession string

	if len(req.SpawnPlan) > 0 {
		for i, spec := range req.SpawnPlan {
			label := strings.TrimSpace(spec.Label)
			if label == "" {
				label = fmt.Sprintf("serve-%d", i)
			}
			pid, sid := spawnServeSpec(t, req, spec)
			spawnPIDs[label] = pid
			if primaryPID == 0 {
				primaryPID = pid
				primarySession = sid
			}
		}
	} else if req.SpawnServe {
		pid, sid := spawnDetachedServe(t, req)
		primaryPID = pid
		primarySession = sid
		spawnPIDs["orphan"] = pid
		req.SpawnedServePID = pid
		req.SpawnedSessionID = sid
	}

	if primaryPID > 0 {
		req.SpawnedServePID = primaryPID
		if req.SpawnedSessionID == "" {
			req.SpawnedSessionID = primarySession
		}
	}

	resp, err := runAgentRun(t, req, req.Args...)
	if resp == nil {
		resp = &Response{}
	}
	resp.ServePID = primaryPID
	resp.SpawnPIDs = spawnPIDs
	if primaryPID > 0 {
		resp.ServeAliveBefore = true // was alive when CLI ran (spawn just succeeded)
		// Re-check after CLI (dry-run should leave alive; kill should not).
		// Small settle for kill path.
		time.Sleep(100 * time.Millisecond)
		resp.ServeAliveAfter = processAlive(primaryPID)
	}

	if len(req.FollowUpArgs) > 0 {
		follow, followErr := runAgentRun(t, req, req.FollowUpArgs...)
		if followErr != nil && err == nil {
			// Prefer surfacing follow-up transport errors.
			err = followErr
		}
		if follow != nil {
			resp.FollowUpStdout = follow.Stdout
			resp.FollowUpStderr = follow.Stderr
			resp.FollowUpExitCode = follow.ExitCode
		}
	}
	return resp, err
}

func fakeTUIHoldSeconds(sec int) string {
	return fmt.Sprintf(`sh -c 'printf "GROK_TTY_BANNER\nGrok › "; sleep %d'`, sec)
}

func applyOpenEnv(req *Request) {
	if req.OpenInstantAttach {
		setEnvKV(req, envOpenAttachInstant, "1")
	}
	if strings.TrimSpace(req.GrokTTYCommand) != "" {
		setEnvKV(req, envGrokTTYCommand, req.GrokTTYCommand)
	}
}

// reclaimServesUnderHome kills serve PIDs recorded in registry JSON under the
// isolated test home. Mirrors harness t.Cleanup used by open/keep-tty trees.
func reclaimServesUnderHome(home string) {
	for _, sub := range []string{"grok-tty-registry", "codex-tty-registry", "stub-tty-registry"} {
		dir := filepath.Join(home, sub)
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
				continue
			}
			b, err := os.ReadFile(filepath.Join(dir, e.Name()))
			if err != nil {
				continue
			}
			var ent struct {
				PID        int    `json:"pid"`
				ListenAddr string `json:"listen_addr"`
				SessionID  string `json:"session_id"`
			}
			if json.Unmarshal(b, &ent) != nil || ent.PID <= 0 {
				continue
			}
			if ent.PID == os.Getpid() {
				continue
			}
			terminatePID(ent.PID)
			_ = os.Remove(filepath.Join(dir, e.Name()))
		}
	}
}

func readRegistryEntry(home, runner, sessionID string) (pid int, listenAddr string, ok bool) {
	path := filepath.Join(home, runner+"-registry", sessionID+".json")
	b, err := os.ReadFile(path)
	if err != nil {
		return 0, "", false
	}
	var ent struct {
		PID        int    `json:"pid"`
		ListenAddr string `json:"listen_addr"`
	}
	if json.Unmarshal(b, &ent) != nil {
		return 0, "", false
	}
	return ent.PID, ent.ListenAddr, ent.PID > 0 || ent.ListenAddr != ""
}

func parsePrefixedSessionID(stderr, runner string) (string, bool) {
	// e.g. "grok-tty: session-1"
	prefix := runner + ":"
	for _, line := range strings.Split(stderr, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, prefix) {
			id := strings.TrimSpace(strings.TrimPrefix(line, prefix))
			if id != "" {
				return id, true
			}
		}
	}
	return "", false
}

func runOpenCleanupFlow(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	applyOpenEnv(req)
	// Always reclaim test home serves at end (product leak fix under test).
	t.Cleanup(func() { reclaimServesUnderHome(req.Home) })

	args := req.Args
	if len(args) == 0 {
		prompt := req.Prompt
		if prompt == "" {
			prompt = "pty-open-cleanup"
		}
		args = []string{"run", "--agent-runner", "grok-tty", "--open", prompt}
	}
	timeout := req.ExecTimeout
	if timeout < 60*time.Second {
		timeout = 60 * time.Second
	}
	resp, err := execCmd(t, req.AgentRun, args, req.TempDir, req.Env, timeout)
	if resp == nil {
		resp = &Response{}
	}
	if err != nil {
		return resp, err
	}

	id, ok := parsePrefixedSessionID(resp.Stderr, "grok-tty")
	if !ok {
		// Open may still have created a registry entry; scan registry dir.
		id = firstRegistrySessionID(req.Home, "grok-tty")
	}
	resp.TerminalSessionID = id
	req.TerminalSessionID = id

	pid, listen, regOK := readRegistryEntry(req.Home, "grok-tty", id)
	if regOK {
		resp.ServePID = pid
		resp.RegistryListenAddr = listen
	}
	if pid > 0 {
		resp.ServeAliveBefore = processAlive(pid)
	} else if listen != "" {
		// Registry without pid: treat TCP reachability as "serve present".
		resp.ServeAliveBefore = portOpen(listen)
	}

	// Harness reclaim (the behavior under test for open-cleanup).
	reclaimServesUnderHome(req.Home)
	time.Sleep(150 * time.Millisecond)

	if pid > 0 {
		resp.ServeAliveAfter = processAlive(pid)
	} else if listen != "" {
		resp.ServeAliveAfter = portOpen(listen)
	}
	return resp, nil
}

func firstRegistrySessionID(home, runner string) string {
	dir := filepath.Join(home, runner+"-registry")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		return strings.TrimSuffix(e.Name(), ".json")
	}
	return ""
}

func portOpen(addr string) bool {
	if strings.TrimSpace(addr) == "" {
		return false
	}
	conn, err := net.DialTimeout("tcp", addr, 300*time.Millisecond)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

func assertExitCode(t *testing.T, resp *Response, want int) {
	t.Helper()
	if resp == nil {
		t.Fatalf("nil response")
	}
	if resp.ExitCode != want {
		t.Fatalf("expected exit code %d, got %d\nstdout:\n%s\nstderr:\n%s",
			want, resp.ExitCode, resp.Stdout, resp.Stderr)
	}
}

func assertSuccess(t *testing.T, resp *Response) {
	t.Helper()
	assertExitCode(t, resp, 0)
}

func assertContains(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("missing %q in:\n%s", want, got)
	}
}

func assertOutput(t *testing.T, resp *Response, stream string, contains ...string) {
	t.Helper()
	var got string
	switch stream {
	case "stdout":
		got = resp.Stdout
	case "stderr":
		got = resp.Stderr
	default:
		t.Fatalf("unknown stream %q", stream)
	}
	for _, want := range contains {
		assertContains(t, got, want)
	}
}

func assertTrailingNewline(t *testing.T, s, label string) {
	t.Helper()
	if s == "" || !strings.HasSuffix(s, "\n") {
		tail := s
		if len(tail) > 32 {
			tail = tail[len(tail)-32:]
		}
		t.Fatalf("%s must end with trailing newline; last bytes %q", label, tail)
	}
}

func assertContainsAny(t *testing.T, got string, options ...string) {
	t.Helper()
	for _, opt := range options {
		if strings.Contains(strings.ToLower(got), strings.ToLower(opt)) {
			return
		}
	}
	t.Fatalf("expected one of %v in:\n%s", options, got)
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
