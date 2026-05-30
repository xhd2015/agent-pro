package models

import (
	"strings"

	"github.com/xhd2015/agent-pro/agent/exec/tool_exec"
)

func ListFree() (models []string, selected string, err error) {
	cmd, err := tool_exec.New("opencode", []string{"models"}, nil)
	if err != nil {
		return nil, "", err
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, "", err
	}

	for _, model := range strings.Split(string(output), "\n") {
		model = strings.TrimSpace(model)
		if strings.Contains(model, "free") || strings.HasPrefix(model, "opencode/") && strings.Contains(model, "-free") {
			models = append(models, model)
		}
	}
	if len(models) > 0 {
		selected = models[0]
	}
	return models, selected, nil
}
