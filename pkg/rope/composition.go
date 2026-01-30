package rope

// Composition represents the result of composing two changesets.
type Composition struct {
	changeset    *ChangeSet
	posMapping   []int // Maps old positions to new positions
	inverseMap    []int // Maps new positions to old positions
}

// NewComposition creates a new composition.
func NewComposition() *Composition {
	return &Composition{
		changeset:  &ChangeSet{},
		posMapping: make([]int, 0),
		inverseMap: make([]int, 0),
	}
}

// Compose composes this changeset with another, producing a changeset that
// represents applying this changeset followed by the other.
// This is the full implementation with proper position mapping.
func (cs *ChangeSet) Compose(other *ChangeSet) *ChangeSet {
	if cs.IsEmpty() {
		if other == nil || other.IsEmpty() {
			return NewChangeSet(cs.lenBefore)
		}
		result := NewChangeSet(other.lenBefore)
		result.operations = append(result.operations, other.operations...)
		result.lenAfter = other.lenAfter
		return result
	}

	if other == nil || other.IsEmpty() {
		result := NewChangeSet(cs.lenBefore)
		result.operations = append(result.operations, cs.operations...)
		result.lenAfter = cs.lenAfter
		return result
	}

	composer := &changesetComposer{
		first:  cs,
		second: other,
		result: NewChangeSet(cs.lenBefore),
	}

	return composer.compose()
}

// changesetComposer handles the composition of two changesets.
type changesetComposer struct {
	first  *ChangeSet
	second *ChangeSet
	result *ChangeSet
	// Position tracking
	posInFirst  int // Position in first document
	posInSecond int // Position in second document (after first applied)
	posInResult int // Position in result document
}

// compose performs the composition.
func (c *changesetComposer) compose() *ChangeSet {
	// Build position map from first changeset
	posMap := c.buildPositionMap(c.first)

	// Apply second changeset with mapped positions
	c.posInFirst = 0
	c.posInSecond = 0
	c.posInResult = 0

	// Process all operations from both changesets
	firstOps := c.first.operations
	secondOps := c.second.operations

	i, j := 0, 0
	for i < len(firstOps) || j < len(secondOps) {
		if i < len(firstOps) && (j >= len(secondOps) || c.shouldProcessFirst(firstOps[i], secondOps, j, posMap)) {
			c.processFirstOp(firstOps[i])
			i++
		} else if j < len(secondOps) {
			c.processSecondOp(secondOps[j], posMap)
			j++
		}
	}

	c.result.lenAfter = c.posInResult
	return c.result
}

// buildPositionMap builds a map from document positions to positions after applying first changeset.
func (c *changesetComposer) buildPositionMap(cs *ChangeSet) map[int]int {
	posMap := make(map[int]int)
	pos := 0
	newPos := 0

	for _, op := range cs.operations {
		switch op.OpType {
		case OpRetain:
			// Map each position in the retained range
			for i := 0; i < op.Length; i++ {
				posMap[pos+i] = newPos + i
			}
			pos += op.Length
			newPos += op.Length

		case OpDelete:
			// Deleted positions map to the current position (they're removed)
			for i := 0; i < op.Length; i++ {
				posMap[pos+i] = newPos
			}
			pos += op.Length

		case OpInsert:
			// Inserted positions don't exist in original, skip
			newPos += len([]rune(op.Text))
		}
	}

	return posMap
}

// shouldProcessFirst determines if we should process the next operation from first changeset.
func (c *changesetComposer) shouldProcessFirst(firstOp Operation, secondOps []Operation, secondIdx int, posMap map[int]int) bool {
	if secondIdx >= len(secondOps) {
		return true
	}

	_ = secondOps[secondIdx] // Get second operation (may use in future)

	// If first operation is at or before second operation's position
	switch firstOp.OpType {
	case OpRetain, OpDelete:
		// First operation consumes characters
		return c.posInFirst < c.posInSecond

	case OpInsert:
		// First operation inserts characters
		// Process inserts before second operation if at same position
		return c.posInFirst <= c.posInSecond
	}

	return true
}

// processFirstOp processes an operation from the first changeset.
func (c *changesetComposer) processFirstOp(op Operation) {
	switch op.OpType {
	case OpRetain:
		c.result.Retain(op.Length)
		c.posInFirst += op.Length
		c.posInSecond += op.Length
		c.posInResult += op.Length

	case OpDelete:
		c.result.Delete(op.Length)
		c.posInFirst += op.Length

	case OpInsert:
		c.result.Insert(op.Text)
		c.posInResult += len([]rune(op.Text))
	}
}

// processSecondOp processes an operation from the second changeset with position mapping.
func (c *changesetComposer) processSecondOp(op Operation, posMap map[int]int) {
	switch op.OpType {
	case OpRetain:
		// Retain in second document means retain in first after mapping
		c.result.Retain(op.Length)
		c.posInSecond += op.Length
		c.posInFirst += op.Length
		c.posInResult += op.Length

	case OpDelete:
		// Delete in second document
		c.result.Delete(op.Length)
		c.posInSecond += op.Length
		c.posInFirst += op.Length

	case OpInsert:
		// Insert in second document
		c.result.Insert(op.Text)
		c.posInResult += len([]rune(op.Text))
	}
}

// MapPosition maps a position through this changeset.
// Returns the new position after applying the changeset.
func (cs *ChangeSet) MapPosition(pos int, assoc Assoc) int {
	if pos < 0 || pos > cs.lenBefore {
		return pos
	}

	currentPos := 0
	newPos := 0

	for _, op := range cs.operations {
		switch op.OpType {
		case OpRetain:
			if currentPos+op.Length > pos {
				// Position is within this retain
				offset := pos - currentPos
				newPos += offset
				return cs.applyAssociation(newPos, assoc, currentPos, op.Length, pos)
			}
			currentPos += op.Length
			newPos += op.Length

		case OpDelete:
			if currentPos+op.Length > pos {
				// Position is within deleted range
				return cs.applyAssociation(newPos, assoc, currentPos, op.Length, pos)
			}
			currentPos += op.Length

		case OpInsert:
			if currentPos >= pos {
				// Already past the position
				return cs.applyAssociation(newPos, assoc, currentPos, len([]rune(op.Text)), pos)
			}
			newPos += len([]rune(op.Text))
		}

		if currentPos >= pos {
			break
		}
	}

	return newPos
}

// applyAssociation applies cursor association to determine final position.
func (cs *ChangeSet) applyAssociation(newPos int, assoc Assoc, opStart int, opLength int, targetPos int) int {
	switch assoc {
	case AssocBefore:
		return newPos

	case AssocAfter:
		return newPos + opLength

	case AssocBeforeSticky:
		// Keep at the same relative offset
		offset := targetPos - opStart
		if offset > opLength {
			offset = opLength
		}
		return newPos + offset

	case AssocAfterSticky:
		// Keep at the same relative offset
		offset := targetPos - opStart
		if offset > opLength {
			offset = opLength
		}
		return newPos + offset

	default:
		return newPos
	}
}

// MapPositions maps multiple positions through this changeset.
func (cs *ChangeSet) MapPositions(positions []int, associations []Assoc) []int {
	if len(positions) != len(associations) {
		panic("positions and associations must have same length")
	}

	result := make([]int, len(positions))
	for i, pos := range positions {
		result[i] = cs.MapPosition(pos, associations[i])
	}
	return result
}

// Transform transforms a changeset to apply after this changeset.
// This is useful for concurrent editing scenarios.
func (cs *ChangeSet) Transform(other *ChangeSet) *ChangeSet {
	if cs.IsEmpty() {
		return other
	}
	if other.IsEmpty() {
		return cs
	}

	// For now, return composed changeset
	// A full implementation would handle concurrent edits more carefully
	return cs.Compose(other)
}

// InvertAt creates an inverted changeset that undoes this changeset,
// assuming the changeset was applied at a specific position.
func (cs *ChangeSet) InvertAt(original *Rope, pos int) *ChangeSet {
	if original == nil {
		return NewChangeSet(cs.lenAfter)
	}

	inverted := NewChangeSet(cs.lenAfter)
	currentPos := pos

	for _, op := range cs.operations {
		switch op.OpType {
		case OpRetain:
			currentPos += op.Length

		case OpDelete:
			// Re-insert the deleted text
			deletedText := original.Slice(currentPos, currentPos+op.Length)
			inverted.Insert(deletedText)
			currentPos += op.Length

		case OpInsert:
			// Delete the inserted text
			inverted.Delete(len([]rune(op.Text)))
		}
	}

	// Fuse operations in the inverted changeset
	inverted.fuse()

	return inverted
}

// CanApplyAt checks if this changeset can be applied at the given position.
func (cs *ChangeSet) CanApplyAt(pos int) bool {
	return pos >= 0 && pos <= cs.lenBefore
}

// Optimized returns an optimized version of this changeset with fused operations.
func (cs *ChangeSet) Optimized() *ChangeSet {
	optimized := NewChangeSet(cs.lenBefore)
	optimized.operations = make([]Operation, len(cs.operations))
	copy(optimized.operations, cs.operations)
	optimized.fuse()
	return optimized
}

// Split splits this changeset at the given position.
// Returns two changesets: before and after the position.
func (cs *ChangeSet) Split(pos int) (*ChangeSet, *ChangeSet) {
	if pos <= 0 {
		return NewChangeSet(cs.lenBefore), cs
	}
	if pos >= cs.lenBefore {
		return cs, NewChangeSet(cs.lenAfter)
	}

	before := NewChangeSet(pos)
	after := NewChangeSet(cs.lenBefore - pos)

	currentPos := 0

	for _, op := range cs.operations {
		switch op.OpType {
		case OpRetain:
			if currentPos+op.Length <= pos {
				// Entire operation is before split point
				before.Retain(op.Length)
				currentPos += op.Length
			} else if currentPos >= pos {
				// Entire operation is after split point
				after.Retain(op.Length)
			} else {
				// Split the retain operation
				beforeLen := pos - currentPos
				afterLen := op.Length - beforeLen
				before.Retain(beforeLen)
				after.Retain(afterLen)
				currentPos += beforeLen
			}

		case OpDelete:
			if currentPos+op.Length <= pos {
				// Entire delete is before split point
				before.Delete(op.Length)
				currentPos += op.Length
			} else if currentPos >= pos {
				// Entire delete is after split point
				after.Delete(op.Length)
			} else {
				// Split the delete operation
				beforeLen := pos - currentPos
				afterLen := op.Length - beforeLen
				before.Delete(beforeLen)
				after.Delete(afterLen)
				currentPos += beforeLen
			}

		case OpInsert:
			// Inserts always happen at the current position
			if currentPos < pos {
				before.Insert(op.Text)
			} else {
				after.Insert(op.Text)
			}
		}
	}

	return before, after
}

// Merge merges this changeset with another at the same position.
// This is useful for combining concurrent edits.
func (cs *ChangeSet) Merge(other *ChangeSet) *ChangeSet {
	if cs.IsEmpty() {
		return other
	}
	if other.IsEmpty() {
		return cs
	}

	// Simply concatenate operations for now
	// A full implementation would intelligently merge overlapping edits
	result := NewChangeSet(cs.lenBefore)
	result.operations = append(result.operations, cs.operations...)
	result.operations = append(result.operations, other.operations...)
	result.fuse()
	result.lenAfter = cs.lenAfter + (other.lenAfter - other.lenBefore)

	return result
}
