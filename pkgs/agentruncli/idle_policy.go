package agentruncli

import "github.com/xhd2015/agent-pro/pkgs/agentstorage"

// IdlePolicy is the session-dir idle-exit file (not meta.json).
type IdlePolicy = agentstorage.IdlePolicy

// IdlePolicyPath is $home/sessions/<sessionID>/idle-policy.json.
func IdlePolicyPath(home, sessionID string) string {
	return agentstorage.IdlePolicyPath(home, sessionID)
}

// WriteIdlePolicy writes compact JSON. Zero timeout + ExitOnIdle → 10m.
func WriteIdlePolicy(home, sessionID string, p IdlePolicy) error {
	return agentstorage.WriteIdlePolicy(home, sessionID, p)
}

// ReadIdlePolicy loads idle-policy.json. Missing file → found=false, err=nil.
func ReadIdlePolicy(home, sessionID string) (p IdlePolicy, found bool, err error) {
	return agentstorage.ReadIdlePolicy(home, sessionID)
}
