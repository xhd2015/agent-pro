package daemon

import (
	"fmt"
	"hash/fnv"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/xhd2015/agent-pro/agents/agent-hub/model"
	"github.com/xhd2015/agent-pro/agents/agent-hub/storage"
)

type Daemon struct {
	home     string
	store    *storage.Store
	mu       sync.Mutex
	lockFile *os.File
	listener net.Listener
	running  bool
}

type Status struct {
	Running    bool   `json:"running"`
	Home       string `json:"home"`
	SocketPath string `json:"socket_path"`
}

func New(home string) *Daemon {
	return &Daemon{
		home:  home,
		store: storage.New(home),
	}
}

func (d *Daemon) Start() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.running {
		return nil
	}
	if err := os.MkdirAll(d.home, 0755); err != nil {
		return err
	}
	lockPath := filepath.Join(d.home, "daemon.lock")
	lock, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("acquire daemon lock: %w", err)
	}
	socketPath := d.SocketPath()
	_ = os.Remove(socketPath)
	ln, err := net.Listen("unix", socketPath)
	if err != nil {
		_ = lock.Close()
		_ = os.Remove(lockPath)
		return err
	}
	if err := d.store.RebuildIndexes(); err != nil {
		_ = ln.Close()
		_ = lock.Close()
		_ = os.Remove(socketPath)
		_ = os.Remove(lockPath)
		return err
	}
	if err := d.store.RebuildSessions(); err != nil {
		_ = ln.Close()
		_ = lock.Close()
		_ = os.Remove(socketPath)
		_ = os.Remove(lockPath)
		return err
	}
	d.lockFile = lock
	d.listener = ln
	d.running = true
	return nil
}

func (d *Daemon) Stop() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if !d.running {
		return nil
	}
	var firstErr error
	if d.listener != nil {
		if err := d.listener.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	if d.lockFile != nil {
		if err := d.lockFile.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	if err := os.Remove(d.SocketPath()); err != nil && !os.IsNotExist(err) && firstErr == nil {
		firstErr = err
	}
	if err := os.Remove(filepath.Join(d.home, "daemon.lock")); err != nil && !os.IsNotExist(err) && firstErr == nil {
		firstErr = err
	}
	d.running = false
	d.listener = nil
	d.lockFile = nil
	return firstErr
}

func (d *Daemon) SocketPath() string {
	h := fnv.New32a()
	_, _ = h.Write([]byte(d.home))
	return filepath.Join(os.TempDir(), fmt.Sprintf("agent-hub-%08x.sock", h.Sum32()))
}

func (d *Daemon) Notify(event model.NormalizedEvent) (model.Envelope, error) {
	if err := event.Validate(); err != nil {
		return model.Envelope{}, err
	}
	return d.store.Append(event, time.Now().UTC())
}

func (d *Daemon) Fetch(consumerID string, limit int, peek bool) (model.FetchResponse, error) {
	if limit == 0 {
		limit = 1
	}
	return d.store.Fetch(consumerID, limit, peek)
}

func (d *Daemon) Status() (Status, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return Status{
		Running:    d.running,
		Home:       d.home,
		SocketPath: d.SocketPath(),
	}, nil
}

func (d *Daemon) Store() *storage.Store {
	return d.store
}
