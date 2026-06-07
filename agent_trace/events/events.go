package events

import (
	"os"
	"path/filepath"
)

type Logger interface {
	Append(line []byte) error
	Sync() error
	Close() error
}

func Open(path string) (Logger, error) {
	if path == "" {
		return nil, nil
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &fileLogger{f: f}, nil
}

type fileLogger struct {
	f *os.File
}

func (l *fileLogger) Append(line []byte) error {
	if l == nil || l.f == nil {
		return nil
	}
	_, err := l.f.Write(line)
	return err
}

func (l *fileLogger) Sync() error {
	if l == nil || l.f == nil {
		return nil
	}
	return l.f.Sync()
}

func (l *fileLogger) Close() error {
	if l == nil || l.f == nil {
		return nil
	}
	return l.f.Close()
}
