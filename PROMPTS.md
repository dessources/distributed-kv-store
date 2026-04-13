# Distributed KV Store: Engineering & Performance Log

## Project Objective
Build a production-quality, high-performance distributed Key-Value store in Go, focusing on low latency, zero allocations, and hardware-level optimizations.

---

## Phase 1: In-Memory Storage Engine

### 1. Concurrency Strategy
*   **Decision:** Map Sharding (Lock Striping).
*   **Logic:** Instead of a single global lock, the store is split into `N` independent shards, each with its own `sync.RWMutex`.
*   **Performance:** Drastically reduces lock contention. Probability of contention decreases by `1/N`.
*   **Avoided:** `sync.Map` (due to interface{} boxing, GC pressure, and O(N) promotion costs) and Global RWMutex (due to write-lock saturation).

### 2. Hardware Optimization: Cache Line Padding
*   **Decision:** Padded Slice of Structs (128 bytes).
*   **Concept:** CPUs fetch memory in 64-byte "Cache Lines."
*   **Problem:** **False Sharing**. Multiple shards on one cache line cause CPU cores to invalidate each other's caches.
*   **Solution:** `_ [128 - ((unsafe.Sizeof(sync.RWMutex{}) + unsafe.Sizeof(map[string][]byte{})) % 128)]byte`.
*   **Result:** Physical isolation of mutexes in memory, enabling true hardware parallelism.

### 3. Shard Indexing & Scaling
*   **Decision:** Power-of-Two Sharding.
*   **Logic:** Use `nextPowerOfTwo(runtime.GOMAXPROCS(0) * 4)`.
*   **Optimization:** Replacing the expensive modulo (`%`) with bitwise AND (`&`). 
    *   *Correction:* `hash(key) & (uint64(len(shards)) - 1)`.

---

## Mistakes & Course Corrections

### Logical Errors
1.  **Existence Bug:** `Get` was returning `false` even if the key existed.
    *   *Lesson:* Correctness is the prerequisite for performance.
2.  **Bitwise Masking:** Attempted `hash & len` instead of `hash & (len - 1)`.
    *   *Result:* Would have caused `index out of bounds` panics.
    *   *Lesson:* Bitmasks for power-of-two indexing must always be `N-1`.

### Synchronization Errors
1.  **Lock Mismatch:** Called `Unlock()` after `RLock()`.
    *   *Result:* Immediate runtime panic.
    *   *Lesson:* Go's `sync` package is unforgiving. `RLock` must pair with `RUnlock`.

### Implementation Refinements
1.  **Pointer Arithmetic:** Switched to local shard pointers (`shard := &s.shards[idx]`) to avoid redundant slice indexing and improve readability.
2.  **Allocation-Free Hashing:** Implemented inline FNV-1a to avoid the heap allocations triggered by `hash/fnv`'s interface usage.

---

## Next Steps
*   [ ] Finalize `main.go` with all fixes.
*   [ ] Implement Parallel Benchmarks (`testing.PB`).
*   [ ] Verify 0 B/op allocation profile.
*   [ ] Select Persistence Strategy (WAL).
