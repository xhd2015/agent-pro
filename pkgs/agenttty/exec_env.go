package agenttty

import agentexec "github.com/xhd2015/agent-pro/agent/exec"

func newExecEnv() *agentexec.Env {
	return agentexec.NewEnv(&agentexec.PathsConfig{
		RootDirName: ".agent-pro",
		DataDirName: "data",
		BinDirName:  "bin",
	}, "AGENT_PRO_CONFIG_HOME")
}