package agentruncli

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/agent-pro/agent/cli/registry"
	"github.com/xhd2015/agent-pro/pkgs/agentui"
)

type webRunConfig struct {
	GrokHome            string
	GrokTTYRunnerBinary string
	DefaultRunner       string
}

func expandHomePath(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", nil
	}
	if path == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return home, nil
	}
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, strings.TrimPrefix(path, "~/")), nil
	}
	return path, nil
}

func applyWebGrokRunOptions(runner string, cfg webRunConfig, opts *agentui.RunOptions) {
	if strings.TrimSpace(cfg.GrokHome) != "" {
		switch registry.AgentRunnerID(runner) {
		case registry.AgentRunnerGrokTTY, registry.AgentRunnerGrok:
			opts.AgentRunnerConfigHome = cfg.GrokHome
		}
	}
	if registry.AgentRunnerID(runner) == registry.AgentRunnerGrokTTY && strings.TrimSpace(cfg.GrokTTYRunnerBinary) != "" {
		opts.AgentRunnerBinary = cfg.GrokTTYRunnerBinary
	}
}