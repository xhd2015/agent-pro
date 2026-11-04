package web

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"syscall"
	"time"
)

// FileStore uses <Dir>/<key>/<FileName>. The caller chooses Dir; there is no
// implicit Marcus, agent-run, or global home. An empty FileName means session.json.
// Existing flat binding records remain readable, including non-DSH runner pins.
type FileStore struct {
	Dir      string
	FileName string
}

func (s FileStore) path(key string) (string, error) {
	if strings.TrimSpace(s.Dir) == "" {
		return "", fmt.Errorf("DSH state directory is required")
	}
	if key == "" || key == "." || key == ".." || strings.ContainsAny(key, "/\\\x00") {
		return "", fmt.Errorf("invalid DSH session key")
	}
	name := s.FileName
	if name == "" {
		name = "session.json"
	}
	if name == "." || name == ".." || strings.ContainsAny(name, "/\\\x00") {
		return "", fmt.Errorf("invalid DSH state filename")
	}
	return filepath.Abs(filepath.Join(s.Dir, key, name))
}

func (s FileStore) Read(ctx context.Context, key string) (State, bool, error) {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return State{}, false, err
		}
	}
	path, err := s.path(key)
	if err != nil {
		return State{}, false, err
	}
	return fileRecord{path: path}.Load()
}

// WithSession holds a filesystem lock across network operations AND intermediate
// saves. Distinct Service/FileStore objects and separate processes share the
// same lock; an OS-released lock is not mistaken for a completed checkpoint.
func (s FileStore) WithSession(ctx context.Context, key string, fn func(Record) error) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	path, err := s.path(key)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	lock, err := os.OpenFile(path+".lock", os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer lock.Close()
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		err = syscall.Flock(int(lock.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			break
		}
		if !errors.Is(err, syscall.EWOULDBLOCK) && !errors.Is(err, syscall.EAGAIN) {
			return err
		}
		timer := time.NewTimer(10 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
	defer syscall.Flock(int(lock.Fd()), syscall.LOCK_UN)
	if err := ctx.Err(); err != nil {
		return err
	}
	return fn(fileRecord{path: path})
}

type fileRecord struct{ path string }

func (r fileRecord) Load() (State, bool, error) {
	raw, err := os.ReadFile(r.path)
	if os.IsNotExist(err) {
		return State{}, false, nil
	}
	if err != nil {
		return State{}, false, err
	}
	var state State
	if err := json.Unmarshal(raw, &state); err != nil {
		return State{}, false, fmt.Errorf("read DSH binding: %w", err)
	}
	if state.Runner == "" {
		return State{}, false, fmt.Errorf("invalid empty runner binding")
	}
	return state, true, nil
}

func (r fileRecord) Save(state State) error {
	if state.Runner == "" {
		return fmt.Errorf("cannot save an empty runner binding")
	}
	// Retain unknown application metadata, but remove omitted known fields (e.g.
	// a pending prompt cleared after admission) rather than accidentally retaining it.
	values := map[string]json.RawMessage{}
	raw, err := os.ReadFile(r.path)
	if err == nil {
		if err := json.Unmarshal(raw, &values); err != nil {
			return fmt.Errorf("read DSH binding before save: %w", err)
		}
		if values == nil {
			return fmt.Errorf("invalid null runner binding")
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	typ := reflect.TypeFor[State]()
	for i := 0; i < typ.NumField(); i++ {
		delete(values, strings.Split(typ.Field(i).Tag.Get("json"), ",")[0])
	}
	encoded, err := json.Marshal(state)
	if err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &fields); err != nil {
		return err
	}
	for key, value := range fields {
		values[key] = value
	}
	data, err := json.MarshalIndent(values, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(r.path), ".dsh-binding-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(append(data, '\n')); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), r.path)
}
