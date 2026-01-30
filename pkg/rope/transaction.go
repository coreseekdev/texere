package rope

import "time"

// Operation represents a single edit operation.
type Operation struct {
	OpType OpType
	Length int       // For Retain and Delete
	Text   string    // For Insert
}

// OpType represents the type of operation.
type OpType int

const (
	OpRetain OpType = iota // Keep n characters
	OpDelete               // Delete n characters
	OpInsert               // Insert text
)

// ChangeSet represents a set of changes to transform one document state to another.
// It is composable and invertible, making it ideal for undo/redo.
type ChangeSet struct {
	operations []Operation
	lenBefore  int // Document length before changes
	lenAfter   int // Document length after changes
}

// NewChangeSet creates a new empty ChangeSet.
func NewChangeSet(lenBefore int) *ChangeSet {
	return &ChangeSet{
		operations: make([]Operation, 0, 8),
		lenBefore:  lenBefore,
		lenAfter:   lenBefore,
	}
}

// Retain keeps n characters unchanged.
func (cs *ChangeSet) Retain(n int) *ChangeSet {
	cs.operations = append(cs.operations, Operation{OpType: OpRetain, Length: n})
	return cs
}

// Delete deletes n characters.
func (cs *ChangeSet) Delete(n int) *ChangeSet {
	cs.operations = append(cs.operations, Operation{OpType: OpDelete, Length: n})
	cs.lenAfter -= n
	return cs
}

// Insert inserts text.
func (cs *ChangeSet) Insert(text string) *ChangeSet {
	cs.operations = append(cs.operations, Operation{OpType: OpInsert, Text: text})
	cs.lenAfter += len([]rune(text))
	return cs
}

// LenBefore returns the document length before applying this changeset.
func (cs *ChangeSet) LenBefore() int {
	return cs.lenBefore
}

// LenAfter returns the document length after applying this changeset.
func (cs *ChangeSet) LenAfter() int {
	return cs.lenAfter
}

// IsEmpty returns true if the changeset has no operations.
func (cs *ChangeSet) IsEmpty() bool {
	return len(cs.operations) == 0
}

// fuse merges consecutive operations of the same type for optimization.
// This reduces the number of operations and improves performance.
// For example: Insert("a") + Insert("b") → Insert("ab")
func (cs *ChangeSet) fuse() {
	if len(cs.operations) <= 1 {
		return
	}

	fused := make([]Operation, 0, len(cs.operations))

	for _, op := range cs.operations {
		if len(fused) > 0 && fused[len(fused)-1].OpType == op.OpType {
			// Merge with previous operation of the same type
			prev := &fused[len(fused)-1]
			switch op.OpType {
			case OpRetain:
				prev.Length += op.Length
			case OpDelete:
				prev.Length += op.Length
			case OpInsert:
				prev.Text += op.Text
			}
		} else {
			fused = append(fused, op)
		}
	}

	cs.operations = fused
}

// Apply applies the changeset to a rope and returns the modified rope.
func (cs *ChangeSet) Apply(r *Rope) *Rope {
	if r == nil || cs.IsEmpty() {
		return r
	}

	// Fuse operations for optimization (reduces number of rope mutations)
	cs.fuse()

	result := r
	pos := 0

	for _, op := range cs.operations {
		switch op.OpType {
		case OpRetain:
			pos += op.Length

		case OpDelete:
			result = result.Delete(pos, pos+op.Length)

		case OpInsert:
			result = result.Insert(pos, op.Text)
			pos += len([]rune(op.Text))
		}
	}

	return result
}

// Invert creates an inverted changeset that undoes this changeset.
// The original rope state is needed to properly invert deletions.
func (cs *ChangeSet) Invert(original *Rope) *ChangeSet {
	if original == nil {
		return NewChangeSet(cs.lenAfter)
	}

	inverted := NewChangeSet(cs.lenAfter)
	pos := 0

	for _, op := range cs.operations {
		switch op.OpType {
		case OpRetain:
			inverted.Retain(op.Length)
			pos += op.Length

		case OpDelete:
			// Re-insert the deleted text
			deletedText := original.Slice(pos, pos+op.Length)
			inverted.Insert(deletedText)
			pos += op.Length

		case OpInsert:
			// Delete the inserted text
			inverted.Delete(len([]rune(op.Text)))
		}
	}

	// Fuse operations in the inverted changeset for optimization
	inverted.fuse()

	return inverted
}

// Compose composes this changeset with another, producing a changeset that
// represents applying this changeset followed by the other.
func (cs *ChangeSet) Compose(other *ChangeSet) *ChangeSet {
	if cs.IsEmpty() {
		return other
	}
	if other == nil || other.IsEmpty() {
		return cs
	}

	// This is a simplified composition
	// A full implementation would handle overlapping operations correctly
	result := NewChangeSet(cs.lenBefore)

	// Apply this changeset's operations
	pos := 0
	for _, op := range cs.operations {
		switch op.OpType {
		case OpRetain:
			result.Retain(op.Length)
			pos += op.Length

		case OpDelete:
			result.Delete(op.Length)
			pos += op.Length

		case OpInsert:
			result.Insert(op.Text)
			pos += len([]rune(op.Text))
		}
	}

	// Apply other changeset's operations (simplified)
	// TODO: Implement proper composition with position mapping
	for _, op := range other.operations {
		switch op.OpType {
		case OpRetain:
			result.Retain(op.Length)
		case OpDelete:
			result.Delete(op.Length)
		case OpInsert:
			result.Insert(op.Text)
		}
	}

	return result
}

// Transaction represents an atomic edit operation with optional selection state.
type Transaction struct {
	changeset  *ChangeSet
	timestamp  time.Time
}

// NewTransaction creates a new transaction from a changeset.
func NewTransaction(changeset *ChangeSet) *Transaction {
	return &Transaction{
		changeset: changeset,
		timestamp: time.Now(),
	}
}

// Changeset returns the transaction's changeset.
func (t *Transaction) Changeset() *ChangeSet {
	return t.changeset
}

// Timestamp returns when the transaction was created.
func (t *Transaction) Timestamp() time.Time {
	return t.timestamp
}

// Apply applies the transaction to a rope.
func (t *Transaction) Apply(r *Rope) *Rope {
	if t == nil || t.changeset == nil {
		return r
	}
	return t.changeset.Apply(r)
}

// Invert creates an inverted transaction for undo.
func (t *Transaction) Invert(original *Rope) *Transaction {
	if t == nil || t.changeset == nil {
		return NewTransaction(NewChangeSet(0))
	}
	return NewTransaction(t.changeset.Invert(original))
}

// IsEmpty returns true if the transaction has no changes.
func (t *Transaction) IsEmpty() bool {
	return t == nil || t.changeset == nil || t.changeset.IsEmpty()
}

// Selection represents a cursor or selection in the document.
// This is a placeholder for future selection support.
type Selection struct {
	Anchor int // Anchor position
	Cursor int // Cursor position (may equal anchor for simple cursor)
}

// NewSelection creates a new selection at the given position.
func NewSelection(pos int) *Selection {
	return &Selection{
		Anchor: pos,
		Cursor: pos,
	}
}

// NewSelectionRange creates a new selection from anchor to cursor.
func NewSelectionRange(anchor, cursor int) *Selection {
	return &Selection{
		Anchor: anchor,
		Cursor: cursor,
	}
}
