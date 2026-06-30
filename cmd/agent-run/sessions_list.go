package main

import (
	"os"
	"path/filepath"

	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
)

func listAllSessions(store agentstorage.Store) ([]agentstorage.SessionMeta, error) {
	root := filepath.Join(store.Home(), "sessions")
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []agentstorage.SessionMeta
	for _, runnerEnt := range entries {
		if !runnerEnt.IsDir() {
			continue
		}
		runner := runnerEnt.Name()
		list, err := store.ListSessions(runner)
		if err != nil {
			continue
		}
		out = append(out, list...)
	}
	return out, nil
}