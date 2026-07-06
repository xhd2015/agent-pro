package main

import (
	"time"

	"github.com/hinshun/vt10x"
	"github.com/xhd2015/agent-pro/pkgs/ttywatch"
)

type RegistryEntry = ttywatch.RegistryEntry

func TTYWatchHome() (string, error) { return ttywatch.TTYWatchHome() }
func registryConfig(home string) ttywatch.RegistryConfig {
	return ttywatch.DefaultRegistryConfig(home)
}
func ReserveCustomSessionID(home, sessionID string) (func(), error) {
	return ttywatch.ReserveCustomSessionID(registryConfig(home), sessionID)
}
func ReserveRegistrySessionID(home string) (string, func(), error) {
	return ttywatch.ReserveRegistrySessionID(registryConfig(home))
}
func WriteRegistry(home string, entry RegistryEntry) error {
	return ttywatch.WriteRegistry(registryConfig(home), entry)
}
func ReadRegistry(home, sessionID string) (*RegistryEntry, error) {
	return ttywatch.ReadRegistry(registryConfig(home), sessionID)
}
func RemoveRegistry(home, sessionID string) { ttywatch.RemoveRegistry(registryConfig(home), sessionID) }
func RemoveRegistryIfMatch(home, sessionID, listenAddr string, pid int) {
	ttywatch.RemoveRegistryIfMatch(registryConfig(home), sessionID, listenAddr, pid)
}
func ListRegistryEntries(home string, prune bool) ([]RegistryEntry, error) {
	return ttywatch.ListRegistryEntries(registryConfig(home), prune)
}
func waitForRegistryEntry(home, sessionID string, timeout time.Duration) (*RegistryEntry, error) {
	return ttywatch.WaitForRegistryEntry(registryConfig(home), sessionID, timeout)
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