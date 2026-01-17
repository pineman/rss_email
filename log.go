package main

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type MonthlyLogger struct {
	dir         string
	mu          sync.Mutex
	file        *os.File
	currentName string
}

func NewMonthlyLogger(dir string) (*MonthlyLogger, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	l := &MonthlyLogger{dir: dir}
	if err := l.rotate(); err != nil {
		return nil, err
	}
	return l, nil
}

func (l *MonthlyLogger) Write(p []byte) (n int, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	expectedName := l.logFileName()
	if l.currentName != expectedName {
		if err := l.rotate(); err != nil {
			return 0, err
		}
	}

	return l.file.Write(p)
}

func (l *MonthlyLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

func (l *MonthlyLogger) logFileName() string {
	return "rss_email_" + time.Now().Format("2006-01") + ".log"
}

func (l *MonthlyLogger) rotate() error {
	if l.file != nil {
		oldName := l.currentName
		l.file.Close()
		go l.compressOldLog(oldName)
	}

	l.currentName = l.logFileName()
	path := filepath.Join(l.dir, l.currentName)

	var err error
	l.file, err = os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	return nil
}

func (l *MonthlyLogger) compressOldLog(name string) {
	path := filepath.Join(l.dir, name)
	gzPath := path + ".gz"

	if _, err := os.Stat(gzPath); err == nil {
		os.Remove(path)
		return
	}

	src, err := os.Open(path)
	if err != nil {
		return
	}
	defer src.Close()

	dst, err := os.Create(gzPath)
	if err != nil {
		return
	}
	defer dst.Close()

	gz := gzip.NewWriter(dst)
	defer gz.Close()

	if _, err := io.Copy(gz, src); err != nil {
		os.Remove(gzPath)
		return
	}

	gz.Close()
	dst.Close()
	src.Close()
	os.Remove(path)
}
