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

type Store struct {
	shards []shard
	wal    *WAL
}

func NewStore(shardCount int, w *WAL) (*Store, error) {
	n := nextPowerofTwo(shardCount)
	shards := make([]shard, n)

	for i := range n {
		shards[i].data = make(map[string][]byte)
	}

	return &Store{
		shards: shards,
		wal:    w,
	}, nil
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

func (s *Store) getShardId(key string) uint64 {
	return hash(key) & uint64(len(s.shards)-1)
}

func (s *Store) Get(key string) ([]byte, bool) {
	shardId := s.getShardId(key)
	shard := &s.shards[shardId]

	shard.m.RUnlock()

	val, ok := shard.data[key]

	shard.m.RUnlock()

	return val, ok

}

func (s *Store) Set(key string, val []byte) {
	shard := &s.shards[s.getShardId(key)]

	shard.m.Lock()

	shard.data[key] = val

	shard.m.Unlock()

}
func (s *Store) Delete(key string) {
	shard := &s.shards[s.getShardId(key)]

	shard.m.Lock()

	delete(shard.data, key)

	shard.m.Unlock()

}
