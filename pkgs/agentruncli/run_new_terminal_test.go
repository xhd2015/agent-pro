package agentruncli

import (
	"slices"
	"strings"
	"testing"

	flags "github.com/xhd2015/less-flags"
)

func TestReconstructRunRemainder_stripsNewTerminalKeepsBinaryAndPrompt(t *testing.T) {
	var recorded flags.Flags
	var open, nt, fromPrompt bool
	var binary, configHome string
	var env []string
	remain, err := flags.Bool("--open", &open).
		Bool("--new-terminal", &nt).
		Bool("--session-id-from-prompt", &fromPrompt).
		String("--agent-runner-binary", &binary).
		String("--agent-runner-config-home", &configHome).
		StringSlice("-e,--env", &env).
		CollectParsedFlags(&recorded).
		Parse([]string{
			"--open", "--new-terminal", "--session-id-from-prompt",
			"--agent-runner-binary", "llm-mock-run-codex",
			"--agent-runner-config-home", "/tmp/codex-home",
			"--env", "LLM_MOCK_MCP=slow_01=1s-10s",
			"--", "wait then say done",
		})
	if err != nil {
		t.Fatal(err)
	}
	got := reconstructRunRemainder(recorded, remain, "")
	joined := strings.Join(got, "\x00")
	if slices.Contains(got, "--new-terminal") {
		t.Fatalf("child argv still has --new-terminal: %#v", got)
	}
	if got[0] != "run" {
		t.Fatalf("got[0]=%q", got[0])
	}
	if !slices.Contains(got, "--open") || !slices.Contains(got, "--session-id-from-prompt") {
		t.Fatalf("missing --open / --session-id-from-prompt: %#v", got)
	}
	if !slices.Contains(got, "llm-mock-run-codex") || !slices.Contains(got, "/tmp/codex-home") {
		t.Fatalf("missing binary/config-home: %#v", got)
	}
	if !strings.Contains(joined, "LLM_MOCK_MCP=slow_01=1s-10s") {
		t.Fatalf("missing MCP env: %#v", got)
	}
	if !slices.Contains(got, "--") || !slices.Contains(got, "wait then say done") {
		t.Fatalf("missing prompt after --: %#v", got)
	}
}

func TestReconstructRunRemainder_promptFileOmitsRemain(t *testing.T) {
	var recorded flags.Flags
	var nt bool
	var pf string
	remain, err := flags.Bool("--new-terminal", &nt).
		String("--prompt-file", &pf).
		CollectParsedFlags(&recorded).
		Parse([]string{"--new-terminal", "--prompt-file", "/tmp/p.txt"})
	if err != nil {
		t.Fatal(err)
	}
	got := reconstructRunRemainder(recorded, remain, pf)
	if slices.Contains(got, "--") {
		t.Fatalf("prompt-file child should not add -- remain: %#v", got)
	}
	if !slices.Contains(got, "/tmp/p.txt") {
		t.Fatalf("missing prompt-file: %#v", got)
	}
}

func TestWithResolvedRunnerPreservesRunnerAcrossNewTerminal(t *testing.T) {
	got := withResolvedRunner([]string{"run", "--open", "--", "hello"}, "codex-tty")
	want := []string{"run", "--agent-runner", "codex-tty", "--open", "--", "hello"}
	if !slices.Equal(got, want) {
		t.Fatalf("child argv = %#v, want %#v", got, want)
	}
}

func TestResolveExecInTerminalFlags(t *testing.T) {
	tests := []struct {
		name     string
		execFlag bool
		noFlag   bool
		envDef   bool
		want     bool
		wantErr  bool
	}{
		{"neither_flag_env_off", false, false, false, false, false},
		{"neither_flag_env_on", false, false, true, true, false},
		{"exec_flag_env_off", true, false, false, true, false},
		{"exec_flag_env_on", true, false, true, true, false},
		{"no_flag_env_off", false, true, false, false, false},
		{"no_flag_env_on_override", false, true, true, false, false},
		{"both_flags", true, true, false, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveExecInTerminalFlags(tt.execFlag, tt.noFlag, tt.envDef)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("want error, got nil and value %v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExecInTerminalEnvEnabled(t *testing.T) {
	t.Setenv("AGENT_RUN_EXEC_IN_TERMINAL", "")
	if execInTerminalEnvEnabled() {
		t.Fatal("empty env: want false (default off)")
	}
	for _, v := range []string{"1", "true", "yes", "on", "TRUE", " On "} {
		t.Run(v, func(t *testing.T) {
			t.Setenv("AGENT_RUN_EXEC_IN_TERMINAL", v)
			if !execInTerminalEnvEnabled() {
				t.Fatalf("env %q: want true", v)
			}
		})
	}
	for _, v := range []string{"0", "false", "no", "off", "anything", ""} {
		t.Run(v, func(t *testing.T) {
			t.Setenv("AGENT_RUN_EXEC_IN_TERMINAL", v)
			if execInTerminalEnvEnabled() {
				t.Fatalf("env %q: want false", v)
			}
		})
	}
}

func TestWithExecPrefix(t *testing.T) {
	tests := []struct {
		name string
		cmd  string
		exec bool
		want string
	}{
		{"exec_on_simple", "agent-run run -- hi", true, "exec agent-run run -- hi"},
		{"exec_off_no_prefix", "agent-run run -- hi", false, "agent-run run -- hi"},
		{"exec_on_empty_cmd", "   ", true, "   "},
		{"exec_off_empty_cmd", "", false, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := withExecPrefix(tt.cmd, tt.exec)
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}
