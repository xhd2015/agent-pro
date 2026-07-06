package sudosetup

import "testing"

// guardProductionDeps panics when a Manager runs under go test without injected
// FS and Runner. Production CLIs must use real osFS/execRunner; tests must use
// fakes (see pkgs/sudosetup/sudosetuptest).
func guardProductionDeps(m *Manager, dep string) {
	if m == nil {
		return
	}
	if !testing.Testing() {
		return
	}
	switch dep {
	case "FS":
		if m.FS != nil {
			return
		}
	case "Runner":
		if m.Runner != nil {
			return
		}
	default:
		return
	}
	panic("sudosetup: Manager." + dep + " must be injected in tests — use sudosetuptest harness or a fake FS/Runner (production deps would modify /etc/sudoers.d and run real sudo)")
}