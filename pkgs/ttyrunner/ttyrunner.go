// Package ttyrunner is a compatibility shim delegating to pkgs/agenttty.
// Sealed doctest trees still import this package path until fully migrated.
package ttyrunner

import (
	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
	"github.com/xhd2015/agent-pro/pkgs/agenttty"
	"github.com/xhd2015/tty-watch/pkgs/ttywatch"
)

type (
	Provider       = agenttty.Provider
	TTYSession     = agenttty.TTYSession
	WritableStatus = agenttty.WritableStatus
	RegistryEntry  = ttywatch.RegistryEntry
)

func IDs() []string                         { return agenttty.IDs() }
func Get(id string) (Provider, bool)      { return agenttty.Get(id) }
func IsTTYRunner(id string) bool            { return agenttty.IsTTYRunner(id) }
func LookupSession(home, terminalSessionID string) (*RegistryEntry, string, error) {
	return agenttty.LookupSession(home, terminalSessionID)
}
func ResolveByTerminalID(home, terminalSessionID string) (*TTYSession, error) {
	return agenttty.ResolveByTerminalID(home, terminalSessionID)
}
func ResolveByAgentSession(store agentstorage.Store, runner, agentSessionID string) (*TTYSession, error) {
	return agenttty.ResolveByAgentSession(store, runner, agentSessionID)
}