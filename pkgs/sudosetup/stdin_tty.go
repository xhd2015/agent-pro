package sudosetup

import (
	"fmt"
	"os"
)

const interactiveSetupHint = "run this command once from an interactive terminal and enter your password when prompted; after that, non-interactive use will work"

// StdinIsTerminal reports whether stdin is an interactive terminal.
func StdinIsTerminal() bool {
	fi, err := os.Stdin.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}

func (m *Manager) stdinIsTerminal() bool {
	if m.StdinIsTerminal != nil {
		return m.StdinIsTerminal()
	}
	return StdinIsTerminal()
}

func errInteractiveTerminalRequired(action string) error {
	return fmt.Errorf("%s requires an interactive terminal (stdin must be a TTY); %s", action, interactiveSetupHint)
}