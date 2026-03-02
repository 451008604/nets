package nets

import (
	"fmt"
	"hash/fnv"
	"sync"
)

const shardCount = 32

// Shard 包含一个读写锁和内部 map。
type shard[K comparable, V any] struct {
	sync.RWMutex
	m map[K]V
}

// ShardedMap 为一个拥有 32 个分片的并发安全 map。
// 每个分片拥有独立的读写锁，以降低竞争并实现高并发读写。
// 键通过 FNV-1a 哈希均匀分布到各分片。
type ShardedMap[K comparable, V any] struct {
	shards [shardCount]*shard[K, V]
}

// NewShardedMap 创建一个空的 ShardedMap 实例。
func NewShardedMap[K comparable, V any]() *ShardedMap[K, V] {
	sm := &ShardedMap[K, V]{}
	for i := range shardCount {
		sm.shards[i] = &shard[K, V]{
			m: make(map[K]V),
		}
	}
	return sm
}

// fnvHash 计算键的 FNV-1a 32 位哈希值。
func fnvHash[K comparable](key K) uint32 {
	h := fnv.New32a()
	// 采用 fmt.Sprintf 将任意可比较类型转为字符串，再写入哈希。
	// 对于 string、int 等基本类型效果良好。
	_, _ = fmt.Fprintf(h, "%v", key)
	return h.Sum32()
}

// getShard 返回键所在的分片索引（0~31）。
func (sm *ShardedMap[K, V]) getShard(key K) *shard[K, V] {
	idx := fnvHash(key) % shardCount
	return sm.shards[idx]
}

// Set 将键值对写入对应分片，使用写锁保证互斥。
func (sm *ShardedMap[K, V]) Set(key K, value V) {
	s := sm.getShard(key)
	s.Lock()
	s.m[key] = value
	s.Unlock()
}

// Get 从对应分片读取键值，使用读锁以支持并发读取。
func (sm *ShardedMap[K, V]) Get(key K) (V, bool) {
	s := sm.getShard(key)
	s.RLock()
	e, ok := s.m[key]
	s.RUnlock()
	return e, ok
}

// Delete 删除对应分片中的键，使用写锁。
func (sm *ShardedMap[K, V]) Delete(key K) {
	s := sm.getShard(key)
	s.Lock()
	delete(s.m, key)
	s.Unlock()
}

// Len 返回整个 ShardedMap 中所有键的总数。
func (sm *ShardedMap[K, V]) Len() int {
	total := 0
	for _, s := range sm.shards {
		s.RLock()
		total += len(s.m)
		s.RUnlock()
	}
	return total
}

// Range 以并发安全的方式遍历所有键值对。
// 每个分片的数据在持有读锁期间被复制出来，回调函数在释放读锁后执行，
// 避免长耗时回调阻塞写操作或导致死锁。
func (sm *ShardedMap[K, V]) Range(fn func(key K, value V) bool) {
	for _, s := range sm.shards {
		s.RLock()
		keys := make([]K, 0, len(s.m))
		vals := make([]V, 0, len(s.m))
		for k, v := range s.m {
			keys = append(keys, k)
			vals = append(vals, v)
		}
		s.RUnlock()
		for i, k := range keys {
			if !fn(k, vals[i]) {
				return
			}
		}
	}
}
