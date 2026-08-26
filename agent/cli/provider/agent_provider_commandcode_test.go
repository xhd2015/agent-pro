package agentprovider

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/agent-pro/agent/cli/commandcode"
	"github.com/xhd2015/agent-pro/agent/cli/registry"
	agentexec "github.com/xhd2015/agent-pro/agent/exec"
)

func TestBuild_Commandcode(t *testing.T) {
	dir := t.TempDir()
	bin := filepath.Join(dir, "cmd")
	if err := os.WriteFile(bin, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv(registry.EnvCommandcodeCLIPath, bin)

	env := agentexec.NewEnv(&agentexec.PathsConfig{
		RootDirName: ".agent-pro",
		DataDirName: "data",
		BinDirName:  "bin",
	}, "AGENT_PRO_CONFIG_HOME")

	built, err := Build(registry.AgentRunnerCommandcode, "", ".", env)
	if err != nil {
		t.Fatalf("Build commandcode: %v", err)
	}
	if built.ID != registry.AgentRunnerCommandcode {
		t.Fatalf("ID = %q, want commandcode", built.ID)
	}
	agent, ok := built.Agent.(*commandcode.CommandcodeAgent)
	if !ok {
		t.Fatalf("Agent type = %T, want *commandcode.CommandcodeAgent", built.Agent)
	}
	if agent.AgentPath != bin {
		t.Fatalf("AgentPath = %q, want %q", agent.AgentPath, bin)
	}
}
