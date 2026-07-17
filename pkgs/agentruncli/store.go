package agentruncli

import (
	"os"

	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
)

func openStore() (agentstorage.Store, error) {
	home := os.Getenv("AGENT_RUN_HOME")
	if home == "" {
		var err error
		home, err = defaultHome()
		if err != nil {
			return nil, err
		}
	}
	return agentstorage.NewFileStore(home)
}

func defaultHome() (string, error) {
	dir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return dir + "/.agent-run", nil
}