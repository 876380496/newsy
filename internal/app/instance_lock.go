package app

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"newsy/internal/logging"
)

type instanceLock struct {
	path string
	file *os.File
}

func acquireInstanceLock(path string) (*instanceLock, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve lock path: %w", err)
	}

	f, err := os.OpenFile(absPath, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open lock file: %w", err)
	}

	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		_ = f.Close()
		if err == syscall.EWOULDBLOCK {
			return nil, fmt.Errorf("newsy is already running; please close the existing instance and try again")
		}
		return nil, fmt.Errorf("lock file %s: %w", absPath, err)
	}

	logging.Infof("instance lock acquired path=%s", absPath)
	return &instanceLock{path: absPath, file: f}, nil
}

func (l *instanceLock) Close() error {
	if l == nil || l.file == nil {
		return nil
	}
	logging.Infof("instance lock releasing path=%s", l.path)
	unlockErr := syscall.Flock(int(l.file.Fd()), syscall.LOCK_UN)
	closeErr := l.file.Close()
	l.file = nil
	if unlockErr != nil {
		return unlockErr
	}
	return closeErr
}
