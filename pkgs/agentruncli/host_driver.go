package agentruncli

import (
	"strings"
	"sync"

	"github.com/xhd2015/agent-pro/pkgs/agentdriver"
)

// processHostDriver is an optional embedding default (e.g. spl agent-run sets
// Binary=abs(spl), Args=["agent-run"]). Explicit Driver on Opts always wins.
var (
	processHostMu     sync.RWMutex
	processHostDriver agentdriver.Driver
	processHostSet    bool
)

// SetProcessHostDriver sets the default host Driver for this process when Opts
// leave Driver zero. Used by embedding hosts (spl agent-run); bare agent-run
// leaves this unset so serve falls back to agentdriver.DefaultSelf.
func SetProcessHostDriver(d agentdriver.Driver) {
	processHostMu.Lock()
	defer processHostMu.Unlock()
	processHostDriver = d
	processHostSet = true
}

// ClearProcessHostDriver clears the process default (tests).
func ClearProcessHostDriver() {
	processHostMu.Lock()
	defer processHostMu.Unlock()
	processHostDriver = agentdriver.Driver{}
	processHostSet = false
}

// mergeHostDriver returns explicit if non-zero Binary or Args; else process default.
func mergeHostDriver(explicit agentdriver.Driver) agentdriver.Driver {
	if strings.TrimSpace(explicit.Binary) != "" || len(explicit.Args) > 0 {
		return explicit
	}
	processHostMu.RLock()
	defer processHostMu.RUnlock()
	if processHostSet {
		return processHostDriver
	}
	return explicit
}
