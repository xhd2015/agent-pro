package agentui

import (
	"fmt"

	"github.com/xhd2015/skills/install"
)

func runSkillCommand(cfg Config, args []string) error {
	if cfg.SkillName == "" || cfg.SkillContent == "" {
		return fmt.Errorf("skill command is not configured for %s", cfg.AgentName)
	}
	if len(args) == 0 {
		return fmt.Errorf("expected skill show or skill install")
	}
	sub := args[0]
	switch sub {
	case "show":
		if len(args) > 1 {
			return fmt.Errorf("unexpected arguments after show")
		}
		fmt.Print(cfg.SkillContent)
		return nil
	case "install":
		return install.HandleInstall(install.InstallOptions{
			SkillDirName: cfg.SkillName,
			SkillContent: cfg.SkillContent,
			Usage:        cfg.AgentName + " skill install",
		}, args[1:])
	default:
		return fmt.Errorf("unknown skill sub-command: %s, expected skill show or skill install", sub)
	}
}
