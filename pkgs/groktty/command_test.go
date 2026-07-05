package groktty

import (
	"os"
	"path/filepath"
	"testing"

	agentexec "github.com/xhd2015/agent-pro/agent/exec"
)

func TestBuildGrokCommandArgv_agentRunnerBinaryWithUserFlags(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "llm-mock-run-grok")
	if err := os.WriteFile(bin, []byte("#!/bin/sh\n"), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("AGENT_RUN_GROK_TTY_COMMAND", "")

	env := agentexec.NewEnv(&agentexec.PathsConfig{
		RootDirName: ".agent-pro",
		DataDirName: "data",
		BinDirName:  "bin",
	}, "AGENT_PRO_CONFIG_HOME")
	argv, err := BuildGrokCommandArgv(env, "", bin+" --model inner-model", "outer-model", "sess-1")
	if err != nil {
		t.Fatalf("BuildGrokCommandArgv: %v", err)
	}
	want := []string{
		bin,
		"--model", "inner-model",
		"--always-approve", "--permission-mode=bypassPermissions",
		"--resume", "sess-1",
	}
	if len(argv) != len(want) {
		t.Fatalf("argv = %#v, want %#v", argv, want)
	}
	for i := range want {
		if argv[i] != want[i] {
			t.Fatalf("argv[%d] = %q, want %q (full argv %#v)", i, argv[i], want[i], argv)
		}
	}
}

func TestBuildGrokCommandArgv_agentRunnerBinaryNameOnly(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "llm-mock-run-grok")
	if err := os.WriteFile(bin, []byte("#!/bin/sh\n"), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("AGENT_RUN_GROK_TTY_COMMAND", "")

	env := agentexec.NewEnv(&agentexec.PathsConfig{
		RootDirName: ".agent-pro",
		DataDirName: "data",
		BinDirName:  "bin",
	}, "AGENT_PRO_CONFIG_HOME")
	argv, err := BuildGrokCommandArgv(env, "", bin, "mock-model", "")
	if err != nil {
		t.Fatalf("BuildGrokCommandArgv: %v", err)
	}
	want := []string{
		bin,
		"--always-approve", "--permission-mode=bypassPermissions",
		"--model", "mock-model",
	}
	if len(argv) != len(want) {
		t.Fatalf("argv = %#v, want %#v", argv, want)
	}
	for i := range want {
		if argv[i] != want[i] {
			t.Fatalf("argv[%d] = %q, want %q (full argv %#v)", i, argv[i], want[i], argv)
		}
	}
}

func TestBuildCodexCommandArgv_innerModelWins(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "codex")
	if err := os.WriteFile(bin, []byte("#!/bin/sh\n"), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("AGENT_RUN_CODEX_TTY_COMMAND", "")

	env := agentexec.NewEnv(&agentexec.PathsConfig{
		RootDirName: ".agent-pro",
		DataDirName: "data",
		BinDirName:  "bin",
	}, "AGENT_PRO_CONFIG_HOME")
	argv, err := BuildCodexCommandArgv(env, "", bin+" --model inner", "outer", "")
	if err != nil {
		t.Fatalf("BuildCodexCommandArgv: %v", err)
	}
	want := []string{
		bin,
		"--model", "inner",
		"--dangerously-bypass-approvals-and-sandbox",
	}
	if len(argv) != len(want) {
		t.Fatalf("argv = %#v, want %#v", argv, want)
	}
	for i := range want {
		if argv[i] != want[i] {
			t.Fatalf("argv[%d] = %q, want %q (full argv %#v)", i, argv[i], want[i], argv)
		}
	}
}