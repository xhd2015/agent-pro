## Preconditions
- The repository contains `cmd/fake-codex`.
- Each test runs in a temporary directory.
- The fake runner must not use host Codex configuration.

## Steps
1. Build `cmd/fake-codex` into the test temporary directory.
2. Write any leaf-specific mock config or legacy script.
3. Run the fake Codex binary with the configured arguments.
4. Capture stdout, stderr, exit status, hook log, and timing.

## Context
- The root harness provides helper functions for writing mock configs, legacy scripts, and hook recorder scripts.

```go
import (
    "bytes"
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "testing"
    "time"
)

func Setup(t *testing.T, req *Request) error {
    _ = writeHookRecorder
    _ = writeMockConfig
    _ = writeLegacyScript
    _ = assertSuccess
    _ = assertContains
    _ = assertNotContains
    _ = parseHookEvents
    req.RepoRoot = filepath.Clean(filepath.Join(DOCTEST_ROOT, "../../.."))
    if _, err := os.Stat(filepath.Join(req.RepoRoot, "go.mod")); err != nil {
        return fmt.Errorf("repo root not found: %w", err)
    }
    req.TempDir = t.TempDir()
    req.FakeCodex = filepath.Join(req.TempDir, "fake-codex")
    req.HookLogPath = filepath.Join(req.TempDir, "hooks.jsonl")
    req.MarkerPath = filepath.Join(req.TempDir, "markers.log")
    build := exec.Command("go", "build", "-o", req.FakeCodex, "./cmd/fake-codex")
    build.Dir = req.RepoRoot
    if out, err := build.CombinedOutput(); err != nil {
        return fmt.Errorf("build fake-codex: %w\n%s", err, string(out))
    }
    return nil
}

func writeFile(t *testing.T, path string, content string) {
    t.Helper()
    if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
        t.Fatalf("mkdir: %v", err)
    }
    if err := os.WriteFile(path, []byte(content), 0644); err != nil {
        t.Fatalf("write %s: %v", path, err)
    }
}

func writeHookRecorder(t *testing.T, req *Request, exitCode int) string {
    t.Helper()
    script := filepath.Join(req.TempDir, "hook-recorder.sh")
    content := fmt.Sprintf(`#!/bin/sh
event="$1"
payload="$(cat)"
printf 'hook %%s\n' "$event" >> %q
printf '{"event":"%%s","payload":%%s}\n' "$event" "$payload" >> %q
exit %d
`, req.MarkerPath, req.HookLogPath, exitCode)
    writeFile(t, script, content)
    if err := os.Chmod(script, 0755); err != nil {
        t.Fatalf("chmod hook recorder: %v", err)
    }
    return script
}

func writeMockConfig(t *testing.T, req *Request, body string) {
    t.Helper()
    req.MockConfigPath = filepath.Join(req.TempDir, "mock.json")
    writeFile(t, req.MockConfigPath, body)
}

func writeLegacyScript(t *testing.T, req *Request, body string) {
    t.Helper()
    req.LegacyScriptPath = filepath.Join(req.TempDir, "script.json")
    writeFile(t, req.LegacyScriptPath, body)
}

func assertSuccess(t *testing.T, resp *Response) {
    t.Helper()
    if resp.Err != nil && resp.ExitCode == 0 {
        t.Fatalf("run failed: %v", resp.Err)
    }
    if resp.ExitCode != 0 {
        t.Fatalf("exit code = %d, stderr:\n%s", resp.ExitCode, resp.Stderr)
    }
}

func assertContains(t *testing.T, got string, want string) {
    t.Helper()
    if !strings.Contains(got, want) {
        t.Fatalf("missing %q in:\n%s", want, got)
    }
}

func assertNotContains(t *testing.T, got string, want string) {
    t.Helper()
    if strings.Contains(got, want) {
        t.Fatalf("unexpected %q in:\n%s", want, got)
    }
}

func parseHookEvents(t *testing.T, log string) []map[string]any {
    t.Helper()
    var events []map[string]any
    for _, line := range strings.Split(strings.TrimSpace(log), "\n") {
        if strings.TrimSpace(line) == "" {
            continue
        }
        var event map[string]any
        if err := json.Unmarshal([]byte(line), &event); err != nil {
            t.Fatalf("parse hook line: %v\n%s", err, line)
        }
        events = append(events, event)
    }
    return events
}
```
