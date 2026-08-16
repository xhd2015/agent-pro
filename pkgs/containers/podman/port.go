package podman

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// ErrNoProcessOnPort is returned when no process is found listening on a port.
var ErrNoProcessOnPort = errors.New("no process found on port")

// CheckPort returns true if the port is in use (someone is listening on it).
func CheckPort(port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", port), 1*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// GetPidOnPort returns the PID of the process listening on the given TCP port.
// Returns ErrNoProcessOnPort if no process is found.
func GetPidOnPort(port int) (int, error) {
	portStr := fmt.Sprintf("%d", port)

	switch runtime.GOOS {
	case "darwin", "linux":
		if pid, err := pidFromLsof(portStr); err == nil {
			return pid, nil
		}
		if runtime.GOOS == "linux" {
			if pid, err := pidFromSS(portStr); err == nil {
				return pid, nil
			}
			if pid, err := pidOnPortFromProc(port); err == nil {
				return pid, nil
			}
		}
		return 0, fmt.Errorf("%w: %d", ErrNoProcessOnPort, port)
	default:
		return 0, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

func pidFromLsof(portStr string) (int, error) {
	cmd := exec.Command("lsof", "-ti", fmt.Sprintf("tcp:%s", portStr), "-sTCP:LISTEN")
	out, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	pidStr := strings.TrimSpace(string(out))
	if idx := strings.IndexByte(pidStr, '\n'); idx > 0 {
		pidStr = pidStr[:idx]
	}
	if pidStr == "" {
		return 0, ErrNoProcessOnPort
	}
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return 0, fmt.Errorf("failed to parse PID '%s': %w", pidStr, err)
	}
	return pid, nil
}

func pidFromSS(portStr string) (int, error) {
	cmd := exec.Command("ss", "-tlnp", fmt.Sprintf("sport = :%s", portStr))
	out, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	for _, line := range strings.Split(string(out), "\n") {
		idx := strings.Index(line, "pid=")
		if idx < 0 {
			continue
		}
		rest := line[idx+4:]
		end := strings.IndexAny(rest, ",) \t\n")
		if end <= 0 {
			continue
		}
		pid, err := strconv.Atoi(rest[:end])
		if err != nil {
			return 0, fmt.Errorf("failed to parse PID '%s': %w", rest[:end], err)
		}
		return pid, nil
	}
	return 0, ErrNoProcessOnPort
}

const tcpListenState = "0A"

func pidOnPortFromProc(port int) (int, error) {
	for _, path := range []string{"/proc/net/tcp", "/proc/net/tcp6"} {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		inode, ok := listenInodeForPort(string(data), port)
		if !ok {
			continue
		}
		if pid, err := pidForSocketInode(inode); err == nil {
			return pid, nil
		}
	}
	return 0, ErrNoProcessOnPort
}

func listenInodeForPort(data string, port int) (string, bool) {
	want := fmt.Sprintf("%04X", port)
	for _, line := range strings.Split(data, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 10 {
			continue
		}
		local, st, inode := fields[1], fields[3], fields[9]
		colon := strings.LastIndexByte(local, ':')
		if colon < 0 || !strings.EqualFold(local[colon+1:], want) {
			continue
		}
		if !strings.EqualFold(st, tcpListenState) {
			continue
		}
		if inode == "" || inode == "0" {
			continue
		}
		return inode, true
	}
	return "", false
}

func pidForSocketInode(inode string) (int, error) {
	want := "socket:[" + inode + "]"
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return 0, err
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(e.Name())
		if err != nil {
			continue
		}
		fds, err := os.ReadDir("/proc/" + e.Name() + "/fd")
		if err != nil {
			continue
		}
		for _, fd := range fds {
			target, err := os.Readlink("/proc/" + e.Name() + "/fd/" + fd.Name())
			if err != nil {
				continue
			}
			if target == want {
				return pid, nil
			}
		}
	}
	return 0, ErrNoProcessOnPort
}

// KillPortPid kills any process listening on the given port and waits for it
// to be free. Returns the killed PID, or 0 if no process was found.
func KillPortPid(port int) (int, error) {
	pid, err := GetPidOnPort(port)
	if err != nil {
		if errors.Is(err, ErrNoProcessOnPort) {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to find process on port %d: %w", port, err)
	}

	if err := syscall.Kill(pid, syscall.SIGKILL); err != nil {
		return pid, fmt.Errorf("failed to kill process %d on port %d: %w", pid, port, err)
	}

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if !CheckPort(port) {
			time.Sleep(200 * time.Millisecond)
			return pid, nil
		}
		time.Sleep(100 * time.Millisecond)
	}

	return pid, fmt.Errorf("port %d is still in use after killing process %d", port, pid)
}
