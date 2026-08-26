# Scenario

**Feature**: explain `--agent-runner` CLI (e2e binary)

```
build explain once; isolate config home; run explain with cmd.Env
```

## Preconditions

- `go` available; `cmd/explain` exists.
- Session cache under `$TMPDIR/explain-agent-runner-doctest-<session>/`.

## Steps

1. Build `explain` once per doctest session (file-locked).
2. Ensure failing stub + working fake-opencode scripts exist.
3. Isolate `ConfigHome` per leaf.
4. Leaves set `Args` (and optionally `WorkingAgentPath`).

```go
import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"
)

type sessionFileData struct {
	AgentRunner      string                     `json:"agent_runner"`
	AgentRunnersMeta map[string]json.RawMessage `json:"agent_runners_meta"`
}

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skipf("skipping: go not found in PATH")
	}
	repoRoot, err := findModuleRoot(d.DOCTEST_ROOT)
	if err != nil {
		return err
	}
	req.RepoRoot = repoRoot

	bin, err := buildExplainOnce(t, d)
	if err != nil {
		return err
	}
	req.Bin = bin

	failing, err := ensureFailingFakeAgent(t, d)
	if err != nil {
		return err
	}
	req.FakeAgentPath = failing

	working, err := ensureWorkingFakeOpencode(t, d)
	if err != nil {
		return err
	}
	_ = working // leaves opt-in via WorkingAgentPath

	req.ConfigHome = filepath.Join(t.TempDir(), "explain-config-home")
	return os.MkdirAll(req.ConfigHome, 0o755)
}

func buildEnv(req *Request) []string {
	strip := map[string]bool{
		"AGENT_PRO_DEDICATED_AGENT_EXPLAIN_DEBUG_CONFIG_HOME": true,
		"EXPLAIN_AGENT_PATH": true,
		"NO_COLOR":           true,
		"FORCE_COLOR":        true,
		"CLICOLOR_FORCE":     true,
	}
	var env []string
	for _, e := range os.Environ() {
		key, _, _ := strings.Cut(e, "=")
		if strip[key] {
			continue
		}
		env = append(env, e)
	}
	agentPath := req.FakeAgentPath
	if req.WorkingAgentPath != "" {
		agentPath = req.WorkingAgentPath
	}
	env = append(env,
		"AGENT_PRO_DEDICATED_AGENT_EXPLAIN_DEBUG_CONFIG_HOME="+req.ConfigHome,
		"EXPLAIN_AGENT_PATH="+agentPath,
		"NO_COLOR=1",
	)
	env = append(env, req.EnvExtra...)
	return env
}

func readSessionFile(path string) (sessionFileData, error) {
	var data sessionFileData
	raw, err := os.ReadFile(path)
	if err != nil {
		return data, err
	}
	if err := json.Unmarshal(raw, &data); err != nil {
		return data, err
	}
	return data, nil
}

func assertExitCode(t *testing.T, resp *Response, want int) {
	t.Helper()
	if resp.ExitCode != want {
		t.Fatalf("exit code = %d, want %d\nstdout:\n%s\nstderr:\n%s",
			resp.ExitCode, want, resp.Stdout, resp.Stderr)
	}
}

func assertContains(t *testing.T, s, substr string) {
	t.Helper()
	if !strings.Contains(s, substr) {
		t.Fatalf("missing %q in:\n%s", substr, s)
	}
}

func sessionCacheDir(sessionID string) string {
	return filepath.Join(os.TempDir(), "explain-agent-runner-doctest-"+sessionID)
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

func findModuleRoot(start string) (string, error) {
	if start == "" {
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		start = wd
	}
	for dir := start; ; dir = filepath.Dir(dir) {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			if _, err := os.Stat(filepath.Join(dir, "cmd", "explain")); err == nil {
				return dir, nil
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not find module root above %s", start)
		}
	}
}

func buildExplainOnce(t *testing.T, d *session.Doctest) (string, error) {
	t.Helper()
	cache := sessionCacheDir(d.DOCTEST_SESSION_ID)
	bin := filepath.Join(cache, "explain")
	ready := filepath.Join(cache, "binaries.ready")
	lock := filepath.Join(cache, "build.lock")

	err := withFileLock(lock, func() error {
		if fileExists(ready) && fileExists(bin) {
			return nil
		}
		repoRoot, err := findModuleRoot(d.DOCTEST_ROOT)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(cache, 0o755); err != nil {
			return err
		}
		cmdDir := filepath.Join(repoRoot, "cmd")
		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, runtime.GOROOT()+"/bin/go", "build", "-o", bin, "./explain")
		cmd.Dir = cmdDir
		var be bytes.Buffer
		cmd.Stderr = &be
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("go build ./explain: %w\n%s", err, be.String())
		}
		return os.WriteFile(ready, []byte("ok\n"), 0o644)
	})
	if err != nil {
		return "", err
	}
	return bin, nil
}

func ensureFailingFakeAgent(t *testing.T, d *session.Doctest) (string, error) {
	t.Helper()
	cache := sessionCacheDir(d.DOCTEST_SESSION_ID)
	path := filepath.Join(cache, "fake-agent-fail")
	lock := filepath.Join(cache, "fake-fail.lock")
	ready := filepath.Join(cache, "fake-fail.ready")
	err := withFileLock(lock, func() error {
		if fileExists(ready) && fileExists(path) {
			return nil
		}
		if err := os.MkdirAll(cache, 0o755); err != nil {
			return err
		}
		script := "#!/bin/sh\necho FAKE_AGENT_INVOKED >&2\nexit 99\n"
		if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
			return err
		}
		return os.WriteFile(ready, []byte("ok\n"), 0o644)
	})
	if err != nil {
		return "", err
	}
	return path, nil
}

func ensureWorkingFakeOpencode(t *testing.T, d *session.Doctest) (string, error) {
	t.Helper()
	cache := sessionCacheDir(d.DOCTEST_SESSION_ID)
	path := filepath.Join(cache, "fake-opencode-ok")
	lock := filepath.Join(cache, "fake-ok.lock")
	ready := filepath.Join(cache, "fake-ok.ready")
	err := withFileLock(lock, func() error {
		if fileExists(ready) && fileExists(path) {
			return nil
		}
		if err := os.MkdirAll(cache, 0o755); err != nil {
			return err
		}
		script := `#!/bin/bash
ARG1="$1"
shift
case "$ARG1" in
  run)
    echo '{"type":"step_start","sessionID":"fake-sess-1","timestamp":0}'
    echo '{"type":"text","part":{"text":"[MOCK OK]"}}'
    echo '{"type":"step_finish","sessionID":"fake-sess-1","timestamp":0}'
    ;;
  models)
    echo "mock-model"
    ;;
  *)
    exit 1
    ;;
esac
`
		if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
			return err
		}
		return os.WriteFile(ready, []byte("ok\n"), 0o644)
	})
	if err != nil {
		return "", err
	}
	return path, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
```
