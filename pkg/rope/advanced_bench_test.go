package rope

import (
	"fmt"
	"strings"
	"testing"
)

// ========== Zero-Allocation Benchmarks ==========

// BenchmarkInsert_ZeroAlloc compares zero-allocation insert vs standard.
func BenchmarkInsert_ZeroAlloc(b *testing.B) {
	text := strings.Repeat("Hello, World! ", 100)
	r := New(text)
	pos := r.Length() / 2
	insertText := "INSERTED"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.InsertZeroAlloc(pos, insertText)
	}
}

// BenchmarkInsert_Standard for comparison.
func BenchmarkInsert_Standard(b *testing.B) {
	text := strings.Repeat("Hello, World! ", 100)
	r := New(text)
	pos := r.Length() / 2
	insertText := "INSERTED"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Insert(pos, insertText)
	}
}

// BenchmarkInsert_Optimized for comparison.
func BenchmarkInsert_Optimized(b *testing.B) {
	text := strings.Repeat("Hello, World! ", 100)
	r := New(text)
	pos := r.Length() / 2
	insertText := "INSERTED"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.InsertOptimized(pos, insertText)
	}
}

// ========== Delete Benchmarks ==========

func BenchmarkDelete_ZeroAlloc(b *testing.B) {
	text := strings.Repeat("Hello, World! ", 100)
	r := New(text)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.DeleteZeroAlloc(10, 20)
	}
}

func BenchmarkDelete_Standard(b *testing.B) {
	text := strings.Repeat("Hello, World! ", 100)
	r := New(text)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Delete(10, 20)
	}
}

func BenchmarkDelete_Optimized(b *testing.B) {
	text := strings.Repeat("Hello, World! ", 100)
	r := New(text)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.DeleteOptimized(10, 20)
	}
}

// ========== Append/Prepend Benchmarks ==========

func BenchmarkAppend_ZeroAlloc(b *testing.B) {
	text := strings.Repeat("Hello, World! ", 100)
	r := New(text)
	appendText := " Appended"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.AppendZeroAlloc(appendText)
	}
}

func BenchmarkPrepend_ZeroAlloc(b *testing.B) {
	text := strings.Repeat("Hello, World! ", 100)
	r := New(text)
	prependText := "Prepended "

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.PrependZeroAlloc(prependText)
	}
}

// ========== Mixed Operation Benchmarks ==========

func BenchmarkMixedOps_ZeroAlloc(b *testing.B) {
	text := strings.Repeat("Hello, World! ", 100)
	r := New(text)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r = r.AppendZeroAlloc(" X")
		r = r.InsertZeroAlloc(r.Length()/2, " Y")
		r = r.DeleteZeroAlloc(0, 1)
	}
}

func BenchmarkMixedOps_Standard(b *testing.B) {
	text := strings.Repeat("Hello, World! ", 100)
	r := New(text)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r = r.Append(" X")
		r = r.Insert(r.Length()/2, " Y")
		r = r.Delete(0, 1)
	}
}

// ========== Sequential Operations ==========

func BenchmarkSequentialInserts_ZeroAlloc(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := New("")
		for j := 0; j < 100; j++ {
			r = r.AppendZeroAlloc(fmt.Sprintf("Item %d ", j))
		}
	}
}

func BenchmarkSequentialInserts_Standard(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := New("")
		for j := 0; j < 100; j++ {
			r = r.Append(fmt.Sprintf("Item %d ", j))
		}
	}
}

// ========== Large Text Benchmarks ==========

func BenchmarkInsert_Large_ZeroAlloc(b *testing.B) {
	text := strings.Repeat("Hello, World! ", 1000) // ~26KB
	r := New(text)
	pos := r.Length() / 2
	insertText := strings.Repeat("X", 1000) // 1KB insert

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.InsertZeroAlloc(pos, insertText)
	}
}

func BenchmarkInsert_Large_Standard(b *testing.B) {
	text := strings.Repeat("Hello, World! ", 1000) // ~26KB
	r := New(text)
	pos := r.Length() / 2
	insertText := strings.Repeat("X", 1000) // 1KB insert

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Insert(pos, insertText)
	}
}

// ========== Byte Position Cache Benchmarks ==========

func BenchmarkByteCache_SingleLookup(b *testing.B) {
	text := strings.Repeat("Hello, World! ", 100)
	cache := NewBytePosCache(text)
	pos := 50

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cache.GetBytePos(pos)
	}
}

func BenchmarkByteCache_MultipleLookups(b *testing.B) {
	text := strings.Repeat("Hello, World! ", 100)
	cache := NewBytePosCache(text)
	positions := []int{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, pos := range positions {
			_ = cache.GetBytePos(pos)
		}
	}
}

// ========== Copy-on-Write Benchmarks ==========

func BenchmarkCowRope_Insert(b *testing.B) {
	text := strings.Repeat("Hello, World! ", 100)
	r := NewCowRope(text)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Insert(r.Length()/2, "X")
	}
}

func BenchmarkCowRope_Delete(b *testing.B) {
	text := strings.Repeat("Hello, World! ", 100)
	r := NewCowRope(text)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Delete(10, 20)
	}
}

func BenchmarkCowRope_ShareAndMutate(b *testing.B) {
	text := strings.Repeat("Hello, World! ", 100)
	r := NewCowRope(text)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r1 := r.Insert(10, "X")
		r2 := r.Insert(20, "Y")
		r3 := r.Insert(30, "Z")
		_ = r1
		_ = r2
		_ = r3
	}
}

// ========== Rebalancing Benchmarks ==========

func BenchmarkRebalance_Balanced(b *testing.B) {
	// Create a balanced rope
	ropes := make([]*Rope, 100)
	for i := 0; i < 100; i++ {
		ropes[i] = New(fmt.Sprintf("Chunk %d ", i))
	}
	r := Concat(ropes...)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Balance()
	}
}

func BenchmarkRebalance_Unbalanced(b *testing.B) {
	// Create an unbalanced rope (left-skewed)
	r := New("")
	for i := 0; i < 100; i++ {
		r = r.Append(fmt.Sprintf("Chunk %d ", i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Balance()
	}
}

// ========== Memory Allocation Benchmarks ==========

func BenchmarkAllocations_InsertZeroAlloc(b *testing.B) {
	text := strings.Repeat("Hello, World! ", 100)
	r := New(text)
	b.ReportAllocs()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r = r.InsertZeroAlloc(r.Length()/2, "X")
	}
}

func BenchmarkAllocations_InsertStandard(b *testing.B) {
	text := strings.Repeat("Hello, World! ", 100)
	r := New(text)
	b.ReportAllocs()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r = r.Insert(r.Length()/2, "X")
	}
}

func BenchmarkAllocations_DeleteZeroAlloc(b *testing.B) {
	text := strings.Repeat("Hello, World! ", 100)
	b.ReportAllocs()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := New(text)
		r = r.DeleteZeroAlloc(10, 20)
	}
}

func BenchmarkAllocations_DeleteStandard(b *testing.B) {
	text := strings.Repeat("Hello, World! ", 100)
	b.ReportAllocs()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := New(text)
		r = r.Delete(10, 20)
	}
}

// ========== Comparison Tests ==========

func TestCompareImplementations(t *testing.T) {
	text := strings.Repeat("Hello, World! ", 50)

	// Test Insert
	r1 := New(text).Insert(50, "INSERTED")
	r2 := New(text).InsertOptimized(50, "INSERTED")
	r3 := New(text).InsertZeroAlloc(50, "INSERTED")

	if r1.String() != r2.String() || r2.String() != r3.String() {
		t.Error("Insert implementations differ")
	}

	// Test Delete
	r1 = New(text).Delete(10, 20)
	r2 = New(text).DeleteOptimized(10, 20)
	r3 = New(text).DeleteZeroAlloc(10, 20)

	if r1.String() != r2.String() || r2.String() != r3.String() {
		t.Error("Delete implementations differ")
	}

	// Test Append
	r1 = New(text).Append("APPENDED")
	r2 = New(text).AppendOptimized("APPENDED")
	r3 = New(text).AppendZeroAlloc("APPENDED")

	if r1.String() != r2.String() || r2.String() != r3.String() {
		t.Error("Append implementations differ")
	}

	t.Log("All implementations produce identical results ✓")
}

// ========== Stress Tests ==========

func TestStress_ManyInserts(t *testing.T) {
	r := New("")
	for i := 0; i < 1000; i++ {
		r = r.InsertZeroAlloc(i, fmt.Sprintf("%d", i%10))
	}
	expectedLen := 1000
	if r.Length() != expectedLen {
		t.Errorf("Expected length %d, got %d", expectedLen, r.Length())
	}
}

func TestStress_ManyDeletes(t *testing.T) {
	// Build rope
	r := New("")
	for i := 0; i < 1000; i++ {
		r = r.Append(fmt.Sprintf("%d", i%10))
	}

	// Delete from beginning
	for i := 0; i < 500; i++ {
		r = r.DeleteZeroAlloc(0, 1)
	}

	if r.Length() != 500 {
		t.Errorf("Expected length 500, got %d", r.Length())
	}
}
