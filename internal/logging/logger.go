package logging

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
)

var (
	mu     sync.Mutex
	logger *log.Logger
	file   *os.File
)

func Init(path string) error {
	mu.Lock()
	defer mu.Unlock()

	if logger != nil {
		return nil
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(absPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}

	file = f
	logger = log.New(file, "", log.LstdFlags|log.Lmicroseconds)
	logger.Printf("[INFO] logger initialized path=%s", absPath)
	return nil
}

func Close() error {
	mu.Lock()
	defer mu.Unlock()

	if logger != nil {
		logger.Printf("[INFO] logger closing")
	}
	if file == nil {
		return nil
	}
	err := file.Close()
	file = nil
	logger = nil
	return err
}

func Infof(format string, args ...any) {
	logf("INFO", format, args...)
}

func Errorf(format string, args ...any) {
	logf("ERROR", format, args...)
}

func logf(level, format string, args ...any) {
	mu.Lock()
	defer mu.Unlock()
	if logger == nil {
		return
	}
	logger.Printf("[%s] %s", level, fmt.Sprintf(format, args...))
}
