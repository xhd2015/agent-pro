package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHandleExecScriptJSON(t *testing.T) {
	scriptPath := writeFakeCodexScript(t, `{
  "events": [
    {
      "type": "item.completed",
      "item": {
        "id": "m1",
        "type": "message",
        "text": "scripted response",
        "status": "completed"
      }
    }
  ]
}`)

	stdout, stderr, err := captureFakeCodexOutput(t, func() error {
		return handle([]string{"exec", "--json", "--script", scriptPath, "hello"})
	})
	if err != nil {
		t.Fatalf("handle: %v\nstderr: %s", err, stderr)
	}
	if !strings.Contains(stdout, `"scripted response"`) {
		t.Fatalf("stdout missing scripted response:\n%s", stdout)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
}

func TestHandleExecScriptFromEnv(t *testing.T) {
	scriptPath := writeFakeCodexScript(t, `{
  "events": [
    {
      "type": "item.completed",
      "item": {
        "id": "m1",
        "type": "message",
        "text": "from env",
        "status": "completed"
      }
    }
  ]
}`)
	t.Setenv("FAKE_CODEX_SCRIPT", scriptPath)

	stdout, _, err := captureFakeCodexOutput(t, func() error {
		return handle([]string{"exec", "--json", "hello"})
	})
	if err != nil {
		t.Fatalf("handle: %v", err)
	}
	if !strings.Contains(stdout, `"from env"`) {
		t.Fatalf("stdout missing env response:\n%s", stdout)
	}
}

func TestHandleExecScriptFlagOverridesEnv(t *testing.T) {
	envPath := writeFakeCodexScript(t, `{
  "events": [{"type": "item.completed", "item": {"id": "m1", "type": "message", "text": "from env", "status": "completed"}}]
}`)
	flagPath := writeFakeCodexScript(t, `{
  "events": [{"type": "item.completed", "item": {"id": "m1", "type": "message", "text": "from flag", "status": "completed"}}]
}`)
	t.Setenv("FAKE_CODEX_SCRIPT", envPath)

	stdout, _, err := captureFakeCodexOutput(t, func() error {
		return handle([]string{"exec", "--json", "--script", flagPath, "hello"})
	})
	if err != nil {
		t.Fatalf("handle: %v", err)
	}
	if !strings.Contains(stdout, `"from flag"`) {
		t.Fatalf("stdout missing flag response:\n%s", stdout)
	}
	if strings.Contains(stdout, `"from env"`) {
		t.Fatalf("stdout used env script despite flag override:\n%s", stdout)
	}
}

func TestHandleExecScriptExitCode(t *testing.T) {
	scriptPath := writeFakeCodexScript(t, `{
  "events": [
    {
      "type": "item.completed",
      "item": {
        "id": "m1",
        "type": "message",
        "text": "before failure",
        "status": "completed"
      }
    }
  ],
  "exit_code": 7,
  "stderr": "planned failure"
}`)

	stdout, stderr, err := captureFakeCodexOutput(t, func() error {
		return handle([]string{"exec", "--json", "--script", scriptPath, "hello"})
	})
	if err == nil {
		t.Fatal("expected exit error")
	}
	exitErr, ok := err.(*exitError)
	if !ok {
		t.Fatalf("err = %T %v, want *exitError", err, err)
	}
	if exitErr.Code != 7 {
		t.Fatalf("exit code = %d, want 7", exitErr.Code)
	}
	if !strings.Contains(stdout, `"before failure"`) {
		t.Fatalf("stdout missing scripted event:\n%s", stdout)
	}
	if !strings.Contains(stderr, "planned failure") {
		t.Fatalf("stderr missing planned failure:\n%s", stderr)
	}
}

func writeFakeCodexScript(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "script.json")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write script: %v", err)
	}
	return path
}

func captureFakeCodexOutput(t *testing.T, fn func() error) (string, string, error) {
	t.Helper()
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		t.Fatalf("stdout pipe: %v", err)
	}
	stderrR, stderrW, err := os.Pipe()
	if err != nil {
		t.Fatalf("stderr pipe: %v", err)
	}
	os.Stdout = stdoutW
	os.Stderr = stderrW

	runErr := fn()

	_ = stdoutW.Close()
	_ = stderrW.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	stdoutBytes, err := io.ReadAll(stdoutR)
	if err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	stderrBytes, err := io.ReadAll(stderrR)
	if err != nil {
		t.Fatalf("read stderr: %v", err)
	}
	return string(stdoutBytes), string(stderrBytes), runErr
}
