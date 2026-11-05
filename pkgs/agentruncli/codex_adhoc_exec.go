package agentruncli

import (
	"errors"
	"os"

	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
)

// Final PTY spawn boundary. tty-watch and Terminal only ever see the credential
// reference in session metadata; the resolved token is not part of their argv,
// CommandEnv, registry, or env-spill records.
func runCodexAdhocExec(args []string) error {
	if len(args) < 4 || args[2] != "--" {
		return errors.New("invalid internal Codex connection invocation")
	}
	c, err := agentstorage.ReadCodexConnection(args[0], args[1])
	if err != nil {
		return err
	}
	// Validate the actual command against the binding before resolving a secret.
	verified, err := c.Apply(args[3:], true)
	if err != nil {
		return err
	}
	env, err := c.ChildEnv(os.Environ())
	if err != nil {
		return err
	}
	return replaceCodexProcess(verified, env)
}
