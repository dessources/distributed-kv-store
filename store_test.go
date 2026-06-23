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

func BenchmarkStore_Delete(b *testing.B) {
	store := NewStore(64)
	key := "benchmark-key"
	val := []byte("benchmark-value")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			store.Set(key, val)
			store.Delete(key)
		}
	})
}

func BenchmarkStore_Get(b *testing.B) {
	store := NewStore(64)
	key := "benchmark-key"
	val := []byte("benchmark-value")

	//pre populate the store to ensure we benchmark the read, not the write.
	store.Set(key, val)
	//wipe the initialization overhead from metrics
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = store.Get(key)
		}
	})
}
