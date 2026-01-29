package rope

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestHash_Consistency verifies that ropes with same content
// but different chunk boundaries produce the same hash
// This is ported from ropey's hash.rs
func TestHash_Consistency_Small(t *testing.T) {
	// Build two ropes with the same contents but different chunk boundaries
	r1 := New("")
	b1 := NewBuilder()
	b1.Append("Hello w")
	b1.Append("orld")
	r1 = b1.Build()

	r2 := New("")
	b2 := NewBuilder()
	b2.Append("Hell")
	b2.Append("o world")
	r2 = b2.Build()

	// Should have same hash
	hash1 := r1.HashCode64()
	hash2 := r2.HashCode64()

	assert.Equal(t, hash1, hash2)
	assert.Equal(t, r1.String(), r2.String())
}

// TestHash_Consistency_Medium tests hash consistency with larger text
func TestHash_Consistency_Medium(t *testing.T) {
	text := "Hello World! This is a test string for hashing. " +
		"It should produce the same hash regardless of chunk boundaries. " +
		"The quick brown fox jumps over the lazy dog. " +
		"こんにちは世界 🌍"

	// Build rope with 5-byte chunks
	r1 := New("")
	b1 := NewBuilder()
	for i := 0; i < len(text); i += 5 {
		end := i + 5
		if end > len(text) {
			end = len(text)
		}
		b1.Append(text[i:end])
	}
	r1 = b1.Build()

	// Build rope with 7-byte chunks
	r2 := New("")
	b2 := NewBuilder()
	for i := 0; i < len(text); i += 7 {
		end := i + 7
		if end > len(text) {
			end = len(text)
		}
		b2.Append(text[i:end])
	}
	r2 = b2.Build()

	// Should have same hash
	hash1 := r1.HashCode64()
	hash2 := r2.HashCode64()

	assert.Equal(t, hash1, hash2)
	assert.Equal(t, r1.String(), r2.String())
}

// TestHash_Consistency_Large tests hash consistency with large text
func TestHash_Consistency_Large(t *testing.T) {
	text := ""
	for i := 0; i < 100; i++ {
		text += "Hello World! " +
			"The quick brown fox jumps over the lazy dog. " +
			"こんにちは世界 🌍🌎🌏\n"
	}

	// Build rope with 521-byte chunks
	r1 := New("")
	b1 := NewBuilder()
	for i := 0; i < len(text); i += 521 {
		end := i + 521
		if end > len(text) {
			end = len(text)
		}
		b1.Append(text[i:end])
	}
	r1 = b1.Build()

	// Build rope with 547-byte chunks
	r2 := New("")
	b2 := NewBuilder()
	for i := 0; i < len(text); i += 547 {
		end := i + 547
		if end > len(text) {
			end = len(text)
		}
		b2.Append(text[i:end])
	}
	r2 = b2.Build()

	// Should have same hash
	hash1 := r1.HashCode64()
	hash2 := r2.HashCode64()

	assert.Equal(t, hash1, hash2)
	assert.Equal(t, r1.String(), r2.String())
}

// TestHash_DifferentContent produces different hashes
func TestHash_DifferentContent(t *testing.T) {
	r1 := New("Hello World")
	r2 := New("Hello World!")

	hash1 := r1.HashCode64()
	hash2 := r2.HashCode64()

	assert.NotEqual(t, hash1, hash2)
}

// TestHash_EmptyRope produces consistent hash for empty rope
func TestHash_EmptyRope(t *testing.T) {
	r1 := Empty()
	r2 := New("")

	// Empty ropes should have same hash
	hash1 := r1.HashCode64()
	hash2 := r2.HashCode64()

	assert.Equal(t, hash1, hash2)
}

// TestHash_HashCode32 produces consistent 32-bit hash
func TestHash_HashCode32(t *testing.T) {
	r1 := New("Hello World")
	r2 := New("Hello World")

	hash1 := r1.HashCode32()
	hash2 := r2.HashCode32()

	assert.Equal(t, hash1, hash2)
}

// TestHash_HashEquals verifies HashEquals method
func TestHash_HashEquals(t *testing.T) {
	r1 := New("Hello World")
	r2 := New("Hello World")
	r3 := New("Hello World!")

	assert.True(t, r1.HashEquals(r2))
	assert.False(t, r1.HashEquals(r3))
}

// TestHash_SingleInsert verifies hash changes after insert
func TestHash_SingleInsert(t *testing.T) {
	r1 := New("Hello World")
	hash1 := r1.HashCode64()

	r2 := r1.Insert(5, "XXX")
	hash2 := r2.HashCode64()

	assert.NotEqual(t, hash1, hash2)
}

// TestHash_Delete verifies hash changes after delete
func TestHash_Delete(t *testing.T) {
	r1 := New("Hello World")
	hash1 := r1.HashCode64()

	r2 := r1.Delete(5, 6)
	hash2 := r2.HashCode64()

	assert.NotEqual(t, hash1, hash2)
}

// TestHash_SplitMerge verifies hash consistency after split/merge
func TestHash_SplitMerge(t *testing.T) {
	text := "Hello World Test String"
	r := New(text)
	hash1 := r.HashCode64()

	left, right := r.Split(6)
	merged := left.AppendRope(right)
	hash2 := merged.HashCode64()

	assert.Equal(t, hash1, hash2)
	assert.Equal(t, text, merged.String())
}

// TestHash_ChunkHashes returns hashes of all chunks
func TestHash_ChunkHashes(t *testing.T) {
	r1 := New("Hello")
	r2 := r1.Append(" World")

	hashes := r2.ChunkHashes()

	// Should have at least 2 chunks
	assert.True(t, len(hashes) >= 2)

	// Hashes should be non-zero
	for _, h := range hashes {
		assert.NotEqual(t, uint32(0), h)
	}
}

// TestHash_CombinedChunkHash returns combined hash
func TestHash_CombinedChunkHash(t *testing.T) {
	r := New("Hello World")
	r = r.Append(" Test")

	hash := r.CombinedChunkHash()

	assert.NotEqual(t, uint32(0), hash)
}

// TestHash_Unicode produces consistent hash for unicode
func TestHash_Unicode(t *testing.T) {
	text := "Hello 世界 🌍"

	r1 := New(text)
	r2 := New(text)

	hash1 := r1.HashCode64()
	hash2 := r2.HashCode64()

	assert.Equal(t, hash1, hash2)
}

// TestHash_CRLF produces consistent hash with CRLF
func TestHash_CRLF(t *testing.T) {
	text := "Line 1\r\nLine 2\r\nLine 3"

	r1 := New(text)
	r2 := New(text)

	hash1 := r1.HashCode64()
	hash2 := r2.HashCode64()

	assert.Equal(t, hash1, hash2)
}

// TestHash_Integrity verifies hash doesn't change for same rope
func TestHash_Integrity(t *testing.T) {
	r := New("Hello World Test")

	hash1 := r.HashCode64()
	hash2 := r.HashCode64()
	hash3 := r.HashCode64()

	// Hash should be stable
	assert.Equal(t, hash1, hash2)
	assert.Equal(t, hash2, hash3)
}
