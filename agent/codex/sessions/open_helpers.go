package sessions

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/xhd2015/agent-pro/pkgs/shell"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

func resolveCodexBin(codexBin string, lookPath func(file string) (string, error)) (string, error) {
	if strings.TrimSpace(codexBin) != "" {
		return codexBin, nil
	}
	if lookPath == nil {
		lookPath = exec.LookPath
	}
	bin, err := lookPath("codex")
	if err != nil {
		return "", fmt.Errorf("codex not found on PATH: %w", err)
	}
	return bin, nil
}

func quotedForkCommandLine(bin string, argv []string) string {
	parts := append([]string{bin}, argv...)
	quoted := make([]string, 0, len(parts))
	for _, p := range parts {
		quoted = append(quoted, shell.ShellQuote(p))
	}
	return strings.Join(quoted, " ")
}

func resumeCodexArgv(sessionID string) []string {
	return []string{"resume", sessionID}
}

func listLiveITermSessions() ([]iterm2.SessionRef, error) {
	script := iterm2.BuildSessionListScript()
	cmd := exec.Command("osascript", "-e", script)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("list iTerm sessions: %w", err)
	}
	return iterm2.ParseSessionListOutput(string(out))
}

func defaultOpenInNewWindow(dir, followUp string) error {
	return iterm2.OpenConfig(dir, &iterm2.Config{
		Mode:             iterm2.ModeForceNew,
		FollowUpCommands: []string{followUp},
		SafeInputIgnore:  true,
	})
}
