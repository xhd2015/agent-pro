package main

import (
	"time"

	"github.com/hinshun/vt10x"
	"github.com/xhd2015/agent-pro/pkgs/ttywatch"
)

type RegistryEntry = ttywatch.RegistryEntry

func TTYWatchHome() (string, error) { return ttywatch.TTYWatchHome() }
func ReserveCustomSessionID(home, sessionID string) (func(), error) {
	return ttywatch.ReserveCustomSessionID(home, sessionID)
}
func ReserveRegistrySessionID(home string) (string, func(), error) {
	return ttywatch.ReserveRegistrySessionID(home)
}
func WriteRegistry(home string, entry RegistryEntry) error { return ttywatch.WriteRegistry(home, entry) }
func ReadRegistry(home, sessionID string) (*RegistryEntry, error) {
	return ttywatch.ReadRegistry(home, sessionID)
}
func RemoveRegistry(home, sessionID string) { ttywatch.RemoveRegistry(home, sessionID) }
func RemoveRegistryIfMatch(home, sessionID, listenAddr string, pid int) {
	ttywatch.RemoveRegistryIfMatch(home, sessionID, listenAddr, pid)
}
func ListRegistryEntries(home string, prune bool) ([]RegistryEntry, error) {
	return ttywatch.ListRegistryEntries(home, prune)
}
func waitForRegistryEntry(home, sessionID string, timeout time.Duration) (*RegistryEntry, error) {
	return ttywatch.WaitForRegistryEntry(home, sessionID, timeout)
}
func tcpReachable(addr string) bool { return ttywatch.TCPReachable(addr) }
func processAlive(pid int) bool     { return ttywatch.ProcessAlive(pid) }

func SanitizeForPrint(data string) string { return ttywatch.SanitizeForPrint(data) }
func renderSnapshotOutput(frame, scrollback string, cols, rows int) string {
	return ttywatch.RenderSnapshotOutput(frame, scrollback, cols, rows)
}
func renderSnapshotScrollback(raw string, cols, rows int) string {
	return ttywatch.RenderSnapshotScrollback(raw, cols, rows)
}
func readSnapshot(listenAddr, sessionID string) (frame, scrollback string, cols, rows int, err error) {
	return ttywatch.ReadSnapshot(listenAddr, sessionID)
}
func renderObserverFrame(data []byte, cols, rows int) []byte {
	return ttywatch.RenderObserverFrame(data, cols, rows)
}
func prepareSessionInjectMode(listenAddr, sessionID string) error {
	return ttywatch.PrepareSessionInjectMode(listenAddr, sessionID)
}

func isScreenSnapshotFrame(data []byte) bool { return ttywatch.IsScreenSnapshotFrame(data) }
func screenSnapshotToText(data []byte, cols, rows int) ([]byte, bool) {
	return ttywatch.ScreenSnapshotToText(data, cols, rows)
}
func renderVTStateToText(vt vt10x.Terminal, cols, rows int) ([]byte, bool) {
	return ttywatch.RenderVTStateToText(vt, cols, rows)
}