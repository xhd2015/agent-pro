# Scenario

**Feature**: shared harness for agent-run web workspace path selector doctests

```
# build binary, isolated AGENT_RUN_HOME, free port, optional process cwd
build agent-run -> temp AGENT_RUN_HOME
  -> agent-run web --port <free> --no-open  (cmd.Dir = WebWorkingDir when set)
  -> Mode=http multi-step API  |  playwright-debug UI @ 390×844
```

## Preconditions

- Repository contains `cmd/agent-run` and `go.mod` at repo root.
- `frontend-agent-run/dist` may be gitignored. Root setup ensures a minimal SPA
  `index.html` with `#root` so `//go:embed dist` compiles. Real Vite builds
  (when present) are left untouched — **ui/** leaves need a real SPA for
  React Router `/workspace` and draft survival (TDD RED until implemented).
- Each leaf uses isolated `AGENT_RUN_HOME` under `t.TempDir()`.
- Playwright leaves require `playwright-debug` on `PATH` (skip otherwise).
- Server binds **127.0.0.1** only; tests use `http://127.0.0.1:<port>/`.
- MRU cap under test: **12** (`maxRecentWorkspaces`).

## Steps

1. Root `Setup` resolves repo root, ensures dist shell, builds `agent-run`, sets home/env.
2. Grouping `Setup` sets `Mode` (`http` vs `ui`) and default token mode.
3. Leaf `Setup` creates fixture dirs, optional config seed, starts web, sets `HTTPSteps` or `PlaywrightScript`.
4. `Run` executes multi-step HTTP or `playwright-debug -e <script>`.
5. Leaf `Assert` checks status codes, JSON fields, config.json, or playwright exit.

## Context

- Default `WebTokenMode`: `explicit` with `Token` `test-token` unless a leaf sets `omit`.
- Session storage is **flat**: `AGENT_RUN_HOME/sessions/<session_id>/`.
- Config path: `AGENT_RUN_HOME/config.json` with `selected_workspace` + `recent_workspaces`.
- Product rule: only **Use this folder** / `PUT /workspace` commits; chips/recent browse only.
- SPA routes: `/`, `/workspace`, `/sessions/:id`, `*`.

```go
import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// maxRecentWorkspaces is the product MRU cap (≈12).
const maxRecentWorkspaces = 12

func Setup(t *testing.T, req *Request) error {
	req.RepoRoot = filepath.Clean(filepath.Join(DOCTEST_ROOT, "../../../.."))
	if _, err := os.Stat(filepath.Join(req.RepoRoot, "go.mod")); err != nil {
		return fmt.Errorf("repo root not found: %w", err)
	}
	req.TempDir = t.TempDir()
	req.Home = filepath.Join(req.TempDir, ".agent-run")
	if req.WebTokenMode == "" {
		req.WebTokenMode = "explicit"
	}
	if req.WebTokenMode == "explicit" && req.Token == "" {
		req.Token = "test-token"
	}
	if home, err := os.UserHomeDir(); err == nil {
		req.OSUserHome = home
	} else {
		req.OSUserHome = os.Getenv("HOME")
	}
	if err := ensureFrontendDistShell(req.RepoRoot); err != nil {
		return err
	}
	req.AgentRun = filepath.Join(req.TempDir, "bin", "agent-run")
	if err := buildAgentRun(t, req.RepoRoot, req.AgentRun); err != nil {
		return err
	}
	req.Env = append(req.Env,
		"AGENT_RUN_HOME="+req.Home,
		"PATH="+filepath.Dir(req.AgentRun)+string(os.PathListSeparator)+os.Getenv("PATH"),
	)
	return nil
}

func ensureFrontendDistShell(repoRoot string) error {
	distDir := filepath.Join(repoRoot, "frontend-agent-run", "dist")
	indexPath := filepath.Join(distDir, "index.html")
	if st, err := os.Stat(indexPath); err == nil && !st.IsDir() {
		return nil
	}
	if err := os.MkdirAll(distDir, 0755); err != nil {
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
	return os.WriteFile(indexPath, []byte(shell), 0644)
}

func buildAgentRun(t *testing.T, repoRoot, out string) error {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
		return fmt.Errorf("mkdir bin: %w", err)
	}
	cmd := exec.Command("go", "build", "-buildvcs=false", "-o", out, "./cmd/agent-run")
	cmd.Dir = repoRoot
	if outBytes, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("build agent-run: %w\n%s", err, string(outBytes))
	}
	return nil
}

func selectorWebTokenMode(req *Request) string {
	switch req.WebTokenMode {
	case "omit", "auto", "explicit":
		return req.WebTokenMode
	default:
		return "explicit"
	}
}

func waitForHealth(baseURL, bearer string, timeout time.Duration) error {
	client := &http.Client{Timeout: 2 * time.Second}
	deadline := time.Now().Add(timeout)
	url := strings.TrimRight(baseURL, "/") + "/api/agent-run/health"
	for time.Now().Before(deadline) {
		httpReq, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return err
		}
		if bearer != "" {
			httpReq.Header.Set("Authorization", "Bearer "+bearer)
		}
		resp, err := client.Do(httpReq)
		if err == nil {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("health check timed out after %s: %s", timeout, url)
}

func parseSelectorTokenFromStderr(stderr string) string {
	for _, line := range strings.Split(stderr, "\n") {
		line = strings.TrimSpace(line)
		const prefix = "agent-run web token: "
		if strings.HasPrefix(line, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(line, prefix))
		}
	}
	return ""
}

func findFreePort(t *testing.T) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("find free port: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()
	return port
}

func stopWeb(t *testing.T, req *Request) {
	t.Helper()
	if req.webCmd == nil || req.webCmd.Process == nil {
		return
	}
	_ = req.webCmd.Process.Signal(os.Interrupt)
	done := make(chan error, 1)
	go func() { done <- req.webCmd.Wait() }()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		_ = req.webCmd.Process.Kill()
		<-done
	}
	req.webCmd = nil
}

func startWebBackground(t *testing.T, req *Request) error {
	t.Helper()
	if req.Port == 0 {
		req.Port = findFreePort(t)
	}
	if req.AgentRun == "" {
		return fmt.Errorf("AgentRun binary not built")
	}

	args := []string{"web", "--port", strconv.Itoa(req.Port), "--no-open"}
	switch selectorWebTokenMode(req) {
	case "omit":
		// open API
	case "auto":
		args = append(args, "--token", "auto")
	default:
		if req.Token == "" {
			req.Token = "test-token"
		}
		args = append(args, "--token", req.Token)
	}
	var stderrBuf bytes.Buffer
	cmd := exec.Command(req.AgentRun, args...)
	if req.WebWorkingDir != "" {
		cmd.Dir = req.WebWorkingDir
	}
	cmd.Env = append(os.Environ(), req.Env...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start agent-run web: %w", err)
	}
	req.webCmd = cmd
	req.BaseURL = "http://127.0.0.1:" + strconv.Itoa(req.Port)
	t.Cleanup(func() { stopWeb(t, req) })

	mode := selectorWebTokenMode(req)
	switch mode {
	case "omit":
		if err := waitForHealth(req.BaseURL, "", 30*time.Second); err != nil {
			stopWeb(t, req)
			return err
		}
	case "auto":
		deadline := time.Now().Add(30 * time.Second)
		for time.Now().Before(deadline) {
			if tok := parseSelectorTokenFromStderr(stderrBuf.String()); tok != "" {
				req.Token = tok
				if err := waitForHealth(req.BaseURL, tok, 2*time.Second); err == nil {
					return nil
				}
			}
			time.Sleep(50 * time.Millisecond)
		}
		stopWeb(t, req)
		return fmt.Errorf("health check timed out (auto token), stderr:\n%s", stderrBuf.String())
	default:
		if err := waitForHealth(req.BaseURL, req.Token, 30*time.Second); err != nil {
			stopWeb(t, req)
			return err
		}
	}
	return nil
}

func requirePlaywright(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("playwright-debug"); err != nil {
		t.Skipf("skipping web-workspace-selector UI test: playwright-debug not on PATH: %v", err)
	}
}

func writeHomeConfig(t *testing.T, home string, cfg map[string]any) error {
	t.Helper()
	if err := os.MkdirAll(home, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(home, "config.json"), data, 0644)
}

func readHomeConfigMap(t *testing.T, home string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(home, "config.json"))
	if err != nil {
		return nil
	}
	var m map[string]any
	if json.Unmarshal(data, &m) != nil {
		return nil
	}
	return m
}

func putWorkspaceBody(path string) string {
	b, _ := json.Marshal(map[string]string{"path": path})
	return string(b)
}

func mustMkdir(t *testing.T, path string) string {
	t.Helper()
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		t.Fatalf("abs %s: %v", path, err)
	}
	return abs
}

// makeSelectDir creates a unique absolute workspace directory under TempDir.
func makeSelectDir(t *testing.T, req *Request, name string) string {
	t.Helper()
	return mustMkdir(t, filepath.Join(req.TempDir, "workspaces", name))
}

// makeFixtureTree creates dirs + a file under FixtureRoot for fs/list leaves.
// Layout:
//
//	<root>/
//	  subdir/
//	  note.txt
func makeFixtureTree(t *testing.T, req *Request) string {
	t.Helper()
	root := mustMkdir(t, filepath.Join(req.TempDir, "fixture-fs"))
	_ = mustMkdir(t, filepath.Join(root, "subdir"))
	note := filepath.Join(root, "note.txt")
	if err := os.WriteFile(note, []byte("hello\n"), 0644); err != nil {
		t.Fatalf("write note: %v", err)
	}
	req.FixtureRoot = root
	return root
}

// makeChooserOptimizeFixture creates a tree for dir-chooser optimize leaves
// (dot dirs/files, sort, UI hide/show files).
// Layout:
//
//	<root>/
//	  .git/            (dir)
//	  .hidden-dir/     (dir)
//	  src/             (dir)
//	  src/inner.txt    (file)
//	  .env             (file)
//	  a.txt            (file)
//	  note.txt         (file)
func makeChooserOptimizeFixture(t *testing.T, req *Request) string {
	t.Helper()
	root := mustMkdir(t, filepath.Join(req.TempDir, "fixture-chooser"))
	_ = mustMkdir(t, filepath.Join(root, ".git"))
	_ = mustMkdir(t, filepath.Join(root, ".hidden-dir"))
	src := mustMkdir(t, filepath.Join(root, "src"))
	for _, pair := range []struct {
		rel  string
		data string
	}{
		{".env", "SECRET=1\n"},
		{"a.txt", "a\n"},
		{"note.txt", "hello\n"},
		{filepath.Join("src", "inner.txt"), "inner\n"},
	} {
		p := filepath.Join(root, pair.rel)
		if err := os.WriteFile(p, []byte(pair.data), 0644); err != nil {
			t.Fatalf("write %s: %v", p, err)
		}
	}
	_ = src
	req.FixtureRoot = root
	return root
}

// makeDirsOnlyFixture creates a directory with only subdirs (no files).
// Used to assert the show-files control is omitted when there are zero files.
func makeDirsOnlyFixture(t *testing.T, req *Request) string {
	t.Helper()
	root := mustMkdir(t, filepath.Join(req.TempDir, "fixture-dirs-only"))
	_ = mustMkdir(t, filepath.Join(root, "alpha"))
	_ = mustMkdir(t, filepath.Join(root, "beta"))
	req.FixtureRoot = root
	return root
}

func fsListPath(dir string) string {
	return "/api/agent-run/fs/list?path=" + url.QueryEscape(dir)
}

func findHTTPResult(resp *Response, name string) (HTTPResult, bool) {
	if resp == nil {
		return HTTPResult{}, false
	}
	for _, r := range resp.HTTPResults {
		if r.Name == name {
			return r, true
		}
	}
	return HTTPResult{}, false
}

func parseJSONMap(t *testing.T, body string) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal([]byte(body), &m); err != nil {
		t.Fatalf("parse JSON: %v body=%q", err, truncate(body, 300))
	}
	return m
}

func jsonStringField(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	v, _ := m[key].(string)
	return strings.TrimSpace(v)
}

func jsonStringSlice(m map[string]any, key string) []string {
	if m == nil {
		return nil
	}
	raw, ok := m[key]
	if !ok || raw == nil {
		return nil
	}
	switch xs := raw.(type) {
	case []any:
		out := make([]string, 0, len(xs))
		for _, x := range xs {
			if s, ok := x.(string); ok {
				out = append(out, s)
			}
		}
		return out
	case []string:
		return xs
	default:
		return nil
	}
}

func cleanPath(p string) string {
	return filepath.Clean(strings.TrimSpace(p))
}

func pathsEqual(a, b string) bool {
	return cleanPath(a) == cleanPath(b)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func seedTokenInPage(token string) string {
	escaped := strings.ReplaceAll(token, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `'`, `\'`)
	return fmt.Sprintf(`
await page.addInitScript(() => {
  localStorage.setItem('agent-run-token', '%s');
});
`, escaped)
}

func mobileViewportScript(body string) string {
	return fmt.Sprintf(`
await page.setViewportSize({ width: 390, height: 844 });
%s
`, body)
}

func jsString(s string) string {
	b, _ := json.Marshal(s)
	// json.Marshal produces a quoted string; strip quotes for embedding in JS single-quoted literals carefully.
	// Prefer double-quoted JS string from JSON.
	return string(b)
}

func assertPlaywrightOK(t *testing.T, resp *Response, err error) {
	t.Helper()
	if err != nil {
		stderr := ""
		if resp != nil {
			stderr = resp.PlaywrightStderr
		}
		t.Fatalf("Run error: %v\nstderr:\n%s", err, stderr)
	}
	if resp == nil {
		t.Fatal("nil Response")
	}
	if resp.PlaywrightExit != 0 {
		t.Fatalf("playwright exit=%d\nstdout:\n%s\nstderr:\n%s",
			resp.PlaywrightExit, resp.PlaywrightStdout, resp.PlaywrightStderr)
	}
}
```
