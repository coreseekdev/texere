package rope

import (
	"sync"
	"testing"
)

// 全局池
var iteratorPool = sync.Pool{
	New: func() interface{} {
		return &Iterator{
			stack: make([]frame, 0, 16),
		}
	},
}

// NewIteratorPooled 使用对象池
func (r *Rope) NewIteratorPooled() *Iterator {
	it := iteratorPool.Get().(*Iterator)
	it.rope = r
	it.position = 0
	it.runePos = -1
	it.stack = it.stack[:0]
	it.exhausted = (r == nil || r.Length() == 0)
	return it
}

// ReleaseIterator 归还迭代器到池
func ReleaseIterator(it *Iterator) {
	iteratorPool.Put(it)
}

// Benchmark: 对比池化 vs 非池化
func BenchmarkIterator_NoPool(b *testing.B) {
	r := New("Hello World Test String")
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			it := r.NewIterator()
			for it.Next() {
				_ = it.Current()
			}
		}
	})
}

func BenchmarkIterator_WithPool(b *testing.B) {
	r := New("Hello World Test String")
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			it := r.NewIteratorPooled()
			for it.Next() {
				_ = it.Current()
			}
			ReleaseIterator(it)
		}
	})
}
