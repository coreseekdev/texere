package rope

// ========== Iterator Interfaces ==========

// This file defines unified interfaces for all iterator types in the rope package.
//
// Design Principles:
// 1. Go 1.23+ compatible - Can be used with for-range loops via iter.Seq patterns
// 2. Type-safe - Each iterator type has its own interface
// 3. Minimal - Core methods only, extension methods on concrete types
// 4. Consistent - All iterators follow the same naming patterns

// ========== Core Iterator Interface ==========

// Seq is the minimal interface that all iterators must implement.
// It follows the standard Go iterator pattern:
// - Next() advances the iterator and returns true if there are more elements
// - After Next() returns true, Current() returns the current element
//
// Note: Named "Seq" instead of "Iterator" to avoid conflict with concrete Iterator type.
//
// Example usage:
//
//	it := rope.New("Hello").NewIterator()
//	for it.Next() {
//	    fmt.Println(it.Current())
//	}
type Seq[T any] interface {
	// Next advances to the next element and returns true if successful.
	// Returns false when the iterator is exhausted.
	Next() bool

	// Current returns the current element.
	// Only valid after Next() has returned true.
	// Behavior is undefined if called before Next() or after exhaustion.
	Current() T
}

// ========== Positional Seq ==========

// PositionalSeq extends Seq with position tracking.
// This allows querying the current position within the iteration.
type PositionalSeq[T any] interface {
	Seq[T]

	// Position returns the current position in the iteration.
	// The meaning of "position" varies by iterator type:
	// - RuneIterator: character position
	// - BytesIterator: byte position
	// - LinesIterator: line number (0-indexed)
	// - GraphemeIterator: grapheme position
	Position() int
}

// ========== Resettable Seq ==========

// ResettableSeq extends Seq with the ability to reset.
// This allows reusing the iterator from the beginning.
type ResettableSeq[T any] interface {
	Seq[T]

	// Reset resets the iterator to its initial state.
	// After reset, Next() can be called to start iteration from the beginning.
	Reset()
}

// ========== State Query Seq ==========

// StatefulSeq extends Seq with state query methods.
// This allows checking iterator state without advancing.
type StatefulSeq[T any] interface {
	Seq[T]

	// HasNext returns true if there are more elements to iterate.
	// Does not advance the iterator.
	HasNext() bool

	// IsExhausted returns true if the iterator is exhausted.
	// After exhaustion, Next() will always return false.
	IsExhausted() bool
}

// ========== Seekable Seq ==========

// SeekableSeq extends Seq with random access capabilities.
// This allows jumping to specific positions in the iteration.
type SeekableSeq[T any] interface {
	PositionalSeq[T]

	// Seek moves the iterator to the specified position.
	// Returns true if successful, false if position is out of bounds.
	// After successful seek, Next() will advance from the new position.
	Seek(pos int) bool
}

// ========== Peekable Seq ==========

// PeekableSeq extends Seq with lookahead capability.
// This allows examining the next element without advancing.
type PeekableSeq[T any] interface {
	Seq[T]

	// Peek returns the next element without advancing the iterator.
	// Returns the element and true if successful, zero value and false if exhausted.
	Peek() (T, bool)
}

// ========== Collecting Seq ==========

// CollectingSeq extends Seq with bulk collection capability.
// This allows collecting all elements into a slice.
type CollectingSeq[T any] interface {
	Seq[T]

	// Collect collects all remaining elements into a slice.
	// The iterator is exhausted after this call.
	Collect() []T
}

// ========== Full Seq Interface ==========

// FullSeq combines all common iterator capabilities.
// This is the most feature-rich interface that most iterators should implement.
type FullSeq[T any] interface {
	PositionalSeq[T]
	ResettableSeq[T]
	StatefulSeq[T]
	PeekableSeq[T]
	CollectingSeq[T]
}

// ========== Type-Specific Iterator Interfaces ==========

// The following interfaces define the specific capabilities of each iterator type.
// These are implemented by the concrete iterator structs.

// RuneIteratorBehavior defines the rune iteration interface.
// Use this when you need character-by-character iteration.
type RuneIteratorBehavior interface {
	FullSeq[rune]

	// HasPrevious returns true if there is a previous element.
	HasPrevious() bool

	// Previous moves to the previous element and returns true if successful.
	Previous() bool

	// Skip advances the iterator by n positions.
	// Returns the number of positions actually skipped (may be less if exhausted).
	Skip(n int) int

	// Seek moves to a specific character position.
	// Returns true if successful.
	Seek(pos int) bool
}

// ReverseIteratorBehavior defines the reverse iteration interface.
// Use this for backward iteration through characters.
type ReverseIteratorBehavior interface {
	FullSeq[rune]

	// PositionFromStart returns the position from the start (not from end).
	PositionFromStart() int

	// SeekFromStart seeks to a position from the start.
	// Returns true if successful.
	SeekFromStart(pos int) bool

	// Skip advances backward by n positions.
	// Returns true if successful.
	Skip(n int) bool
}

// BytesIteratorBehavior defines the byte iteration interface.
// Use this for byte-by-byte iteration (e.g., binary data processing).
type BytesIteratorBehavior interface {
	FullSeq[byte]

	// BytePosition returns the current byte position.
	BytePosition() int

	// Skip advances by n bytes.
	// Returns true if successful.
	Skip(n int) bool

	// Seek moves to a specific byte position.
	// Returns true if successful.
	Seek(byteIdx int) bool

	// HasPeek returns true if Peek() would succeed.
	HasPeek() bool
}

// LinesIteratorBehavior defines the line iteration interface.
// Use this for line-by-line text processing.
type LinesIteratorBehavior interface {
	Seq[string]

	// LineNumber returns the current line number (0-indexed).
	LineNumber() int

	// Reset resets to the first line.
	Reset()

	// CurrentWithEnding returns the current line including line ending.
	CurrentWithEnding() (string, error)

	// ToSlice collects all lines into a slice.
	ToSlice() ([]string, error)
}

// GraphemeIteratorBehavior defines the grapheme cluster iteration interface.
// Use this for proper Unicode text handling (user-perceived characters).
type GraphemeIteratorBehavior interface {
	Seq[Grapheme]

	// Position returns the character position of current grapheme.
	Position() int

	// Reset resets to the first grapheme.
	Reset()
}

// ========== Adapter Functions for Go 1.23+ iter.Seq ==========

// These functions allow using rope iterators with Go 1.23+ for-range loops.
//
// Example usage with Go 1.23+:
//
//	r := rope.New("Hello")
//	for ch := range rope.IterRunes(r) {
//	    fmt.Println(ch)
//	}

// IterRunes returns an iter.Seq for rune iteration.
// Compatible with Go 1.23+ for-range loops.
//
// Example:
//
//	for ch := range rope.IterRunes(r) {
//	    fmt.Printf("%c\n", ch)
//	}
func IterRunes(r *Rope) func(yield func(rune) bool) {
	return func(yield func(rune) bool) {
		it := r.NewIterator()
		for it.Next() {
			if !yield(it.Current()) {
				return
			}
		}
	}
}

// IterBytes returns an iter.Seq for byte iteration.
// Compatible with Go 1.23+ for-range loops.
func IterBytes(r *Rope) func(yield func(byte) bool) {
	return func(yield func(byte) bool) {
		it := r.NewBytesIterator()
		for it.Next() {
			if !yield(it.Current()) {
				return
			}
		}
	}
}

// IterGraphemes returns an iter.Seq for grapheme iteration.
// Compatible with Go 1.23+ for-range loops.
func IterGraphemes(r *Rope) func(yield func(Grapheme) bool) {
	return func(yield func(Grapheme) bool) {
		it := r.Graphemes()
		for it.Next() {
			if !yield(it.Current()) {
				return
			}
		}
	}
}

// IterLines returns an iter.Seq for line iteration.
// Compatible with Go 1.23+ for-range loops.
func IterLines(r *Rope) func(yield func(string) bool) {
	return func(yield func(string) bool) {
		it := r.LinesIterator()
		for it.Next() {
			line, err := it.Current()
			if err != nil {
				return
			}
			if !yield(line) {
				return
			}
		}
	}
}

// IterReverse returns an iter.Seq for reverse rune iteration.
// Compatible with Go 1.23+ for-range loops.
func IterReverse(r *Rope) func(yield func(rune) bool) {
	return func(yield func(rune) bool) {
		it := r.IterReverse()
		for it.Next() {
			ch, err := it.Current()
			if err != nil {
				return
			}
			if !yield(ch) {
				return
			}
		}
	}
}
