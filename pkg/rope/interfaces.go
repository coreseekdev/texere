package rope

// ========== Focused Interfaces for Rope (ISP Principle) ==========
//
// These interfaces break down the large Rope API into focused, composable interfaces.
// Consumers can depend only on the capabilities they need.
//
// Note: Methods that can fail now return errors instead of panicking.

// ReadOnlyDocument provides read-only access to document content.
type ReadOnlyDocument interface {
	// Length queries
	Length() int
	LengthBytes() int
	LengthChars() int

	// Content access
	String() string
	Bytes() []byte
	Slice(start, end int) (string, error)
}

// CharAtAccessor provides character-by-character access.
type CharAtAccessor interface {
	CharAt(pos int) (rune, error)
}

// ByteAtAccessor provides byte-by-byte access.
type ByteAtAccessor interface {
	ByteAt(pos int) (byte, error)
}

// MutableDocument provides document modification operations.
type MutableDocument interface {
	Insert(pos int, text string) (*Rope, error)
	Delete(start, end int) (*Rope, error)
	Replace(start, end int, text string) (*Rope, error)
}

// SplittableDocument provides split operations.
type SplittableDocument interface {
	Split(pos int) (*Rope, *Rope, error)
}

// Concatenable provides concatenation operations.
type Concatenable interface {
	Concat(other *Rope) *Rope
}

// Cloneable provides cloning operations.
type Cloneable interface {
	Clone() *Rope
}

// Searchable provides search operations.
type Searchable interface {
	Contains(substring string) bool
	Index(substring string) int
	LastIndex(substring string) int
}

// Validatable provides validation operations.
type Validatable interface {
	Validate() error
}

// Balanceable provides balance operations.
type Balanceable interface {
	Balance() *Rope
	Optimize() *Rope
	IsBalanced() bool
}

// DocumentMetrics provides metrics about document structure.
type DocumentMetrics interface {
	Size() int
	Depth() int
	Stats() *TreeStats
}

// ========== Composite Interfaces ==========

// FullDocument combines all document capabilities.
type FullDocument interface {
	ReadOnlyDocument
	CharAtAccessor
	ByteAtAccessor
	MutableDocument
	SplittableDocument
	Concatenable
	Cloneable
	Searchable
	Validatable
	Balanceable
	DocumentMetrics
}

// ReadOnly provides read-only capabilities including content access and search.
type ReadOnly interface {
	ReadOnlyDocument
	CharAtAccessor
	ByteAtAccessor
	Searchable
}

// ReadWrite provides both read and write capabilities.
type ReadWrite interface {
	ReadOnlyDocument
	MutableDocument
	Cloneable
}

// Editable provides mutation and splitting capabilities.
type Editable interface {
	MutableDocument
	SplittableDocument
}

// ========== Type Assertions ==========

// Ensure Rope implements all focused interfaces
var (
	_ ReadOnlyDocument   = (*Rope)(nil)
	_ CharAtAccessor     = (*Rope)(nil)
	_ ByteAtAccessor     = (*Rope)(nil)
	_ MutableDocument    = (*Rope)(nil)
	_ SplittableDocument = (*Rope)(nil)
	_ Concatenable       = (*Rope)(nil)
	_ Cloneable          = (*Rope)(nil)
	_ Searchable         = (*Rope)(nil)
	_ Validatable        = (*Rope)(nil)
	_ Balanceable        = (*Rope)(nil)
	_ DocumentMetrics    = (*Rope)(nil)
	_ FullDocument       = (*Rope)(nil)
)
