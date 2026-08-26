package explain

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/terminal/color"
)

func TestLoadConfig_MissingReturnsEmpty(t *testing.T) {
	home := t.TempDir()
	t.Setenv(debugConfigHomeEnv, home)

	cfg, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}
	if cfg.AgentRunner != "" || cfg.Model != "" || cfg.Version != 0 {
		t.Fatalf("expected empty config, got %+v", cfg)
	}
}

func TestLoadConfig_CorruptFailsHard(t *testing.T) {
	home := t.TempDir()
	t.Setenv(debugConfigHomeEnv, home)
	if err := os.WriteFile(filepath.Join(home, configFileName), []byte("{not-json"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := loadConfig()
	if err == nil {
		t.Fatal("expected parse error")
	}
	if !strings.Contains(err.Error(), "parse config.json") {
		t.Fatalf("error = %q, want parse config.json", err.Error())
	}
}

func TestSetConfig_MergeAndClear(t *testing.T) {
	home := t.TempDir()
	t.Setenv(debugConfigHomeEnv, home)

	// Preserve unknown keys across merge writes.
	if err := os.WriteFile(filepath.Join(home, configFileName), []byte(`{"version":1,"extra":true}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := runSetConfig(setConfigPrefs{agentRunner: "codex", model: "m1"}); err != nil {
		t.Fatalf("set: %v", err)
	}
	cfg, err := loadConfig()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.AgentRunner != "codex" || cfg.Model != "m1" {
		t.Fatalf("cfg = %+v", cfg)
	}
	root, err := loadConfigMap()
	if err != nil {
		t.Fatal(err)
	}
	if root["extra"] != true {
		t.Fatalf("expected extra preserved, got %#v", root["extra"])
	}

	if err := runSetConfig(setConfigPrefs{clearAgentRunner: true}); err != nil {
		t.Fatalf("clear runner: %v", err)
	}
	cfg, err = loadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AgentRunner != "" {
		t.Fatalf("agent_runner still %q", cfg.AgentRunner)
	}
	if cfg.Model != "m1" {
		t.Fatalf("model cleared unexpectedly: %q", cfg.Model)
	}

	if err := runSetConfig(setConfigPrefs{clearModel: true}); err != nil {
		t.Fatalf("clear model: %v", err)
	}
	cfg, err = loadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Model != "" {
		t.Fatalf("model still %q", cfg.Model)
	}
}

func TestSetConfig_RequiresPreference(t *testing.T) {
	home := t.TempDir()
	t.Setenv(debugConfigHomeEnv, home)
	err := runSetConfig(setConfigPrefs{})
	if err == nil || !strings.Contains(err.Error(), "--set-config requires") {
		t.Fatalf("error = %v", err)
	}
}

func TestSetConfig_RejectsUnknownRunner(t *testing.T) {
	home := t.TempDir()
	t.Setenv(debugConfigHomeEnv, home)
	err := runSetConfig(setConfigPrefs{agentRunner: "nope"})
	if err == nil || !strings.Contains(err.Error(), "unsupported agent runner") {
		t.Fatalf("error = %v", err)
	}
}

func TestShowConfig_MissingPrintsEmptyObject(t *testing.T) {
	home := t.TempDir()
	t.Setenv(debugConfigHomeEnv, home)

	rOut, wOut, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = wOut
	showErr := runShowConfig()
	_ = wOut.Close()
	os.Stdout = old
	if showErr != nil {
		t.Fatalf("show: %v", showErr)
	}
	var buf [64]byte
	n, _ := rOut.Read(buf[:])
	_ = rOut.Close()
	got := strings.TrimSpace(string(buf[:n]))
	if got != "{}" {
		t.Fatalf("stdout = %q, want {}", got)
	}
}

func TestRunExplain_UsesConfigAgentRunner(t *testing.T) {
	home := t.TempDir()
	t.Setenv(debugConfigHomeEnv, home)
	t.Setenv("HOME", t.TempDir())

	if err := runSetConfig(setConfigPrefs{agentRunner: "codex", model: "from-config"}); err != nil {
		t.Fatal(err)
	}

	mock := &mockRunner{
		startOutputs: []mockRunOutput{{sessionID: "sess-cfg", output: "ok"}},
	}
	if err := RunExplainWithRunner([]string{"hello-config"}, mock); err != nil {
		t.Fatalf("RunExplainWithRunner: %v", err)
	}
	if mock.lastModel != "from-config" {
		t.Fatalf("model = %q, want from-config", mock.lastModel)
	}

	base := filepath.Join(home, "sessions")
	entries, err := os.ReadDir(base)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("sessions = %d", len(entries))
	}
	data, err := readSession(filepath.Join(base, entries[0].Name()))
	if err != nil {
		t.Fatal(err)
	}
	if data.AgentRunner != "codex" {
		t.Fatalf("AgentRunner = %q, want codex", data.AgentRunner)
	}
}

func TestRunExplain_FlagOverridesConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv(debugConfigHomeEnv, home)
	t.Setenv("HOME", t.TempDir())

	if err := runSetConfig(setConfigPrefs{agentRunner: "codex", model: "cfg-model"}); err != nil {
		t.Fatal(err)
	}

	mock := &mockRunner{
		startOutputs: []mockRunOutput{{sessionID: "sess-flag", output: "ok"}},
	}
	err := RunExplainWithRunner([]string{"--agent-runner", "grok", "--model", "cli-model", "hello-flag"}, mock)
	if err != nil {
		t.Fatal(err)
	}
	if mock.lastModel != "cli-model" {
		t.Fatalf("model = %q, want cli-model", mock.lastModel)
	}
	base := filepath.Join(home, "sessions")
	entries, err := os.ReadDir(base)
	if err != nil {
		t.Fatal(err)
	}
	data, err := readSession(filepath.Join(base, entries[0].Name()))
	if err != nil {
		t.Fatal(err)
	}
	if data.AgentRunner != "grok" {
		t.Fatalf("AgentRunner = %q, want grok", data.AgentRunner)
	}
}

func TestRunExplain_NoConfigSkipsFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv(debugConfigHomeEnv, home)
	t.Setenv("HOME", t.TempDir())

	if err := runSetConfig(setConfigPrefs{agentRunner: "codex"}); err != nil {
		t.Fatal(err)
	}

	mock := &mockRunner{
		startOutputs: []mockRunOutput{{sessionID: "sess-nocfg", output: "ok"}},
	}
	if err := RunExplainWithRunner([]string{"--no-config", "hello-nocfg"}, mock); err != nil {
		t.Fatal(err)
	}
	base := filepath.Join(home, "sessions")
	entries, err := os.ReadDir(base)
	if err != nil {
		t.Fatal(err)
	}
	data, err := readSession(filepath.Join(base, entries[0].Name()))
	if err != nil {
		t.Fatal(err)
	}
	if data.AgentRunner != "opencode" {
		t.Fatalf("AgentRunner = %q, want opencode (built-in)", data.AgentRunner)
	}
}

func TestRunExplain_SetConfigAndShowConfigCLI(t *testing.T) {
	home := t.TempDir()
	t.Setenv(debugConfigHomeEnv, home)

	if err := RunExplainWithRunner([]string{"--set-config", "--agent-runner", "commandcode"}, &mockRunner{}); err != nil {
		t.Fatalf("set-config: %v", err)
	}
	raw, err := os.ReadFile(filepath.Join(home, configFileName))
	if err != nil {
		t.Fatal(err)
	}
	var cfg Config
	if err := json.Unmarshal(raw, &cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.AgentRunner != "commandcode" {
		t.Fatalf("cfg.AgentRunner = %q", cfg.AgentRunner)
	}

	err = RunExplainWithRunner([]string{"--set-config", "--show-config"}, &mockRunner{})
	if err == nil || !strings.Contains(err.Error(), "mutually exclusive") {
		t.Fatalf("expected mutex error, got %v", err)
	}

	err = RunExplainWithRunner([]string{"--set-config"}, &mockRunner{})
	if err == nil || !strings.Contains(err.Error(), "--set-config requires") {
		t.Fatalf("expected requires error, got %v", err)
	}

	err = RunExplainWithRunner([]string{"--show-config", "hello"}, &mockRunner{})
	if err == nil || !strings.Contains(err.Error(), "does not take a message") {
		t.Fatalf("expected show message error, got %v", err)
	}

	err = RunExplainWithRunner([]string{"--clear-agent-runner"}, &mockRunner{})
	if err == nil || !strings.Contains(err.Error(), "require --set-config") {
		t.Fatalf("expected clear requires set-config, got %v", err)
	}
}

func TestExplainHelpListsConfigFlags(t *testing.T) {
	t.Parallel()
	for _, flag := range []string{"--set-config", "--show-config", "--no-config", "--clear-agent-runner", "--clear-model", "--color", "--no-color"} {
		if !strings.Contains(explainHelp, flag) {
			t.Fatalf("help missing %s:\n%s", flag, explainHelp)
		}
	}
}

func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stderr
	os.Stderr = w
	fn()
	_ = w.Close()
	os.Stderr = old
	var buf [4096]byte
	n, _ := r.Read(buf[:])
	_ = r.Close()
	return string(buf[:n])
}

func TestPrintAgentRunnerFromConfigNotice_ColorModes(t *testing.T) {
	t.Parallel()
	var buf strings.Builder
	printAgentRunnerFromConfigNotice(&buf, color.Always, "codex")
	got := buf.String()
	wantPrefix := "\x1b[90mnotice:\x1b[0m agent-runner=codex (from config)\n"
	if got != wantPrefix {
		t.Fatalf("Always: got %q, want %q", got, wantPrefix)
	}

	buf.Reset()
	printAgentRunnerFromConfigNotice(&buf, color.Never, "codex")
	got = buf.String()
	if got != "notice: agent-runner=codex (from config)\n" {
		t.Fatalf("Never: got %q", got)
	}
}

func TestRunExplain_NoticeWhenRunnerFromConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv(debugConfigHomeEnv, home)
	t.Setenv("HOME", t.TempDir())

	if err := runSetConfig(setConfigPrefs{agentRunner: "codex"}); err != nil {
		t.Fatal(err)
	}

	mock := &mockRunner{
		startOutputs: []mockRunOutput{{sessionID: "sess-notice", output: "ok"}},
	}
	stderr := captureStderr(t, func() {
		if err := RunExplainWithRunner([]string{"--no-color", "hello-notice"}, mock); err != nil {
			t.Fatalf("RunExplainWithRunner: %v", err)
		}
	})
	if !strings.Contains(stderr, "notice: agent-runner=codex (from config)") {
		t.Fatalf("stderr missing notice:\n%s", stderr)
	}
	if strings.Contains(stderr, "\x1b[") {
		t.Fatalf("expected no ANSI with --no-color:\n%q", stderr)
	}
}

func TestRunExplain_NoticeColoredWithColorFlag(t *testing.T) {
	home := t.TempDir()
	t.Setenv(debugConfigHomeEnv, home)
	t.Setenv("HOME", t.TempDir())

	if err := runSetConfig(setConfigPrefs{agentRunner: "grok"}); err != nil {
		t.Fatal(err)
	}

	mock := &mockRunner{
		startOutputs: []mockRunOutput{{sessionID: "sess-color", output: "ok"}},
	}
	stderr := captureStderr(t, func() {
		if err := RunExplainWithRunner([]string{"--color", "hello-color"}, mock); err != nil {
			t.Fatalf("RunExplainWithRunner: %v", err)
		}
	})
	if !strings.Contains(stderr, "\x1b[90mnotice:\x1b[0m agent-runner=grok (from config)") {
		t.Fatalf("stderr missing gray notice:\n%q", stderr)
	}
}

func TestRunExplain_NoNoticeWhenFlagOrNoConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv(debugConfigHomeEnv, home)
	t.Setenv("HOME", t.TempDir())

	if err := runSetConfig(setConfigPrefs{agentRunner: "codex"}); err != nil {
		t.Fatal(err)
	}

	mock := &mockRunner{
		startOutputs: []mockRunOutput{{sessionID: "sess-x", output: "ok"}},
	}
	stderr := captureStderr(t, func() {
		if err := RunExplainWithRunner([]string{"--agent-runner", "grok", "hello-flag"}, mock); err != nil {
			t.Fatalf("flag override: %v", err)
		}
	})
	if strings.Contains(stderr, "notice:") {
		t.Fatalf("flag override should not notice:\n%s", stderr)
	}

	stderr = captureStderr(t, func() {
		if err := RunExplainWithRunner([]string{"--no-config", "hello-nocfg"}, mock); err != nil {
			t.Fatalf("no-config: %v", err)
		}
	})
	if strings.Contains(stderr, "notice:") {
		t.Fatalf("--no-config should not notice:\n%s", stderr)
	}
}

func TestRunExplain_ColorConflict(t *testing.T) {
	home := t.TempDir()
	t.Setenv(debugConfigHomeEnv, home)
	err := RunExplainWithRunner([]string{"--color", "--no-color", "hello"}, &mockRunner{})
	if err == nil || !strings.Contains(err.Error(), "cannot be specified together") {
		t.Fatalf("error = %v", err)
	}
}
