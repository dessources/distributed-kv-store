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
	sync.RWMutex
	m map[string][]byte

	_ [128 - ((unsafe.Sizeof(sync.RWMutex{}) + unsafe.Sizeof(map[string][]byte{})) % 128)]byte
}

type Store struct {
	shards []shard
}

func NewStore(shard_count int) *Store {

	shard_count = nextPowerOfTwo(shard_count)

	shards := make([]shard, shard_count)

	for i := 0; i < shard_count; i++ {
		shards[i].m = make(map[string][]byte)
	}

	return &Store{
		shards: shards,
	}
}

func nextPowerOfTwo(n int) int {
	if n <= 1 {
		return 1
	}

	return 1 << (bits.Len(uint(n - 1)))
}

func hash(key string) uint64 {
	hash := offset64

	for i := 0; i < len(key); i++ {
		hash = hash ^ uint64(key[i])
		hash = hash * prime64
	}

	return hash
}

func (s *Store) Get(key string) ([]byte, bool) {

	shardId := hash(key) & uint64(len(s.shards)-1)
	shard := &s.shards[shardId]

	shard.RLock()

	v, ok := shard.m[key]

	shard.RUnlock()

	return v, ok

}

func (s *Store) Set(key string, value []byte) {
	shardId := hash(key) & uint64(len(s.shards)-1)
	shard := &s.shards[shardId]

	shard.Lock()

	shard.m[key] = value

	shard.Unlock()
}

func (s *Store) Delete(key string) {
	shardId := hash(key) & uint64(len(s.shards)-1)
	shard := &s.shards[shardId]

	shard.Lock()

	delete(shard.m, key)

	shard.Unlock()
}
