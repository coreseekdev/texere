package rope

import (
	"runtime"
	"strings"
	"testing"
)

// ========== Optimization Comparison Tests ==========

// BenchmarkString_Old vs New
func BenchmarkString_Old(b *testing.B) {
	r := New(strings.Repeat("Hello, World! ", 1000))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.StringOld()
	}
}

func BenchmarkString_New(b *testing.B) {
	r := New(strings.Repeat("Hello, World! ", 1000))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.String()
	}
}

func BenchmarkString_Bytes(b *testing.B) {
	r := New(strings.Repeat("Hello, World! ", 1000))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.StringBytes()
	}
}

// ========== Append Comparison ==========

func BenchmarkAppend_Old(b *testing.B) {
	r := New(strings.Repeat("Hello, World! ", 100))
	text := " Appended"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Append(text)
	}
}

func BenchmarkAppend_New(b *testing.B) {
	r := New(strings.Repeat("Hello, World! ", 100))
	text := " Appended"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.AppendOptimized(text)
	}
}

// ========== Insert Comparison ==========

func BenchmarkInsert_Old(b *testing.B) {
	r := New(strings.Repeat("Hello, World! ", 100))
	text := "INSERTED"
	pos := r.Length() / 2
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Insert(pos, text)
	}
}

func BenchmarkInsert_New(b *testing.B) {
	r := New(strings.Repeat("Hello, World! ", 100))
	text := "INSERTED"
	pos := r.Length() / 2
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.InsertOptimized(pos, text)
	}
}

// ========== Delete Comparison ==========

func BenchmarkDelete_Old(b *testing.B) {
	r := New(strings.Repeat("Hello, World! ", 100))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Delete(10, 20)
	}
}

func BenchmarkDelete_New(b *testing.B) {
	r := New(strings.Repeat("Hello, World! ", 100))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.DeleteOptimized(10, 20)
	}
}

// ========== Comparison Test ==========

func TestOptimizationComparison(t *testing.T) {
	// Create test rope
	text := strings.Repeat("Hello, World! ", 100)
	r := New(text)

	// Test String()
	oldStr := r.StringOld()
	newStr := r.String()
	byteStr := r.StringBytes()

	if oldStr != newStr || newStr != byteStr {
		t.Errorf("String implementations differ!\nOld: %q\nNew: %q\nBytes: %q",
			oldStr[:100], newStr[:100], byteStr[:100])
	}

	// Test Append
	oldAppend := r.Append(" Appended")
	newAppend := r.AppendOptimized(" Appended")

	if oldAppend.String() != newAppend.String() {
		t.Error("Append implementations differ!")
	}

	// Test Insert
	oldInsert := r.Insert(500, "X")
	newInsert := r.InsertOptimized(500, "X")

	if oldInsert.String() != newInsert.String() {
		t.Error("Insert implementations differ!")
	}

	// Test Delete
	oldDelete := r.Delete(100, 200)
	newDelete := r.DeleteOptimized(100, 200)

	if oldDelete.String() != newDelete.String() {
		t.Error("Delete implementations differ!")
	}

	t.Log("All optimization implementations produce identical results ✓")
}

// ========== Memory Allocation Tests ==========

func TestMemory_String(t *testing.T) {
	r := New(strings.Repeat("Hello, World! ", 1000))

	// Test old implementation
	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	for i := 0; i < 100; i++ {
		_ = r.StringOld()
	}

	runtime.ReadMemStats(&m2)
	oldAlloc := m2.TotalAlloc - m1.TotalAlloc

	// Test new implementation
	runtime.GC()
	runtime.ReadMemStats(&m1)

	for i := 0; i < 100; i++ {
		_ = r.String()
	}

	runtime.ReadMemStats(&m2)
	newAlloc := m2.TotalAlloc - m1.TotalAlloc

	t.Logf("String() - Old: %d bytes, New: %d bytes, Improvement: %.1fx",
		oldAlloc, newAlloc, float64(oldAlloc)/float64(newAlloc))
}

func TestMemory_Append(t *testing.T) {
	r := New(strings.Repeat("Hello, World! ", 100))
	text := " Appended"

	// Test old implementation
	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	for i := 0; i < 100; i++ {
		r = r.Append(text)
	}

	runtime.ReadMemStats(&m2)
	oldAlloc := m2.TotalAlloc - m1.TotalAlloc

	// Test new implementation
	r2 := New(strings.Repeat("Hello, World! ", 100))
	runtime.GC()
	runtime.ReadMemStats(&m1)

	for i := 0; i < 100; i++ {
		r2 = r2.AppendOptimized(text)
	}

	runtime.ReadMemStats(&m2)
	newAlloc := m2.TotalAlloc - m1.TotalAlloc

	t.Logf("Append() - Old: %d bytes, New: %d bytes, Improvement: %.1fx",
		oldAlloc, newAlloc, float64(oldAlloc)/float64(newAlloc))
}

func TestMemory_Insert(t *testing.T) {
	text := strings.Repeat("Hello, World! ", 100)
	pos := len([]rune(text)) / 2
	insertText := "INSERTED"

	// Test old implementation
	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	r := New(text)
	for i := 0; i < 100; i++ {
		r = r.Insert(pos, insertText)
	}

	runtime.ReadMemStats(&m2)
	oldAlloc := m2.TotalAlloc - m1.TotalAlloc

	// Test new implementation
	runtime.GC()
	runtime.ReadMemStats(&m1)

	r2 := New(text)
	for i := 0; i < 100; i++ {
		r2 = r2.InsertOptimized(pos, insertText)
	}

	runtime.ReadMemStats(&m2)
	newAlloc := m2.TotalAlloc - m1.TotalAlloc

	t.Logf("Insert() - Old: %d bytes, New: %d bytes, Improvement: %.1fx",
		oldAlloc, newAlloc, float64(oldAlloc)/float64(newAlloc))
}
