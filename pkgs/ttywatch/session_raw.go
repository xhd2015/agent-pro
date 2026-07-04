package ttywatch

import (
	"fmt"
	"os"
	"reflect"
	"unsafe"

	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/ptywrap"
	"golang.org/x/sys/unix"
)

// ensureSessionPTYInjectMode disables canonical mode on the session PTY so
// follow-up inject delivers bytes immediately, while keeping signal generation
// and output post-processing intact for existing run/watch behavior.
func ensureSessionPTYInjectMode(mgr *ptywrap.Manager, sessionID string) error {
	ptmx, err := sessionPTYMaster(mgr, sessionID)
	if err != nil {
		return err
	}
	fd := int(ptmx.Fd())
	termios, err := unix.IoctlGetTermios(fd, ioctlGetTermios())
	if err != nil {
		return fmt.Errorf("get session pty termios: %w", err)
	}
	termios.Lflag &^= unix.ICANON
	termios.Cc[unix.VMIN] = 1
	termios.Cc[unix.VTIME] = 0
	if err := unix.IoctlSetTermios(fd, ioctlSetTermios(), termios); err != nil {
		return fmt.Errorf("set session pty inject mode: %w", err)
	}
	return nil
}

func sessionPTYMaster(mgr *ptywrap.Manager, sessionID string) (*os.File, error) {
	mv := reflect.ValueOf(mgr).Elem()
	sessionsField := mv.FieldByName("sessions")
	if !sessionsField.IsValid() {
		return nil, fmt.Errorf("ptywrap manager sessions field not found")
	}
	sessionVal := sessionsField.MapIndex(reflect.ValueOf(sessionID))
	if !sessionVal.IsValid() {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}
	ptmxField := sessionVal.Elem().FieldByName("ptmx")
	if !ptmxField.IsValid() {
		return nil, fmt.Errorf("session ptmx field not found")
	}
	ptmxPtr := reflect.NewAt(ptmxField.Type(), unsafe.Pointer(ptmxField.UnsafeAddr())).Elem()
	ptmx, ok := ptmxPtr.Interface().(*os.File)
	if !ok || ptmx == nil {
		return nil, fmt.Errorf("session ptmx unavailable")
	}
	return ptmx, nil
}