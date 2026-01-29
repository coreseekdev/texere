package rope

import (
	"strings"
)

// ========== Optimized String Operations ==========

// StringOptimized returns the complete content of the rope as a string.
// Optimized version using strings.Builder to avoid intermediate allocations.
func (r *Rope) StringOptimized() string {
	if r == nil || r.length == 0 {
		return ""
	}

	// Optimization 1: Use strings.Builder with pre-allocated capacity
	var b strings.Builder
	b.Grow(r.size)

	// Optimization 2: Iterate chunks directly (more efficient than char-by-char)
	it := r.Chunks()
	for it.Next() {
		b.WriteString(it.Current())
	}

	return b.String()
}

// StringOld is the old implementation kept for comparison/testing.
func (r *Rope) StringOld() string {
	if r == nil || r.length == 0 {
		return ""
	}
	return r.root.Slice(0, r.length)
}

// ========== Optimized Slice Operations ==========

// SliceOptimized returns a substring using optimized algorithm.
// This avoids the recursive string concatenation in the original implementation.
func (r *Rope) SliceOptimized(start, end int) string {
	if r == nil {
		return ""
	}
	if start < 0 || end > r.length || start > end {
		panic("slice bounds out of range")
	}
	if start == end {
		return ""
	}

	// Use strings.Builder for efficient concatenation
	var b strings.Builder

	// Estimate size (rough estimate)
	estimatedSize := r.size * (end - start) / r.length
	b.Grow(estimatedSize)

	it := r.Chunks()
	charsToSkip := start
	charsToCollect := end - start

	for it.Next() && charsToCollect > 0 {
		chunk := it.Current()
		chunkRunes := []rune(chunk)

		if charsToSkip > 0 {
			if len(chunkRunes) <= charsToSkip {
				// Skip entire chunk
				charsToSkip -= len(chunkRunes)
				continue
			} else {
				// Skip part of chunk
				chunk = string(chunkRunes[charsToSkip:])
				charsToSkip = 0
			}
		}

		// Collect from chunk
		if len(chunkRunes) <= charsToCollect {
			b.WriteString(chunk)
			charsToCollect -= len(chunkRunes)
		} else {
			// Take only what we need from this chunk
			b.WriteString(string(chunkRunes[:charsToCollect]))
			break
		}
	}

	return b.String()
}

// ========== Byte-Optimized Operations ==========

// StringBytes returns the string with minimal allocations.
// This is even more optimized than String() for large ropes.
func (r *Rope) StringBytes() string {
	if r == nil || r.length == 0 {
		return ""
	}

	// Pre-allocate exact size
	result := make([]byte, 0, r.size)

	it := r.Chunks()
	for it.Next() {
		result = append(result, it.Current()...)
	}

	return string(result)
}
