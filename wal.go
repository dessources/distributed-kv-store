package main

import (
	"os"
	"sync"
)

const (
	OpGet    byte = 0
	OpSet    byte = 1
	OpDelete byte = 2
)

type WAL struct {
	mu sync.Mutex
	f  *os.File
}

func (l *WAL) NewWal(path string) (*WAL, error) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		return nil, err
	}

	return &WAL{f: f}, nil
}
