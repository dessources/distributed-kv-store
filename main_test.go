package main

import "testing"

func BenchmarkStore_Set(b *testing.B) {
	store := NewStore(64)
	key := "benchmark-key"
	val := []byte("benchmark-value")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			store.Set(key, val)
		}
	})
}
