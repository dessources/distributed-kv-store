// Store.go defines the main storage data structure for our kv store
package main

import (
	"math/bits"
	"sync"
	"unsafe"
)

const (
	offset64 uint64 = 14695981039346656037
	prime64  uint64 = 1099511628211
)

type shard struct {
	m    sync.RWMutex
	data map[string][]byte
	_    [128 - ((unsafe.Sizeof(sync.RWMutex{}) + unsafe.Sizeof(map[string][]byte{})) % 128)]byte
}

type shards struct {
	arr []shard
}

func nextPowerofTwo(n int) int {
	if n <= 1 {
		return 1
	}

	return 1 << (bits.Len(uint(n - 1)))
}

func hash(key string) uint64 {
	h := offset64
	for i := 0; i < len(key); i++ {
		h = h ^ uint64(key[i])
		h = h * prime64
	}

	return h
}
