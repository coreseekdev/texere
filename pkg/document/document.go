// Package document provides the Document interface abstraction.
// This interface is used by the OT (Operational Transformation) layer
// to work with different document implementations (String, Rope, etc.).
package document

// Document represents an immutable text document.
// All operations that modify the document return a new Document instance.
type Document interface {
	// Length returns the number of characters (Unicode code points) in the document.
	Length() int

	// Slice returns a substring from start to end (exclusive).
	// The indices are character positions (not byte positions).
	// Panics if indices are out of bounds.
	Slice(start, end int) string

	// String returns the complete document content as a string.
	String() string

	// Bytes returns the complete document content as a byte slice.
	Bytes() []byte

	// Clone creates a copy of the document.
	// For immutable implementations, this may return the same instance.
	Clone() Document
}
