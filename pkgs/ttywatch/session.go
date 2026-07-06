package ttywatch

import (
	"context"
	"fmt"
	"os"
	"syscall"
	"time"

	ptyclient "github.com/xhd2015/dot-pkgs/go-pkgs/shell/ptywrap/client"
)

func (s *EphemeralSession) registryConfig() RegistryConfig {
	return registryConfigFor(s.Home, s.RegistrySubdir)
}

// EphemeralSession manages a short-lived tty-watch session without CLI subprocesses.
type EphemeralSession struct {
	Home           string
	RegistrySubdir string
	SessionID      string
	Command        []string
	ExtraPaths     []string

	entry       *RegistryEntry
	serveCancel context.CancelFunc
	serveDone   chan error
	inProcess   bool
}

// NewEphemeralSession returns a session scoped to home, id, and command argv.
func NewEphemeralSession(home, sessionID string, command []string) *EphemeralSession {
	return &EphemeralSession{
		Home:      home,
		SessionID: sessionID,
		Command:   append([]string(nil), command...),
	}
}

// StartInProcess launches ServeSession in the current process.
func (s *EphemeralSession) StartInProcess(ctx context.Context) error {
	if s == nil {
		return fmt.Errorf("ephemeral session is nil")
	}
	serveCtx, cancel := context.WithCancel(ctx)
	s.serveCancel = cancel
	s.inProcess = true
	s.serveDone = make(chan error, 1)
	go func() {
		s.serveDone <- ServeSession(serveCtx, ServeOptions{
			SessionID:      s.SessionID,
			Command:        s.Command,
			Home:           s.Home,
			RegistrySubdir: s.RegistrySubdir,
			ExtraPaths:     s.ExtraPaths,
		})
	}()

	entry, err := WaitForRegistryEntry(s.registryConfig(), s.SessionID, 15*time.Second)
	if err != nil {
		cancel()
		<-s.serveDone
		return err
	}
	s.entry = entry
	return nil
}

// StartDetached re-executes binaryPath with a slug-based serve token.
func (s *EphemeralSession) StartDetached(ctx context.Context, binaryPath string) error {
	if s == nil {
		return fmt.Errorf("ephemeral session is nil")
	}
	result, err := HeadlessRun(ctx, HeadlessRunOptions{
		Home:           s.Home,
		RegistrySubdir: s.RegistrySubdir,
		SessionID:      s.SessionID,
		Command:        s.Command,
		BinaryPath:     binaryPath,
		ExtraPaths:     s.ExtraPaths,
		KeepAlive:      true,
	})
	if err != nil {
		return err
	}
	s.entry = result.Entry
	return nil
}

// Send injects follow-up text into the live session.
func (s *EphemeralSession) Send(message string) error {
	entry, err := s.loadEntry()
	if err != nil {
		return err
	}
	if !TCPReachable(entry.ListenAddr) {
		RemoveRegistryIfMatch(s.registryConfig(), s.SessionID, entry.ListenAddr, entry.PID)
		return fmt.Errorf("tty-watch session %s not found", s.SessionID)
	}
	if err := PrepareSessionInjectMode(entry.ListenAddr, s.SessionID); err != nil {
		return err
	}
	return InjectInput(entry.ListenAddr, s.SessionID, []byte(message))
}

// Snapshot returns rendered printable snapshot text.
func (s *EphemeralSession) Snapshot() (string, error) {
	entry, err := s.loadEntry()
	if err != nil {
		return "", err
	}
	if !TCPReachable(entry.ListenAddr) {
		RemoveRegistryIfMatch(s.registryConfig(), s.SessionID, entry.ListenAddr, entry.PID)
		return "", fmt.Errorf("tty-watch session %s not found", s.SessionID)
	}
	return SnapshotText(entry.ListenAddr, s.SessionID)
}

// Kill terminates the session and removes its registry entry.
func (s *EphemeralSession) Kill() error {
	entry, err := s.loadEntry()
	if err != nil {
		return nil
	}
	if !TCPReachable(entry.ListenAddr) {
		RemoveRegistryIfMatch(s.registryConfig(), s.SessionID, entry.ListenAddr, entry.PID)
		s.entry = nil
		return nil
	}

	c := ptyclient.NewClient("http://" + entry.ListenAddr)
	_ = c.Delete(s.SessionID)

	if s.inProcess && s.serveCancel != nil {
		s.serveCancel()
		if s.serveDone != nil {
			select {
			case <-s.serveDone:
			case <-time.After(3 * time.Second):
			}
		}
	} else if entry.PID > 0 && entry.PID != os.Getpid() && processAlive(entry.PID) {
		_ = syscall.Kill(entry.PID, syscall.SIGTERM)
		deadline := time.Now().Add(2 * time.Second)
		for time.Now().Before(deadline) && processAlive(entry.PID) {
			time.Sleep(100 * time.Millisecond)
		}
		if processAlive(entry.PID) {
			_ = syscall.Kill(entry.PID, syscall.SIGKILL)
		}
	}

	RemoveRegistryIfMatch(s.registryConfig(), s.SessionID, entry.ListenAddr, entry.PID)
	s.entry = nil
	return nil
}

// WaitWritable polls check readiness against live scrollback.
func (s *EphemeralSession) WaitWritable(check CheckWritableFunc, timeout time.Duration) WritableStatus {
	entry, err := s.loadEntry()
	if err != nil {
		return WritableStatus{Reason: err.Error()}
	}
	return WaitUntilWritable(check, entry.ListenAddr, s.SessionID, timeout)
}

// Entry returns the current registry entry when available.
func (s *EphemeralSession) Entry() *RegistryEntry {
	if s == nil || s.entry == nil {
		return nil
	}
	copy := *s.entry
	return &copy
}

func (s *EphemeralSession) loadEntry() (*RegistryEntry, error) {
	if s.entry != nil && s.entry.ListenAddr != "" {
		return s.entry, nil
	}
	entry, err := ReadRegistry(s.registryConfig(), s.SessionID)
	if err != nil {
		return nil, err
	}
	s.entry = entry
	return entry, nil
}